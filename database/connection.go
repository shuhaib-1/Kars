package database

import (
	"fmt"
	"log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	// "your_project/models"
)

var DB *gorm.DB


func ConnectDB(){
	dsn := "host=localhost user=postgres password=password dbname=kars port=5432 sslmode=disable TimeZone=Asia/Shanghai"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
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