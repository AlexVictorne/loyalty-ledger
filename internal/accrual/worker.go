package accrual

import (
	"context"
	"time"
)

type OrderAccrualService interface {
	GetNewOrders(ctx context.Context) ([]string, error)
	GetProcessingOrders(ctx context.Context) ([]string, error)
	UpdateOrderAndBalanceFromAccrual(ctx context.Context, orderNumber string, status string, accrual *float64) error
	BatchSetOrdersProcessing(ctx context.Context, orderNumbers []string) ([]string, error)
}

type WorkerConfig struct {
	Interval   time.Duration // интервал между запусками
	NumWorkers int           // число параллельных воркеров
}
