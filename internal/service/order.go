package service

import (
	"context"
	"errors"
	"loyalty-ledger/internal/model"
	"loyalty-ledger/internal/repository"
	"sort"
	"time"
)

var (
	ErrOrderNumberEmpty = errors.New("order number required")
	ErrOrderInvalid     = errors.New("order number is invalid")
)

type OrderService struct {
	repo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

// Возможные статусы результата регистрации заказа
const (
	OrderStatusAccepted       = "accepted"
	OrderStatusDuplicateOwn   = "duplicate_own"
	OrderStatusDuplicateOther = "duplicate_other"
	OrderStatusInvalid        = "invalid"
)

func isValidLuhn(number string) bool {
	var sum int
	double := false
	for i := len(number) - 1; i >= 0; i-- {
		d := int(number[i] - '0')
		if d < 0 || d > 9 {
			return false
		}
		if double {
			d = d * 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return sum%10 == 0
}

func (s *OrderService) RegisterOrder(ctx context.Context, userID int64, number string) (string, error) {
	if number == "" {
		return OrderStatusInvalid, ErrOrderNumberEmpty
	}
	for _, c := range number {
		if c < '0' || c > '9' {
			return OrderStatusInvalid, ErrOrderInvalid
		}
	}
	if !isValidLuhn(number) {
		return OrderStatusInvalid, ErrOrderInvalid
	}
	existing, _ := s.repo.GetOrderByNumber(ctx, number)
	if existing != nil {
		if existing.UserID == userID {
			return OrderStatusDuplicateOwn, nil
		}
		return OrderStatusDuplicateOther, nil
	}
	order := &model.Order{
		Number:    number,
		UserID:    userID,
		Status:    model.OrderStatusNew,
		CreatedAt: time.Now(),
	}
	err := s.repo.CreateOrder(ctx, order)
	if err != nil {
		return "", err
	}
	return OrderStatusAccepted, nil
}

func (s *OrderService) GetOrdersByUser(ctx context.Context, userID int64) ([]*model.Order, error) {
	orders, err := s.repo.GetOrdersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].CreatedAt.After(orders[j].CreatedAt)
	})
	return orders, nil
}
