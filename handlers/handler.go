package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"tcddbot/config"
	"tcddbot/model"
	"tcddbot/service"
	"time"

	"tcddbot/util"
	"tcddbot/worker"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const NOTIFICATION_INTERVAL = 1 * time.Hour

type Station struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Handler struct {
	bot         *tgbotapi.BotAPI
	db          *sql.DB
	cfg         *config.Config
	trainSvc    *service.TrainService
	stations    []Station
	stationsMux sync.RWMutex
	workerPool  *worker.Pool
}

func NewHandler(bot *tgbotapi.BotAPI, db *sql.DB, cfg *config.Config) *Handler {
	h := &Handler{
		bot:      bot,
		db:       db,
		cfg:      cfg,
		trainSvc: service.NewTrainService(cfg),
	}

	if err := h.loadStations(); err != nil {
		log.Printf("Error loading stations: %v", err)
	}

	// Initialize worker pool with 5 workers and 100 queue size
	h.workerPool = worker.NewPool(5, 100, h.processSubscription)

	return h
}

func (h *Handler) loadStations() error {
	file, err := os.Open("./stations.json")
	if err != nil {
		return fmt.Errorf("error opening stations.json: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &h.stations)
	if err != nil {
		return err
	}

	log.Printf("Loaded %d stations, first station: %s", len(h.stations), h.stations[0].Name)
	return nil
}

func (h *Handler) HandleUpdate(ctx context.Context, update tgbotapi.Update) {
    if update.CallbackQuery != nil {
        h.handleCallback(ctx, update.CallbackQuery)
        return
    }

    if update.Message == nil || !update.Message.IsCommand() {
        return
    }

    switch update.Message.Command() {
    case CommandStart, CommandHelp:
        h.handleHelp(update)
    case CommandSearchStation:
        h.handleStationSearch(update)
    case CommandSubscribe:
        h.handleSubscription(ctx, update)
    case CommandListSubscriptions:
        h.handleListSubscriptions(ctx, update)
    }
}

func (h *Handler) handleHelp(update tgbotapi.Update) {
    msg := tgbotapi.NewMessage(update.Message.Chat.ID, CommandDescriptions[update.Message.Command()])
    msg.ParseMode = "Markdown"
    h.bot.Send(msg)
}

func (h *Handler) handleStationSearch(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	keyword := strings.TrimSpace(update.Message.CommandArguments())
	if keyword == "" {
		msg := tgbotapi.NewMessage(chatID, MsgInvalidStationSearch)
		h.bot.Send(msg)
		return
	}

	var matchingStations []string
	h.stationsMux.RLock()
	for _, station := range h.stations {
		if strings.Contains(strings.ToLower(station.Name), strings.ToLower(keyword)) {
			matchingStations = append(matchingStations, station.Name)
		}
	}
	h.stationsMux.RUnlock()

	if len(matchingStations) > 0 {
		msg := tgbotapi.NewMessage(chatID, strings.Join(matchingStations, "\n"))
		h.bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(chatID, "Girilen anahtar kelime ile eşleşen istasyon bulunamadı.")
		h.bot.Send(msg)
	}
}

func (h *Handler) handleSubscription(ctx context.Context, update tgbotapi.Update) {
    chatID := update.Message.Chat.ID
    parts := strings.SplitN(strings.TrimSpace(update.Message.CommandArguments()), "-", 3)
    if len(parts) != 3 {
        msg := tgbotapi.NewMessage(chatID, MsgInvalidSubscription)
        h.bot.Send(msg)
        return
    }

    departureStationName := strings.TrimSpace(parts[0])
    arrivalStationName := strings.TrimSpace(parts[1])
    travelDate := strings.TrimSpace(parts[2])

    var departureStationID, arrivalStationID int
    h.stationsMux.RLock()
    for _, station := range h.stations {
        if station.Name == departureStationName {
            departureStationID = station.ID
        }
        if station.Name == arrivalStationName {
            arrivalStationID = station.ID
        }
    }
    h.stationsMux.RUnlock()

    if departureStationID == 0 || arrivalStationID == 0 {
        msg := tgbotapi.NewMessage(chatID, "Lütfen geçerli kalkış ve varış istasyonlarını girin.")
        h.bot.Send(msg)
        return
    }

    // İlk önce mevcut abonelik kontrolü
    var count int
    err := h.db.QueryRow(`SELECT COUNT(*) FROM subscriptions WHERE chat_id = ? AND departure_station_id = ? AND arrival_station_id = ? AND travel_date = ? AND deleted_at IS NULL`, 
        chatID, departureStationID, arrivalStationID, travelDate).Scan(&count)
    if err != nil {
        log.Printf("Error checking existing subscription: %v", err)
        msg := tgbotapi.NewMessage(chatID, "Bir hata oluştu. Lütfen daha sonra tekrar deneyin.")
        h.bot.Send(msg)
        return
    }
    if count > 0 {
        msg := tgbotapi.NewMessage(chatID, "Bu kalkış ve varış istasyonları için zaten bir aboneliğiniz var.")
        h.bot.Send(msg)
        return
    }

    // Tren ve koltuk kontrolü
    response, err := h.trainSvc.CheckAvailability(ctx, departureStationID, arrivalStationID, travelDate)
    if err != nil {
        // Check if it's a "no trains available" error
        if strings.Contains(err.Error(), "no trains available") {
            msg := tgbotapi.NewMessage(chatID, "Bu tarih için henüz sefer bulunmamaktadır. Lütfen daha sonra tekrar deneyiniz.")
            h.bot.Send(msg)
            return
        }
        
        log.Printf("Error checking train availability: %v", err)
        msg := tgbotapi.NewMessage(chatID, "Tren uygunluğu kontrol edilirken hata oluştu. Lütfen daha sonra tekrar deneyin.")
        h.bot.Send(msg)
        return
    }

    availableSeats := util.FindAvailableSeats(response.TrainLegs)
    if len(availableSeats) > 0 {
        yhtFound := false
        for _, seat := range availableSeats {
            if seat.IsYHT {
                yhtFound = true
                loc, _ := time.LoadLocation("Europe/Istanbul")
                
                var seatDetails []string
                for className, count := range seat.AvailableSeats {
                    seatDetails = append(seatDetails, fmt.Sprintf("%s: %d", className, count))
                }
                
                msgText := fmt.Sprintf(
                    "Hemen müsait YHT bulundu!\nTren: %s\nKalkış: %s\nMüsait Koltuklar:\n%s",
                    seat.Train.Name,
                    seat.DepartureTime.In(loc).Format("02.01.2006 15:04"),
                    strings.Join(seatDetails, "\n"),
                )
                msg := tgbotapi.NewMessage(chatID, msgText)
                h.bot.Send(msg)
                return
            }
        }

        if !yhtFound {
            msg := tgbotapi.NewMessage(chatID, "YHT dışında müsait tren bulundu, aramaya devam edilecek.")
            h.bot.Send(msg)
            
            // Continue with subscription creation
            _, err = h.db.Exec(
                `INSERT INTO subscriptions (chat_id, departure_station_id, arrival_station_id, travel_date) VALUES (?, ?, ?, ?)`,
                chatID, departureStationID, arrivalStationID, travelDate)
            if err != nil {
                log.Printf("Error creating subscription: %v", err)
                msg := tgbotapi.NewMessage(chatID, "Abonelik oluşturulurken bir hata oluştu. Lütfen daha sonra tekrar deneyin.")
                h.bot.Send(msg)
                return
            }
        }
    } else {
        // No seats available, create subscription
        _, err = h.db.Exec(
            `INSERT INTO subscriptions (chat_id, departure_station_id, arrival_station_id, travel_date) VALUES (?, ?, ?, ?)`,
            chatID, departureStationID, arrivalStationID, travelDate)
        if err != nil {
            log.Printf("Error creating subscription: %v", err)
            msg := tgbotapi.NewMessage(chatID, "Abonelik oluşturulurken bir hata oluştu. Lütfen daha sonra tekrar deneyin.")
            h.bot.Send(msg)
            return
        }

        msg := tgbotapi.NewMessage(chatID, "Şu an için müsait koltuk yok. Aboneliğiniz oluşturuldu! Koltuk bulunduğunda size haber vereceğim.")
        h.bot.Send(msg)
    }
}

func (h *Handler) StartPeriodicCheck(ctx context.Context) {
    h.workerPool.Start(ctx)
    ticker := time.NewTicker(h.cfg.CheckInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            h.workerPool.Stop()
            return
        case <-ticker.C:
            h.queueSubscriptionChecks(ctx)
        }
    }
}

func (h *Handler) StartCleanup(ctx context.Context) {
    ticker := time.NewTicker(h.cfg.CleanupInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := h.cleanupOldSubscriptions(ctx); err != nil {
                log.Printf("Error cleaning up old subscriptions: %v", err)
            }
        }
    }
}

