package config

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {

	url := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: url}))
	if err != nil {
		log.Fatal("Postgres database is not available.", err)
	}
	DB = db
}
