package db

import (
    "database/sql"
    _ "modernc.org/sqlite"
)

func Initialize(dbPath string) (*sql.DB, error) {
    db, err := sql.Open("sqlite", dbPath)
    if (err != nil) {
        return nil, err
    }

    // Enable WAL mode
    if _, err := db.Exec(`PRAGMA journal_mode=WAL`); err != nil {
        return nil, err
    }

    // Create tables if they don't exist
    if err := createTables(db); err != nil {
        return nil, err
    }

    return db, nil
}

func createTables(db *sql.DB) error {
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS subscriptions (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            chat_id INTEGER,
            departure_station_id INTEGER,
            arrival_station_id INTEGER,
            travel_date TEXT,
            last_notified DATETIME,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            deleted_at DATETIME
        )`)
    return err
}
