package repository

import (
	"context"
	"errors"
	"loyalty-ledger/internal/model"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresBalanceRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresBalanceRepository(pool *pgxpool.Pool) *PostgresBalanceRepository {
	return &PostgresBalanceRepository{pool: pool}
}

func (r *PostgresBalanceRepository) GetBalance(ctx context.Context, userID int64) (*model.Balance, error) {
	row := r.pool.QueryRow(ctx, `SELECT user_id, current, updated_at FROM balances WHERE user_id = $1`, userID)
	b := &model.Balance{}
	if err := row.Scan(&b.UserID, &b.Current, &b.UpdatedAt); err != nil {
		return nil, err
	}
	return b, nil
}

func (r *PostgresBalanceRepository) Accrue(ctx context.Context, userID int64, sum int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	now := time.Now()
	_, err = tx.Exec(ctx, `INSERT INTO balances (user_id, current, updated_at) VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE SET current = balances.current + $2, updated_at = $3`, userID, sum, now)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PostgresBalanceRepository) Withdraw(ctx context.Context, userID int64, orderNumber string, sum int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var current int64
	row := tx.QueryRow(ctx, `SELECT current FROM balances WHERE user_id = $1 FOR UPDATE`, userID)
	if err := row.Scan(&current); err != nil {
		return err
	}
	if current < sum {
		return errors.New("insufficient funds")
	}
	now := time.Now()
	_, err = tx.Exec(ctx, `UPDATE balances SET current = current - $1, updated_at = $2 WHERE user_id = $3`, sum, now, userID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO withdrawals (user_id, order_number, sum, processed_at) VALUES ($1, $2, $3, $4)`, userID, orderNumber, sum, now)
	if err != nil {
		// check for unique violation (duplicate order)
		if pgErr, ok := err.(interface{ SQLState() string }); ok && pgErr.SQLState() == "23505" {
			return errors.New("order already withdrawn")
		}
		return err
	}
	return tx.Commit(ctx)
}

func (r *PostgresBalanceRepository) GetWithdrawals(ctx context.Context, userID int64) ([]*model.Withdrawal, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, user_id, order_number, sum, processed_at FROM withdrawals WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*model.Withdrawal
	for rows.Next() {
		w := &model.Withdrawal{}
		if err := rows.Scan(&w.ID, &w.UserID, &w.OrderNumber, &w.Sum, &w.ProcessedAt); err != nil {
			return nil, err
		}
		result = append(result, w)
	}
	return result, nil
}
