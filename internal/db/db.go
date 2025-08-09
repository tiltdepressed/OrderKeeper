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

	for i := range 5 {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Connection attempt %d failed: %v", i+1, err)
		time.Sleep(time.Second * time.Duration(math.Pow(2, float64(i))))
	}
	if err = db.AutoMigrate(&models.Delivery{}, &models.Payment{}, &models.Item{}, &models.Order{}); err != nil {
		log.Fatalf("Could not migrate: %v", err)
	}
	return db, nil
}
