package service

import (
	"context"
	"errors"
	"loyalty-ledger/internal/model"
	"loyalty-ledger/internal/repository"
	"loyalty-ledger/pkg/validate"
	"sort"
)

var (
	ErrInsufficientFunds     = errors.New("insufficient funds")
	ErrInvalidSum            = errors.New("sum must be positive")
	ErrOrderNumberRequired   = errors.New("order number required")
	ErrOrderAlreadyWithdrawn = errors.New("order already withdrawn")
)

type BalanceService struct {
	repo repository.BalanceRepository
}

func NewBalanceService(repo repository.BalanceRepository) *BalanceService {
	return &BalanceService{repo: repo}
}

func (s *BalanceService) GetBalance(ctx context.Context, userID int64) (*model.Balance, error) {
	return s.repo.GetBalance(ctx, userID)
}

func (s *BalanceService) Accrue(ctx context.Context, userID int64, sum int64) error {
	if sum <= 0 {
		return ErrInvalidSum
	}
	return s.repo.Accrue(ctx, userID, sum)
}

func (s *BalanceService) Withdraw(ctx context.Context, userID int64, orderNumber string, sum int64) error {
	if sum <= 0 {
		return ErrInvalidSum
	}
	if orderNumber == "" {
		return ErrOrderNumberRequired
	}
	if !validate.IsValidLuhn(orderNumber) {
		return ErrOrderNumberRequired
	}
	err := s.repo.Withdraw(ctx, userID, orderNumber, sum)
	if err != nil {
		switch err.Error() {
		case "insufficient funds":
			return ErrInsufficientFunds
		case "order already withdrawn":
			return ErrOrderAlreadyWithdrawn
		}
	}
	return err
}

func (s *BalanceService) GetWithdrawals(ctx context.Context, userID int64) ([]*model.Withdrawal, error) {
	list, err := s.repo.GetWithdrawals(ctx, userID)
	if err != nil {
		return nil, err
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].ProcessedAt == nil {
			return false
		}
		if list[j].ProcessedAt == nil {
			return true
		}
		return list[i].ProcessedAt.After(*list[j].ProcessedAt)
	})
	return list, nil
}

func (s *BalanceService) GetBalanceWithWithdrawn(ctx context.Context, userID int64) (*model.BalanceWithWithdrawn, error) {
	bal, err := s.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}
	withdrawals, err := s.GetWithdrawals(ctx, userID)
	if err != nil {
		return nil, err
	}
	var withdrawn int64
	for _, w := range withdrawals {
		withdrawn += w.Sum
	}
	return &model.BalanceWithWithdrawn{
		Current:   bal.Current,
		Withdrawn: withdrawn,
	}, nil
}
