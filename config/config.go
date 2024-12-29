package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
    BotToken          string
    DBPath            string
    APIEndpoint       string
    AuthToken         string
    UnitID           string
    CheckInterval     time.Duration
    CleanupInterval   time.Duration
}

func Load() (*Config, error) {
    if err := godotenv.Load("../.env"); err != nil {
        fmt.Println("No .env file found")
    }

    return &Config{
        BotToken:        os.Getenv("BOT_TOKEN"),
        DBPath:         os.Getenv("DB_PATH"),
        APIEndpoint:    "https://web-api-prod-ytp.tcddtasimacilik.gov.tr/tms/train",
        AuthToken:      os.Getenv("AUTHORIZATION_TOKEN"),
        UnitID:        "3895",
        CheckInterval:  5 * time.Second,
        CleanupInterval: 1 * time.Hour,
    }, nil
}
