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
	ID       int    `json:"id"`
	Name     string `json:"name"`
	CityName string `json:"cityName"`
	PairIDs  []int  `json:"pairs"`
}

type Handler struct {
	bot         *tgbotapi.BotAPI
	db          *sql.DB
	cfg         *config.Config
	trainSvc    *service.TrainService
	stations    []Station
	stationsMux sync.RWMutex
	workerPool  *worker.Pool
	userStates  map[int64]*UserState
	statesMux   sync.RWMutex
}

func NewHandler(bot *tgbotapi.BotAPI, db *sql.DB, cfg *config.Config) *Handler {
	h := &Handler{
		bot:        bot,
		db:         db,
		cfg:        cfg,
		trainSvc:   service.NewTrainService(cfg),
		userStates: make(map[int64]*UserState),
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
	if (err != nil) {
		return fmt.Errorf("error opening stations.json: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if (err != nil) {
		return err
	}

	err = json.Unmarshal(data, &h.stations)
	if (err != nil) {
		return err
	}

	log.Printf("Loaded %d stations, first station: %s", len(h.stations), h.stations[0].Name)
	return nil
}

// Update HandleUpdate to handle non-command messages
func (h *Handler) HandleUpdate(ctx context.Context, update tgbotapi.Update) {
    if update.CallbackQuery != nil {
        h.handleCallback(ctx, update.CallbackQuery)
        return
    }

    if update.Message == nil {
        return
    }

    if update.Message.IsCommand() {
        switch update.Message.Command() {
        case CommandStart, CommandHelp:
            h.handleHelp(update)
        case CommandSearchStation:
            h.handleStationSearch(update)
        case CommandSubscribe:
            h.handleSubscriptionStart(update)
        case CommandListSubscriptions:
            h.handleListSubscriptions(ctx, update)
        }
        return
    }

    // Handle non-command messages (station search and date input)
    h.HandleMessage(update)
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
		if strings.Contains(strings.ToLower(station.Name), strings.ToLower(keyword)) ||
			strings.Contains(strings.ToLower(station.CityName), strings.ToLower(keyword)) {
			matchingStations = append(matchingStations, fmt.Sprintf("%s (%s)", station.Name, station.CityName))
		}
	}
	h.stationsMux.RUnlock()

	if len(matchingStations) > 0 {
		var responseText strings.Builder
        responseText.WriteString("🔍 *Bulunan İstasyonlar:*\n\n")
        for i, station := range matchingStations {
            responseText.WriteString(fmt.Sprintf("%d. %s\n", i+1, station))
        }
        responseText.WriteString("\n💡 Bu istasyon adlarını takip oluştururken kullanabilirsiniz.")
        
        msg := tgbotapi.NewMessage(chatID, responseText.String())
        msg.ParseMode = "Markdown"
        h.bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(chatID, "❌ *İstasyon Bulunamadı*\n\n"+
            "Lütfen farklı bir arama yapın.\n"+
            "💡 Kısmi kelimeler ile de arama yapabilirsiniz.\n"+
            "Örnek: 'ist' yazarak İstanbul'daki istasyonları bulabilirsiniz.")
        msg.ParseMode = "Markdown"
        h.bot.Send(msg)
	}
}

func (h *Handler) handleSubscriptionStart(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	h.statesMux.Lock()
	h.userStates[chatID] = &UserState{
		State:       StateSelectDeparture,
		CurrentPage: 0,
	}
	h.statesMux.Unlock()

	msg := tgbotapi.NewMessage(chatID, "🔍 *KALKIŞ İstasyonu Seçimi*\n\n"+
        "*İstasyon adını yazın:*\n"+
        "• Örnek: ankara, istanbul, izmir\n\n"+
        "💡 En az 2 karakter girmelisiniz")
    msg.ParseMode = "Markdown"
    h.bot.Send(msg)
}

// Update handleCallback to handle search dummy button
func (h *Handler) handleCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) {
    chatID := callback.Message.Chat.ID

    if strings.HasPrefix(callback.Data, "station_") {
        h.handleStationSelection(callback)
        return
    }

    switch callback.Data {
    case CallbackDateToday:
        h.handleDateSelection(callback, time.Now())
    case CallbackDateTomorrow:
        h.handleDateSelection(callback, time.Now().AddDate(0, 0, 1))
    case CallbackDateCustom:
        // Send message asking for custom date input
        msg := tgbotapi.NewMessage(chatID, "Lütfen tarihi GG-AA-YYYY formatında girin:")
        h.bot.Send(msg)
    }

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
            callback.Message.Text+"\n\n✅ Seçilen abonelik başarıyla iptal edildi.",
        )
        h.bot.Send(editMsg)
        h.bot.Send(tgbotapi.NewCallback(callback.ID, "Abonelik başarıyla iptal edildi."))
    }
}

