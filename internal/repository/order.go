package repository

import (
	"context"
	"loyalty-ledger/internal/model"
)

type OrderRepository interface {
	BatchUpdateOrderStatus(ctx context.Context, orderNumbers []string, fromStatus, toStatus string) ([]string, error)
	CreateOrder(ctx context.Context, order *model.Order) error
	GetOrderByNumber(ctx context.Context, number string) (*model.Order, error)
	GetOrdersByUser(ctx context.Context, userID int64) ([]*model.Order, error)
	GetOrdersByStatus(ctx context.Context, statuses ...string) ([]string, error)
	UpdateOrderStatus(ctx context.Context, orderNumber string, status string, accrual *int64) error
}
