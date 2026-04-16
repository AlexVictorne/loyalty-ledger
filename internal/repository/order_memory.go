package repository

import (
	"context"
	"loyalty-ledger/internal/model"
	"sync"
)

type InMemoryOrderRepository struct {
	mu     sync.RWMutex
	orders map[string]*model.Order // key: order number
}

func NewInMemoryOrderRepository() *InMemoryOrderRepository {
	return &InMemoryOrderRepository{
		orders: make(map[string]*model.Order),
	}
}

func (r *InMemoryOrderRepository) CreateOrder(ctx context.Context, order *model.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.orders[order.Number]; exists {
		return model.ErrOrderExists
	}
	orderCopy := *order
	r.orders[order.Number] = &orderCopy
	return nil
}

func (r *InMemoryOrderRepository) GetOrderByNumber(ctx context.Context, number string) (*model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	order, exists := r.orders[number]
	if !exists {
		return nil, nil
	}
	orderCopy := *order
	return &orderCopy, nil
}

func (r *InMemoryOrderRepository) GetOrdersByUser(ctx context.Context, userID int64) ([]*model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*model.Order
	for _, order := range r.orders {
		if order.UserID == userID {
			orderCopy := *order
			result = append(result, &orderCopy)
		}
	}
	return result, nil
}

func (r *InMemoryOrderRepository) GetOrdersByStatus(ctx context.Context, statuses ...string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	statusSet := make(map[string]struct{}, len(statuses))
	for _, s := range statuses {
		statusSet[s] = struct{}{}
	}
	var result []string
	for _, order := range r.orders {
		if _, ok := statusSet[order.Status]; ok {
			result = append(result, order.Number)
		}
	}
	return result, nil
}

func (r *InMemoryOrderRepository) UpdateOrderStatus(ctx context.Context, orderNumber string, status string, accrual *int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	order, exists := r.orders[orderNumber]
	if !exists {
		return model.ErrOrderNotFound
	}
	order.Status = status
	if accrual != nil {
		order.Accrual = *accrual
	}
	return nil
}

// BatchUpdateOrderStatus updates status for multiple orders if current status matches fromStatus.
func (r *InMemoryOrderRepository) BatchUpdateOrderStatus(ctx context.Context, orderNumbers []string, fromStatus, toStatus string) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var updated []string
	for _, num := range orderNumbers {
		order, ok := r.orders[num]
		if ok && order.Status == fromStatus {
			order.Status = toStatus
			updated = append(updated, num)
		}
	}
	return updated, nil
}
