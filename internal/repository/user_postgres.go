package repository

import (
	"context"
	"errors"
	"loyalty-ledger/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id, created_at`
	row := r.pool.QueryRow(ctx, query, user.Login, user.Password)
	err := row.Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		// Проверяем ошибку уникальности (duplicate key)
		var pgErr interface{ SQLState() string }
		if errors.As(err, &pgErr) && pgErr.SQLState() == "23505" {
			return model.ErrUserExists
		}
		return err
	}
	return nil
}

func (r *PostgresUserRepository) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, login, password_hash, created_at FROM users WHERE login = $1`
	row := r.pool.QueryRow(ctx, query, login)
	err := row.Scan(&user.ID, &user.Login, &user.Password, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}
