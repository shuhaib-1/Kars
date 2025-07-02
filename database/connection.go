package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	// "your_project/models"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	time.Sleep(2 * time.Second)
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	DB = db

	// Ping the database to check the connection
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}

	// Use Ping to verify the connection
	err = sqlDB.Ping()
	if err != nil {
		log.Fatal("Database connection check failed:", err)
	} else {
		fmt.Println("Database connected successfully")
	}
}
