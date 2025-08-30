package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"orderkeeper/internal/models"
	"orderkeeper/internal/service"
	"orderkeeper/internal/service/mocks"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestOrderHandler_GetOrderByIDHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockOrderService(ctrl)
	orderHandler := NewOrderHandler(mockService)

	router := chi.NewRouter()
	router.Get("/order/{id}", orderHandler.GetOrderByIDHandler)

	t.Run("success 200 OK", func(t *testing.T) {
		testOrder := models.Order{OrderUID: "test-ok-123"}

		mockService.EXPECT().GetOrderByID("test-ok-123").Return(testOrder, nil)

		req := httptest.NewRequest(http.MethodGet, "/order/test-ok-123", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var returnedOrder models.Order
		err := json.Unmarshal(rr.Body.Bytes(), &returnedOrder)
		assert.NoError(t, err)
		assert.Equal(t, testOrder.OrderUID, returnedOrder.OrderUID)
	})

	t.Run("not found 404", func(t *testing.T) {
		mockService.EXPECT().GetOrderByID("test-404").Return(models.Order{}, service.ErrOrderNotFound)

		req := httptest.NewRequest(http.MethodGet, "/order/test-404", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("internal error 500", func(t *testing.T) {
		mockService.EXPECT().GetOrderByID("test-500").Return(models.Order{}, errors.New("some unexpected db error"))

		req := httptest.NewRequest(http.MethodGet, "/order/test-500", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
