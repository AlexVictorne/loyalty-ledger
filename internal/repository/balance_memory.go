package repository

import (
	"context"
	"errors"
	"loyalty-ledger/internal/model"
	"sync"
	"time"
)

type InMemoryBalanceRepository struct {
	mu          sync.RWMutex
	balances    map[int64]*model.Balance
	withdrawals map[int64][]*model.Withdrawal
}

func NewInMemoryBalanceRepository() *InMemoryBalanceRepository {
	return &InMemoryBalanceRepository{
		balances:    make(map[int64]*model.Balance),
		withdrawals: make(map[int64][]*model.Withdrawal),
	}
}

func (r *InMemoryBalanceRepository) GetBalance(ctx context.Context, userID int64) (*model.Balance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.balances[userID]
	if !ok {
		return &model.Balance{UserID: userID, Current: 0, UpdatedAt: time.Now()}, nil
	}
	copy := *b
	return &copy, nil
}

func (r *InMemoryBalanceRepository) Accrue(ctx context.Context, userID int64, sum int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.balances[userID]
	now := time.Now()
	if !ok {
		b = &model.Balance{UserID: userID, Current: sum, UpdatedAt: now}
		r.balances[userID] = b
	} else {
		b.Current += sum
		b.UpdatedAt = now
	}
	return nil
}

func (r *InMemoryBalanceRepository) Withdraw(ctx context.Context, userID int64, orderNumber string, sum int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.balances[userID]
	if !ok || b.Current < sum {
		return errors.New("insufficient funds")
	}
	// deduplication: check if orderNumber already withdrawn
	for _, w := range r.withdrawals[userID] {
		if w.OrderNumber == orderNumber {
			return errors.New("order already withdrawn")
		}
	}
	b.Current -= sum
	b.UpdatedAt = time.Now()
	w := &model.Withdrawal{
		ID:          int64(len(r.withdrawals[userID]) + 1),
		UserID:      userID,
		OrderNumber: orderNumber,
		Sum:         sum,
		ProcessedAt: ptrTime(time.Now()),
	}
	r.withdrawals[userID] = append(r.withdrawals[userID], w)
	return nil
}

func (r *InMemoryBalanceRepository) GetWithdrawals(ctx context.Context, userID int64) ([]*model.Withdrawal, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := r.withdrawals[userID]
	result := make([]*model.Withdrawal, len(list))
	for i, w := range list {
		copy := *w
		result[i] = &copy
	}
	return result, nil
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