func (h *Handler) cleanupOldSubscriptions(ctx context.Context) error {
    _, err := h.db.ExecContext(ctx, `
        UPDATE subscriptions 
        SET deleted_at = CURRENT_TIMESTAMP 
        WHERE travel_date < DATE('now', '-1 day') 
        AND deleted_at IS NULL`)
    return err
}

func (h *Handler) queueSubscriptionChecks(ctx context.Context) {
    rows, err := h.db.QueryContext(ctx, `
        SELECT chat_id, departure_station_id, arrival_station_id, travel_date 
        FROM subscriptions 
        WHERE deleted_at IS NULL`)
    if err != nil {
        log.Printf("Error querying subscriptions: %v", err)
        return
    }
    defer rows.Close()

    for rows.Next() {
        var job worker.Job
        if err := rows.Scan(&job.ChatID, &job.DepartureStation, &job.ArrivalStation, &job.TravelDate); err != nil {
            log.Printf("Error scanning row: %v", err)
            continue
        }
        h.workerPool.AddJob(job)
    }
}

func (h *Handler) processSubscription(ctx context.Context, job worker.Job) error {
    // Check if we should notify based on last notification time
    var lastNotified sql.NullTime
    err := h.db.QueryRowContext(ctx, `
        SELECT last_notified 
        FROM subscriptions 
        WHERE chat_id = ? AND departure_station_id = ? AND arrival_station_id = ? AND travel_date = ? AND deleted_at IS NULL`,
        job.ChatID, job.DepartureStation, job.ArrivalStation, job.TravelDate).Scan(&lastNotified)
    
    if err != nil && err != sql.ErrNoRows {
        return fmt.Errorf("check last notification: %w", err)
    }

    // If last notification was less than an hour ago, skip
    if lastNotified.Valid && time.Since(lastNotified.Time) < NOTIFICATION_INTERVAL {
        return nil
    }

    response, err := h.trainSvc.CheckAvailability(ctx, job.DepartureStation, job.ArrivalStation, job.TravelDate)
    if err != nil {
        return fmt.Errorf("check availability: %w", err)
    }

    availableSeats := util.FindAvailableSeats(response.TrainLegs)
    if len(availableSeats) > 0 {
        for _, seat := range availableSeats {
            if err := h.notifyAvailability(job.ChatID, seat.Train, job.DepartureStation, job.ArrivalStation, 
                seat.DepartureTime.Format("2006-01-02T15:04:05")); err != nil {
                return fmt.Errorf("notify availability: %w", err)
            }
            
            // Update last_notified time
            if !seat.IsYHT {
                _, err = h.db.ExecContext(ctx, `
                    UPDATE subscriptions 
                    SET last_notified = CURRENT_TIMESTAMP 
                    WHERE chat_id = ? AND departure_station_id = ? AND arrival_station_id = ? AND travel_date = ?`,
                    job.ChatID, job.DepartureStation, job.ArrivalStation, job.TravelDate)
                if err != nil {
                    return fmt.Errorf("update last notification: %w", err)
                }
            } else {
                // If YHT is found, deactivate the subscription
                return h.deactivateSubscription(ctx, job.ChatID, job.DepartureStation, job.ArrivalStation, job.TravelDate)
            }
        }
    }

    return nil
}