func (h *Handler) handleStationSelection(callback *tgbotapi.CallbackQuery) {
    chatID := callback.Message.Chat.ID
    stationID := strings.TrimPrefix(callback.Data, "station_")

    h.statesMux.Lock()
    state := h.userStates[chatID]

    if state.State == StateSelectDeparture {
        // Store selected departure station
        state.DepartureStation = stationID
        state.State = StateSelectArrival
        state.CurrentPage = 0
        h.statesMux.Unlock()

        msg := tgbotapi.NewEditMessageText(chatID, callback.Message.MessageID,
            "🔍 *VARIŞ İstasyonu Seçimi*\n\n"+
                "*İstasyon adını yazın:*\n"+
                "• Örnek: ankara, istanbul, izmir\n\n"+
                "💡 En az 2 karakter girmelisiniz")
        msg.ParseMode = "Markdown"
        h.bot.Send(msg)
    } else if state.State == StateSelectArrival {
        // Check if departure and arrival stations are the same
        if stationID == state.DepartureStation {
            h.statesMux.Unlock()
            msg := tgbotapi.NewMessage(chatID, "❌ Kalkış ve varış istasyonları aynı olamaz. Lütfen farklı bir istasyon seçin.")
            h.bot.Send(msg)
            return
        }

        // Check if arrival station is in pair_ids of departure station
        depID, _ := strconv.Atoi(state.DepartureStation)
        arrID, _ := strconv.Atoi(stationID)
        
        h.stationsMux.RLock()
        var depStation Station
        var validPair bool
        for _, station := range h.stations {
            if station.ID == depID {
                depStation = station
                break
            }
        }
        h.stationsMux.RUnlock()

        // Check if arrival station is in departure station's pair_ids
        for _, pairID := range depStation.PairIDs {
            if pairID == arrID {
                validPair = true
                break
            }
        }

        if !validPair {
            h.statesMux.Unlock()
            msg := tgbotapi.NewMessage(chatID, "❌ Bu istasyonlar arasında sefer bulunmamaktadır. Lütfen farklı bir istasyon seçin.")
            h.bot.Send(msg)
            return
        }

        // Continue with valid station selection
        state.ArrivalStation = stationID
        state.State = StateSelectDate
        h.statesMux.Unlock()

        // Create date selection keyboard
        keyboard := [][]tgbotapi.InlineKeyboardButton{
            {
                tgbotapi.NewInlineKeyboardButtonData("Bugün", CallbackDateToday),
                tgbotapi.NewInlineKeyboardButtonData("Yarın", CallbackDateTomorrow),
            },
            {
                tgbotapi.NewInlineKeyboardButtonData("Özel Tarih", CallbackDateCustom),
            },
        }
        markup := tgbotapi.NewInlineKeyboardMarkup(keyboard...)
        msg := tgbotapi.NewEditMessageText(chatID, callback.Message.MessageID, "Lütfen tarih seçin:")
        msg.ReplyMarkup = &markup
        h.bot.Send(msg)
    }
}

