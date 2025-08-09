// Package repository
package repository

import (
	"orderkeeper/internal/models"

	"gorm.io/gorm"
)

type OrderRepository interface {
	CreateOrder(order models.Order) error
	GetAllOrders() ([]models.Order, error)
	GetOrderByID(id string) (models.Order, error)
}

type orderRepo struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepo{db: db}
}

func (r *orderRepo) CreateOrder(order models.Order) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *orderRepo) GetAllOrders() ([]models.Order, error) {
	var orders []models.Order
	err := r.db.
		Preload("Delivery").
		Preload("Payment").
		Preload("Items").
		Find(&orders).Error
	return orders, err
}

func (r *orderRepo) GetOrderByID(id string) (models.Order, error) {
	var order models.Order
	err := r.db.
		Preload("Delivery").
		Preload("Payment").
		Preload("Items").
		First(&order, "order_uid = ?", id).Error
	return order, err
}