func (h *Handler) notifyAvailability(chatID int64, trainInfo model.Trains, departureStationID, arrivalStationID int, departureTime string) error {
    departureTimeParsed, err := time.Parse("2006-01-02T15:04:05", departureTime)
    if err != nil {
        return fmt.Errorf("parse departure time: %w", err)
    }

    loc, err := time.LoadLocation("Europe/Istanbul")
    if err != nil {
        return fmt.Errorf("load timezone: %w", err)
    }

    departureTimeTurkish := departureTimeParsed.In(loc).Format("02.01.2006 15:04")

    var departureStationName, arrivalStationName string
    h.stationsMux.RLock()
    for _, station := range h.stations {
        if station.ID == departureStationID {
            departureStationName = station.Name
        }
        if station.ID == arrivalStationID {
            arrivalStationName = station.Name
        }
    }
    h.stationsMux.RUnlock()

    var seatDetails []string
    for _, cabinClass := range trainInfo.CabinClassAvailabilities {
        if cabinClass.CabinClass.Name != "TEKERLEKLİ SANDALYE" && cabinClass.AvailabilityCount > 0 {
            seatDetails = append(seatDetails, fmt.Sprintf("%s: %d", cabinClass.CabinClass.Name, cabinClass.AvailabilityCount))
        }
    }

    var msgPrefix string
    if trainInfo.Type == "YHT" {
        msgPrefix = "YHT bulundu!"
    } else {
        msgPrefix = "Kara tren bulundu"
    }

    msgText := fmt.Sprintf("%s\nTren: %s (%s)\nKalkış: %s\nKalkış İstasyonu: %s\nVarış İstasyonu: %s\nMüsait Koltuklar:\n%s",
        msgPrefix,
        trainInfo.Name,
        trainInfo.Type,
        departureTimeTurkish,
        departureStationName,
        arrivalStationName,
        strings.Join(seatDetails, "\n"))

    msg := tgbotapi.NewMessage(chatID, msgText)
    _, err = h.bot.Send(msg)
    return err
}

