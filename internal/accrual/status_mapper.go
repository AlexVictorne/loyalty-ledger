package accrual

import (
	"fmt"
	"loyalty-ledger/internal/model"
)

// Конвертируем статус внешней accrual-системы во внутренний статус заказа
func MapToInternalStatus(external string) (string, error) {
	switch external {
	case StatusProcessed:
		return model.OrderStatusProcessed, nil
	case StatusInvalid:
		return model.OrderStatusInvalid, nil
	case StatusProcessing:
		return model.OrderStatusProcessing, nil
	case StatusRegistered:
		return model.OrderStatusNew, nil
	default:
		return "", fmt.Errorf("unknown external accrual status: %q", external)
	}
}
