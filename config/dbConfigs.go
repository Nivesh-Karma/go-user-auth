package config

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() {

	url := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: url}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Postgres database is not available.", err)
	}
	DB = db
}
