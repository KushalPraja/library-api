package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

var DB *sql.DB

func Connect() {
	var err error
	DB, err = sql.Open("sqlite3", "./../example.db")
	if err != nil {
		log.Fatal("Failed to open database", err)
	}

	if err := DB.Ping(); err != nil {
		log.Fatal("Failed to ping database", err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS library (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		Book_name TEXT NOT NULL,
		Author TEXT NOT NULL,
		ISBN INTEGER NOT NULL
	);`

	if _, err := DB.Exec(createTable); err != nil {
		log.Fatal("Failed to create table", err)
	}
}
