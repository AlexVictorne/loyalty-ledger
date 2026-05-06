package repository

import (
	"context"
	"loyalty-ledger/internal/model"
)

type BalanceRepository interface {
	GetBalance(ctx context.Context, userID int64) (*model.Balance, error)
	Accrue(ctx context.Context, userID int64, sum int64) error
	Withdraw(ctx context.Context, userID int64, orderNumber string, sum int64) error
	GetWithdrawals(ctx context.Context, userID int64) ([]*model.Withdrawal, error)
}
