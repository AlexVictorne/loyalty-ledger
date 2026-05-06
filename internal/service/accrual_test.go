package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"loyalty-ledger/internal/model"
	"loyalty-ledger/internal/repository"
)

func TestAccrualService_InMemory(t *testing.T) {
	t.Run("happy path: accrual and status update", func(t *testing.T) {
		ctx := context.Background()
		orderRepo := repository.NewInMemoryOrderRepository()
		balanceRepo := repository.NewInMemoryBalanceRepository()
		svc := NewAccrualService(orderRepo, balanceRepo)

		order := &model.Order{Number: "123", UserID: 42, Status: model.OrderStatusProcessing, Accrual: 0}
		err := orderRepo.CreateOrder(ctx, order)
		require.NoError(t, err)

		accrual := 10.5
		err = svc.UpdateOrderAndBalanceFromAccrual(ctx, "123", model.OrderStatusProcessed, &accrual)
		require.NoError(t, err)

		updatedOrder, err := orderRepo.GetOrderByNumber(ctx, "123")
		require.NoError(t, err)
		assert.Equal(t, model.OrderStatusProcessed, updatedOrder.Status)
		assert.Equal(t, int64(1050), updatedOrder.Accrual)

		balance, err := balanceRepo.GetBalance(ctx, 42)
		require.NoError(t, err)
		assert.Equal(t, int64(1050), balance.Current)
	})

	t.Run("no accrual: only status update", func(t *testing.T) {
		ctx := context.Background()
		orderRepo := repository.NewInMemoryOrderRepository()
		balanceRepo := repository.NewInMemoryBalanceRepository()
		svc := NewAccrualService(orderRepo, balanceRepo)

		order := &model.Order{Number: "124", UserID: 43, Status: model.OrderStatusProcessing, Accrual: 0}
		err := orderRepo.CreateOrder(ctx, order)
		require.NoError(t, err)

		err = svc.UpdateOrderAndBalanceFromAccrual(ctx, "124", model.OrderStatusProcessed, nil)
		require.NoError(t, err)

		updatedOrder, err := orderRepo.GetOrderByNumber(ctx, "124")
		require.NoError(t, err)
		assert.Equal(t, model.OrderStatusProcessed, updatedOrder.Status)
		assert.Equal(t, int64(0), updatedOrder.Accrual)

		balance, err := balanceRepo.GetBalance(ctx, 43)
		require.NoError(t, err)
		assert.Equal(t, int64(0), balance.Current)
	})

	t.Run("order not found", func(t *testing.T) {
		ctx := context.Background()
		orderRepo := repository.NewInMemoryOrderRepository()
		balanceRepo := repository.NewInMemoryBalanceRepository()
		svc := NewAccrualService(orderRepo, balanceRepo)

		err := svc.UpdateOrderAndBalanceFromAccrual(ctx, "notfound", model.OrderStatusProcessed, nil)
		assert.Error(t, err)
	})
}
