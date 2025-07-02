package main

import (
	"kars/database"
	"kars/routes"
	"kars/utils"
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	utils.Init()
	database.ConnectDB()
	database.MigrateModels()
	app := fiber.New()
	routes.Routes(app)
	if err := app.Listen("0.0.0.0:3000"); err != nil {
		log.Fatal("Failed to start the server:", err)
	}
}
