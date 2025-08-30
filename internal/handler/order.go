// Package handler
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"orderkeeper/internal/models"
	"orderkeeper/internal/service"
	"orderkeeper/pkg/utils"
	"strings"

	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func validateOrder(order *models.Order) error {
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

	for _, item := range order.Items {
		if item.CHRTID <= 0 {
			return errors.New("item chrt_id must be positive")
		}
		if item.Price <= 0 {
			return errors.New("item price must be positive")
		}
		if item.Name == "" {
			return errors.New("item name is required")
		}
	}

	return nil
}

// CreateOrderHandler godoc
// @Summary Create a new order
// @Description Create a new order from JSON data
// @Tags orders
// @Accept  json
// @Produce  json
// @Param order body models.Order true "Order data"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /order [post]
func (h *OrderHandler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	if err := validateOrder(&order); err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "Validation failed: " + err.Error(),
		})
		return
	}

	err := h.orderService.CreateOrder(order)
	if err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}
	utils.JSONResponse(w, http.StatusCreated, map[string]string{
		"message": "Order successfully created",
	})
}

// GetOrderByIDHandler godoc
// @Summary Get order by ID
// @Description Get order details by order_uid
// @Tags orders
// @Accept  json
// @Produce  json
// @Param id path string true "Order ID"
// @Success 200 {object} models.Order
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /order/{id} [get]
func (h *OrderHandler) GetOrderByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "Order ID is required",
		})
		return
	}

	order, err := h.orderService.GetOrderByID(id)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			utils.JSONResponse(w, http.StatusNotFound, map[string]string{
				"error": err.Error(),
			})
		} else {
			utils.JSONResponse(w, http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}
		return
	}

	utils.JSONResponse(w, http.StatusOK, order)
}
