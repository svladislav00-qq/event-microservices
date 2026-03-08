package postgres

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func GetDBForAccounts() {
	var err error = godotenv.Load()
	if err != nil {
		log.Println("No .env file found. relying on system environment")
	}

	dsn := os.Getenv("ACCOUNT_DATABASE_URL")
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database")
	}
}

func GetDBForEvent() {
	var err error = godotenv.Load()
	if err != nil {
		log.Println("No .env file found. relying on system environment")
	}

	dsn := os.Getenv("EVENT_DATABASE_URL")
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database")
	}
}

func GetDBForAttendee() {
	var err error = godotenv.Load()
	if err != nil {
		log.Println("No .env file found. relying on system environment")
	}

	dsn := os.Getenv("ATTENDEE_DATABASE_URL")
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database")
	}
}
