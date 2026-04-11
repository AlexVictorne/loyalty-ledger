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
