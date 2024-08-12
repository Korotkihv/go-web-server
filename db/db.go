package db

import (
	"database/sql"
	"log"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	globalDB *sql.DB
	once     sync.Once
)

// InitGlobalDB инициализирует глобальное подключение к базе данных.
func InitGlobalDB(dbPath string) {
	once.Do(func() {
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		err = db.Ping()
		if err != nil {
			log.Fatalf("Failed to ping database: %v", err)
		}

		if err := RunMigrations(db); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}

		globalDB = db
		log.Println("Database connection established")
	})
}

func GetDB() *sql.DB {
	return globalDB
}
