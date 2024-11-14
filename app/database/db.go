package database

import (
	"fmt"
	"fusion/app/database/models"
	"fusion/app/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func ConnectDB(config utils.AppConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.DatabaseHost, config.DatabasePort, config.DatabaseUser, config.DatabasePassword, config.DatabaseName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
	})

	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		log.Fatalf("failed to create extension uuid-ossp: %v", err)
		return nil, err
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Permissions{},
		&models.Session{},
		&models.Verification{},
		&models.Product{},
		&models.Category{},
		&models.Review{},
		&models.Favourite{},
		&models.Cart{},
		&models.Order{},
		&models.OrderProduct{},
	); err != nil {
		return nil, err
	}

	return db, nil
}
