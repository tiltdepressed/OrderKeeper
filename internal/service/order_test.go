package service

import (
	"errors"
	"orderkeeper/internal/cache"
	"orderkeeper/internal/models"
	"orderkeeper/internal/repository/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestOrderService_GetOrderByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockOrderRepository(ctrl)
	orderCache := cache.NewOrderCache()

	orderService := NewOrderService(mockRepo, orderCache)

	testOrder := models.Order{OrderUID: "test-uid-123"}

	t.Run("found in cache", func(t *testing.T) {
		orderCache.Set(testOrder)

		result, err := orderService.GetOrderByID("test-uid-123")

		assert.NoError(t, err)
		assert.Equal(t, testOrder.OrderUID, result.OrderUID)
	})

	t.Run("not in cache, found in db", func(t *testing.T) {
		cleanCache := cache.NewOrderCache()
		serviceWithCleanCache := NewOrderService(mockRepo, cleanCache)

		mockRepo.EXPECT().GetOrderByID("test-uid-123").Return(testOrder, nil)

		result, err := serviceWithCleanCache.GetOrderByID("test-uid-123")

		assert.NoError(t, err)
		assert.Equal(t, testOrder.OrderUID, result.OrderUID)

		_, exists := cleanCache.Get("test-uid-123")
		assert.True(t, exists)
	})

	t.Run("not found anywhere", func(t *testing.T) {
		cleanCache := cache.NewOrderCache()
		serviceWithCleanCache := NewOrderService(mockRepo, cleanCache)

		mockRepo.EXPECT().GetOrderByID("non-existent-uid").Return(models.Order{}, gorm.ErrRecordNotFound)

		_, err := serviceWithCleanCache.GetOrderByID("non-existent-uid")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrOrderNotFound))
	})
}