func (h *Handler) createStationKeyboard(page int, filter string) *tgbotapi.InlineKeyboardMarkup {
    var filteredStations []Station
    h.stationsMux.RLock()
    
    // Get current state if any
    chatID := int64(0) // You'll need to pass this from the calling function
    h.statesMux.RLock()
    state := h.userStates[chatID]
    h.statesMux.RUnlock()

    for _, station := range h.stations {
        // If we're selecting arrival station, only show valid pairs
        if state != nil && state.State == StateSelectArrival {
            depID, _ := strconv.Atoi(state.DepartureStation)
            if station.ID == depID {
                continue // Skip departure station
            }
            // Only include stations that are valid pairs
            var isValidPair bool
            for _, s := range h.stations {
                if s.ID == depID {
                    for _, pairID := range s.PairIDs {
                        if pairID == station.ID {
                            isValidPair = true
                            break
                        }
                    }
                    break
                }
            }
            if !isValidPair {
                continue
            }
        }

        if strings.Contains(strings.ToLower(station.Name), strings.ToLower(filter)) ||
			strings.Contains(strings.ToLower(station.CityName), strings.ToLower(filter)) {
            filteredStations = append(filteredStations, station)
        }
    }
    h.stationsMux.RUnlock()

    // If no stations found
    if len(filteredStations) == 0 {
        keyboard := [][]tgbotapi.InlineKeyboardButton{
            {
                tgbotapi.NewInlineKeyboardButtonData("❌ İstasyon bulunamadı, yeniden deneyin", "new_search"),
            },
        }
        markup := tgbotapi.NewInlineKeyboardMarkup(keyboard...)
        return &markup
    }

    var keyboard [][]tgbotapi.InlineKeyboardButton
    for _, station := range filteredStations {
		displayName := fmt.Sprintf("%s (%s)", station.Name, station.CityName)
        keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{
            tgbotapi.NewInlineKeyboardButtonData(displayName, "station_"+strconv.Itoa(station.ID)),
        })
    }

    markup := tgbotapi.NewInlineKeyboardMarkup(keyboard...)
    return &markup
}

