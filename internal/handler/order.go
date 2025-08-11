// Package handler
package handler

import (
	"encoding/json"
	"net/http"
	"orderkeeper/internal/models"
	"orderkeeper/internal/service"
	"orderkeeper/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
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
			"error": "Invalid request body",
		})
		return
	}

	if order.OrderUID == "" {
		utils.JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "OrderUID is required",
		})
		return
	}

	order.Delivery.OrderUID = order.OrderUID
	order.Payment.OrderUID = order.OrderUID
	for i := range order.Items {
		order.Items[i].OrderUID = order.OrderUID
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
		utils.JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	utils.JSONResponse(w, http.StatusOK, order)
}
