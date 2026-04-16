package service

import (
	"context"
	"errors"
	"loyalty-ledger/internal/model"
	"loyalty-ledger/internal/repository"
	"loyalty-ledger/pkg/validate"
	"sort"
)

// Ошибки теперь в model/errors.go

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
		return model.ErrInvalidSum
	}
	return s.repo.Accrue(ctx, userID, sum)
}

func (s *BalanceService) Withdraw(ctx context.Context, userID int64, orderNumber string, sum int64) error {
	if sum <= 0 {
		return model.ErrInvalidSum
	}
	if orderNumber == "" {
		return model.ErrOrderNumberRequired
	}
	if !validate.IsValidLuhn(orderNumber) {
		return model.ErrOrderNumberRequired
	}
	err := s.repo.Withdraw(ctx, userID, orderNumber, sum)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInsufficientFunds):
			return model.ErrInsufficientFunds
		case errors.Is(err, model.ErrOrderAlreadyWithdrawn):
			return model.ErrOrderAlreadyWithdrawn
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
