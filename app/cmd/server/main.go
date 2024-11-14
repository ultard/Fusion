package main

import (
	"encoding/json"
	"fmt"
	"fusion/app/database"
	"fusion/app/handlers"
	"fusion/app/middleware"
	"fusion/app/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"log"
)

func main() {
	var config utils.AppConfig
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	email := utils.NewEmailService(
		config.SmtpHost,
		config.SmtpPort,
		config.SmtpUser,
		config.SmtpPassword,
		config.SmtpSender,
	)

	jwt := utils.NewJWTService(config.SessionSecret)
	db, err := database.ConnectDB(config)

	app := fiber.New(fiber.Config{
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
		ServerHeader: "Fusion",
		AppName:      fmt.Sprintf("Fusion App v%s", config.AppVersion),
	})

	app.Use(cors.New())
	app.Use(recover.New())

	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	app.Use(middleware.InjectorMiddleware(config, db, jwt, email))
	handlers.RegisterAuthRoutes(app, db, config, jwt, email)
	handlers.RegisterUserRoutes(app, db)
	handlers.RegisterProductRoutes(app, db)
	handlers.RegisterOrderRoutes(app, db)
	handlers.RegisterCartRoute(app, db)

	app.Listen(":" + config.AppPort)
	defer app.Shutdown()
}