func (h *Handler) deactivateSubscription(ctx context.Context, chatID int64, departureStationID, arrivalStationID int, travelDate string) error {
    _, err := h.db.ExecContext(ctx, `
        UPDATE subscriptions 
        SET deleted_at = CURRENT_TIMESTAMP 
        WHERE chat_id = ? 
        AND departure_station_id = ? 
        AND arrival_station_id = ? 
        AND travel_date = ?`,
        chatID, departureStationID, arrivalStationID, travelDate)
    return err
}

func (h *Handler) handleListSubscriptions(ctx context.Context, update tgbotapi.Update) {
    chatID := update.Message.Chat.ID
    
    subscriptions, err := h.getActiveSubscriptions(ctx, chatID)
    if err != nil {
        log.Printf("Error getting subscriptions: %v", err)
        msg := tgbotapi.NewMessage(chatID, "Abonelikleriniz getirilirken bir hata oluştu.")
        h.bot.Send(msg)
        return
    }

    if len(subscriptions) == 0 {
        msg := tgbotapi.NewMessage(chatID, "Aktif aboneliğiniz bulunmamaktadır.")
        h.bot.Send(msg)
        return
    }

    // Create message with inline keyboard
    var keyboard [][]tgbotapi.InlineKeyboardButton
    var messageText strings.Builder
    messageText.WriteString("Aktif Abonelikleriniz:\n\n")

    for i, sub := range subscriptions {
        messageText.WriteString(fmt.Sprintf("%d. %s → %s (%s)\n",
            i+1, sub.DepartureStation, sub.ArrivalStation, sub.TravelDate))
        
        keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{
            tgbotapi.NewInlineKeyboardButtonData(
                fmt.Sprintf("🗑️ %s → %s aboneliğini iptal et", sub.DepartureStation, sub.ArrivalStation),
                fmt.Sprintf("%s%d", CancelSubscriptionPrefix, sub.ID),
            ),
        })
    }

    msg := tgbotapi.NewMessage(chatID, messageText.String())
    msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
    h.bot.Send(msg)
}

func (h *Handler) handleCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) {
    if strings.HasPrefix(callback.Data, CancelSubscriptionPrefix) {
        subscriptionID, err := strconv.ParseInt(strings.TrimPrefix(callback.Data, CancelSubscriptionPrefix), 10, 64)
        if err != nil {
            log.Printf("Error parsing subscription ID: %v", err)
            return
        }

        if err := h.cancelSubscription(ctx, callback.Message.Chat.ID, subscriptionID); err != nil {
            log.Printf("Error canceling subscription: %v", err)
            h.bot.Send(tgbotapi.NewCallback(callback.ID, "Abonelik iptal edilirken bir hata oluştu."))
            return
        }

        // Update the message to remove the button
        editMsg := tgbotapi.NewEditMessageText(
            callback.Message.Chat.ID,
            callback.Message.MessageID,
            callback.Message.Text + "\n\n✅ Seçilen abonelik başarıyla iptal edildi.",
        )
        h.bot.Send(editMsg)
        h.bot.Send(tgbotapi.NewCallback(callback.ID, "Abonelik başarıyla iptal edildi."))
    }
}

func (h *Handler) getActiveSubscriptions(ctx context.Context, chatID int64) ([]SubscriptionInfo, error) {
    rows, err := h.db.QueryContext(ctx, `
        SELECT 
            id,
            departure_station_id,
            arrival_station_id,
            travel_date
        FROM subscriptions 
        WHERE chat_id = ? AND deleted_at IS NULL
        ORDER BY created_at DESC`,
        chatID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var subscriptions []SubscriptionInfo
    for rows.Next() {
        var sub SubscriptionInfo
        var departureID, arrivalID int
        if err := rows.Scan(&sub.ID, &departureID, &arrivalID, &sub.TravelDate); err != nil {
            return nil, err
        }

        h.stationsMux.RLock()
        for _, station := range h.stations {
            if station.ID == departureID {
                sub.DepartureStation = station.Name
            }
            if station.ID == arrivalID {
                sub.ArrivalStation = station.Name
            }
        }
        h.stationsMux.RUnlock()

        subscriptions = append(subscriptions, sub)
    }

    return subscriptions, nil
}

func (h *Handler) cancelSubscription(ctx context.Context, chatID int64, subscriptionID int64) error {
    result, err := h.db.ExecContext(ctx, `
        UPDATE subscriptions 
        SET deleted_at = CURRENT_TIMESTAMP 
        WHERE id = ? AND chat_id = ? AND deleted_at IS NULL`,
        subscriptionID, chatID)
    if err != nil {
        return err
    }

    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rows == 0 {
        return fmt.Errorf("subscription not found or already cancelled")
    }

    return nil
}