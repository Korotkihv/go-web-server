package db

import (
	"database/sql"
	"log"
)

func RunMigrations(db *sql.DB) error {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS headerinfo (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		headername TEXT NOT NULL,
		value TEXT NOT NULL,
		port INTEGER NOT NULL
	);
	`

	_, err := db.Exec(createTableQuery)
	if err != nil {
		panic(err)
	}

	log.Println("Migrations completed successfully!")
	return nil
}