func (h *Handler) createSubscription(chatID int64, departureStationID, arrivalStationID, travelDate string) {
	// Convert station IDs from string to int
	depID, _ := strconv.Atoi(departureStationID)
	arrID, _ := strconv.Atoi(arrivalStationID)
	
	// Create subscription in database
	_, err := h.db.Exec(
		`INSERT INTO subscriptions (chat_id, departure_station_id, arrival_station_id, travel_date) 
		 VALUES (?, ?, ?, ?)`,
		chatID, depID, arrID, travelDate)
	
	if err != nil {
		log.Printf("Error creating subscription: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Abonelik oluşturulurken bir hata oluştu. Lütfen daha sonra tekrar deneyin.")
		h.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Aboneliğiniz başarıyla oluşturuldu! Uygun koltuk bulunduğunda size haber vereceğim.")
	h.bot.Send(msg)
}

// Update HandleMessage for better search handling
func (h *Handler) HandleMessage(update tgbotapi.Update) {
    chatID := update.Message.Chat.ID

    h.statesMux.RLock()
    state := h.userStates[chatID]
    h.statesMux.RUnlock()

    if state == nil {
        return
    }

    if state.State == StateSelectDeparture || state.State == StateSelectArrival {
        // Handle station search
        query := strings.TrimSpace(update.Message.Text)
        if len(query) < 2 {
            msg := tgbotapi.NewMessage(chatID, "❌ *Çok Kısa Arama*\n\n"+
                "Lütfen en az 2 karakter girin.\n"+
                "💡 Örnek: 'ank', 'ist', 'izm' gibi")
            msg.ParseMode = "Markdown"
            h.bot.Send(msg)
            return
        }

        var matchingStations []Station
        h.stationsMux.RLock()
        for _, station := range h.stations {
            if strings.Contains(strings.ToLower(station.Name), strings.ToLower(query)) ||
				strings.Contains(strings.ToLower(station.CityName), strings.ToLower(query)) {
                matchingStations = append(matchingStations, station)
            }
        }
        h.stationsMux.RUnlock()

        if len(matchingStations) == 0 {
            msg := tgbotapi.NewMessage(chatID, "Bu arama için istasyon bulunamadı. Lütfen farklı bir arama yapın.")
            h.bot.Send(msg)
            return
        }

        var keyboard [][]tgbotapi.InlineKeyboardButton
        for _, station := range matchingStations {
			displayName := fmt.Sprintf("%s (%s)", station.Name, station.CityName)
            keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{
                tgbotapi.NewInlineKeyboardButtonData(displayName, "station_"+strconv.Itoa(station.ID)),
            })
        }

        var msgText string
        if state.State == StateSelectDeparture {
            msgText = fmt.Sprintf("🔍 *'%s' için bulunan KALKIŞ istasyonları:*", query)
        } else {
            msgText = fmt.Sprintf("🔍 *'%s' için bulunan VARIŞ istasyonları:*", query)
        }

        msg := tgbotapi.NewMessage(chatID, msgText)
        markup := tgbotapi.NewInlineKeyboardMarkup(keyboard...)
        msg.ReplyMarkup = &markup
        msg.ParseMode = "Markdown"
        h.bot.Send(msg)
        return
    }

    // Handle custom date input
    if state.State == StateSelectDate {
        // Parse custom date
        date, err := time.Parse("02-01-2006", update.Message.Text)
        if err != nil {
            msg := tgbotapi.NewMessage(chatID, "Geçersiz tarih formatı. Lütfen GG-AA-YYYY formatında girin:")
            h.bot.Send(msg)
            return
        }

        h.handleDateSelection(&tgbotapi.CallbackQuery{
            Message: update.Message,
        }, date)
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
			// If YHT is found, notify and deactivate subscription
			if seat.IsYHT {
				if err := h.notifyAvailability(job.ChatID, seat.Train, job.DepartureStation, job.ArrivalStation,
					seat.DepartureTime.Format("2006-01-02T15:04:05")); err != nil {
					return fmt.Errorf("notify YHT availability: %w", err)
				}
				return h.deactivateSubscription(ctx, job.ChatID, job.DepartureStation, job.ArrivalStation, job.TravelDate)
			}
		}

		// For non-YHT trains, notify hourly and continue subscription
		if time.Since(lastNotified.Time) >= NOTIFICATION_INTERVAL {
			for _, seat := range availableSeats {
				if err := h.notifyAvailability(job.ChatID, seat.Train, job.DepartureStation, job.ArrivalStation,
					seat.DepartureTime.Format("2006-01-02T15:04:05")); err != nil {
					return fmt.Errorf("notify availability: %w", err)
				}
			}
			
			// Update last notification time
			_, err = h.db.ExecContext(ctx, `
				UPDATE subscriptions 
				SET last_notified = CURRENT_TIMESTAMP 
				WHERE chat_id = ? AND departure_station_id = ? AND arrival_station_id = ? AND travel_date = ?`,
				job.ChatID, job.DepartureStation, job.ArrivalStation, job.TravelDate)
			if err != nil {
				return fmt.Errorf("update last notification: %w", err)
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
			seatDetails = append(seatDetails, fmt.Sprintf("🎫 %s: %d koltuk", cabinClass.CabinClass.Name, cabinClass.AvailabilityCount))
		}
	}

	var msgPrefix string
	if trainInfo.Type == "YHT" {
		msgPrefix = "🚅 *YHT BİLETİ BULUNDU!*"
	} else {
		msgPrefix = "🚂 Konvansiyonel tren bulundu"
	}

	msgText := fmt.Sprintf("%s\n\n"+
		"🚉 *Güzergah:* %s → %s\n"+
		"🕒 *Kalkış Zamanı:* %s\n"+
		"🎫 *Tren:* %s (%s)\n\n"+
		"*Müsait Koltuklar:*\n%s",
		msgPrefix,
		departureStationName,
		arrivalStationName,
		departureTimeTurkish,
		trainInfo.Name,
		trainInfo.Type,
		strings.Join(seatDetails, "\n"))

	msg := tgbotapi.NewMessage(chatID, msgText)
	msg.ParseMode = "Markdown"
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

// Add the missing handleDateSelection method
func (h *Handler) handleDateSelection(callback *tgbotapi.CallbackQuery, selectedDate time.Time) {
    chatID := callback.Message.Chat.ID

    h.statesMux.Lock()
    state := h.userStates[chatID]
    h.statesMux.Unlock()

    if state == nil {
        return
    }

    // Format date for subscription
    dateStr := selectedDate.Format("02-01-2006")

    // Check current date
    if selectedDate.Before(time.Now().AddDate(0, 0, -1)) {
        msg := tgbotapi.NewMessage(chatID, "Geçmiş bir tarih seçemezsiniz. Lütfen gelecek bir tarih seçin.")
        h.bot.Send(msg)
        return
    }

    // Call CheckAvailability before creating subscription
    depID, _ := strconv.Atoi(state.DepartureStation)
    arrID, _ := strconv.Atoi(state.ArrivalStation)

    // First check if subscription already exists
    var count int
    err := h.db.QueryRow(`SELECT COUNT(*) FROM subscriptions WHERE chat_id = ? AND departure_station_id = ? AND arrival_station_id = ? AND travel_date = ? AND deleted_at IS NULL`,
        chatID, depID, arrID, dateStr).Scan(&count)
    if err != nil {
        log.Printf("Error checking existing subscription: %v", err)
        msg := tgbotapi.NewMessage(chatID, "Bir hata oluştu. Lütfen daha sonra tekrar deneyin.")
        h.bot.Send(msg)
        return
    }
    if count > 0 {
        msg := tgbotapi.NewMessage(chatID, "Bu güzergah için zaten bir takibiniz bulunmaktadır.")
        h.bot.Send(msg)
        return
    }

    response, err := h.trainSvc.CheckAvailability(context.Background(), depID, arrID, dateStr)
    if err != nil {
        if strings.Contains(err.Error(), "no trains available") {
            msg := tgbotapi.NewMessage(chatID, "Bu tarih için henüz sefer bulunmamaktadır. Lütfen daha sonra tekrar deneyiniz.")
            h.bot.Send(msg)
            return
        }
        log.Printf("Error checking availability: %v", err)
    }

    var yhtFound bool
    if response != nil {
        availableSeats := util.FindAvailableSeats(response.TrainLegs)
        if len(availableSeats) > 0 {
            for _, seat := range availableSeats {
                if seat.IsYHT {
                    yhtFound = true
                    h.notifyAvailability(chatID, seat.Train, depID, arrID,
                        seat.DepartureTime.Format("2006-01-02T15:04:05"))
                    msg := tgbotapi.NewMessage(chatID, "✨ YHT bulundu! Yukarıdaki seferi hemen kontrol ediniz.\n"+
                        "🎯 Takip oluşturulmadı çünkü bilet şu an müsait!")
                    h.bot.Send(msg)
                    break
                }
            }
            if !yhtFound {
                h.createSubscription(chatID, state.DepartureStation, state.ArrivalStation, dateStr)
                msg := tgbotapi.NewMessage(chatID, "🎫 Konvansiyonel tren bulundu\n"+
                    "✅ Takip oluşturuldu ve YHT için aramaya devam edilecek\n"+
                    "📱 Müsait YHT bulunduğunda anında bildirim alacaksınız!")
                h.bot.Send(msg)
            }
        } else {
            h.createSubscription(chatID, state.DepartureStation, state.ArrivalStation, dateStr)
            msg := tgbotapi.NewMessage(chatID, "🔍 Şu an için müsait koltuk bulunmuyor\n"+
                "✅ Takip başarıyla oluşturuldu\n"+
                "📱 Uygun koltuk bulunduğunda anında bildirim alacaksınız!")
            h.bot.Send(msg)
        }
    } else {
        // No response or error, create subscription
        h.createSubscription(chatID, state.DepartureStation, state.ArrivalStation, dateStr)
        msg := tgbotapi.NewMessage(chatID, "Aboneliğiniz oluşturuldu! Koltuk bulunduğunda size haber vereceğim.")
        h.bot.Send(msg)
    }

    // Clean up state
    h.statesMux.Lock()
    delete(h.userStates, chatID)
    h.statesMux.Unlock()
}