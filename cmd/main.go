package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "tcddbot/config"
    "tcddbot/db"
    "tcddbot/handlers"
    "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Initialize database
    database, err := db.Initialize(cfg.DBPath)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer database.Close()

    // Initialize bot
    bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
    if err != nil {
        log.Fatalf("Failed to initialize bot: %v", err)
    }

    // Initialize handler
    handler := handlers.NewHandler(bot, database, cfg)

    // Setup signal handling
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // Start background tasks
    go handler.StartPeriodicCheck(ctx)
    go handler.StartCleanup(ctx)

    // Handle updates
    updates := bot.GetUpdatesChan(tgbotapi.UpdateConfig{
        Timeout: 60,
    })

    for {
        select {
        case update := <-updates:
            handler.HandleUpdate(ctx, update)
        case <-sigChan:
            log.Println("Shutting down gracefully...")
            cancel()
            return
        case <-ctx.Done():
            return
        }
    }
}