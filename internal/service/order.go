//go:generate go run go.uber.org/mock/mockgen -source=order.go -destination=mocks/mock_order_service.go -package=mocks
package service

import (
	"errors"
	"orderkeeper/internal/cache"
	"orderkeeper/internal/models"
	"orderkeeper/internal/repository"
	"strings"

	"gorm.io/gorm"
)

var ErrOrderNotFound = errors.New("order not found")

type OrderService interface {
	CreateOrder(order models.Order) error
	GetOrderByID(id string) (models.Order, error)
	RestoreCache() error
}

type orderService struct {
	repo  repository.OrderRepository
	cache *cache.OrderCache
}

func NewOrderService(repo repository.OrderRepository, cache *cache.OrderCache) OrderService {
	return &orderService{repo: repo, cache: cache}
}

func (s *orderService) validateOrder(order *models.Order) error {
	if order.OrderUID == "" {
		return errors.New("order_uid is required")
	}
	if order.TrackNumber == "" {
		return errors.New("track_number is required")
	}
	if order.Delivery.Name == "" {
		return errors.New("delivery name is required")
	}
	if order.Delivery.Phone == "" {
		return errors.New("delivery phone is required")
	}
	if !strings.Contains(order.Delivery.Email, "@") {
		return errors.New("delivery email is invalid")
	}
	if order.Payment.Transaction == "" {
		return errors.New("payment transaction is required")
	}
	if order.Payment.Amount <= 0 {
		return errors.New("payment amount must be positive")
	}
	if len(order.Items) == 0 {
		return errors.New("order must contain at least one item")
	}
	return nil
}

func (s *orderService) CreateOrder(order models.Order) error {
	if err := s.validateOrder(&order); err != nil {
		return err
	}

	if err := s.repo.CreateOrder(order); err != nil {
		return err
	}

	s.cache.Set(order)
	return nil
}

func (s *orderService) GetOrderByID(id string) (models.Order, error) {
	if order, exists := s.cache.Get(id); exists {
		return order, nil
	}

	order, err := s.repo.GetOrderByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Order{}, ErrOrderNotFound
		}
		return models.Order{}, err
	}

	s.cache.Set(order)
	return order, nil
}

func (s *orderService) RestoreCache() error {
	orders, err := s.repo.GetAllOrders()
	if err != nil {
		return err
	}

	s.cache.LoadFromDB(orders)
	return nil
}
