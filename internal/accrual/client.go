package accrual

import (
	"context"
)

// OrderStatus — возможные статусы заказа в accrual system
const (
	StatusRegistered = "REGISTERED"
	StatusInvalid    = "INVALID"
	StatusProcessing = "PROCESSING"
	StatusProcessed  = "PROCESSED"
)

// OrderInfo — структура ответа accrual system
// accrual может отсутствовать, если не начислено
// processed_at не входит в спецификацию accrual system, но может быть полезен для локального хранения
// (оставляем только то, что приходит с accrual system)
type OrderInfo struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

// Client интерфейс для работы с accrual system
// Может быть реализован реальным HTTP-клиентом или мок-реализацией для тестов
type Client interface {
	GetOrderInfo(ctx context.Context, orderNumber string) (OrderInfo, error)
}
