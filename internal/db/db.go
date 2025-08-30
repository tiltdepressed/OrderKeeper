// Package db
package db

import (
	"log"
	"math"
	"orderkeeper/internal/models"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	err error
)

func InitDB() (*gorm.DB, error) {
	dsn := os.Getenv("DSN")
	maxAttempts := 5
	initialDelay := 2 * time.Second
	var dbInstance *gorm.DB
	var err error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		dbInstance, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			// Миграция моделей
			err = dbInstance.AutoMigrate(
				&models.Order{},
				&models.Delivery{},
				&models.Payment{},
				&models.Item{},
			)
			if err != nil {
				return nil, err
			}
			return dbInstance, nil
		}

		log.Printf("Attempt %d/%d: failed to connect to database: %v", attempt, maxAttempts, err)
		if attempt < maxAttempts {
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * initialDelay
			log.Printf("Retrying in %v...", delay)
			time.Sleep(delay)
		}
	}

	return nil, err
}
