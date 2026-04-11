package model

import (
	"errors"
	"time"
)

var ErrOrderExists = errors.New("order already exists")

// Order — бизнес-сущность заказа
type Order struct {
	ID        int64     `db:"id"`
	Number    string    `db:"number" json:"number"`
	UserID    int64     `db:"user_id"`
	Status    string    `db:"status" json:"status"`
	Accrual   int64     `db:"accrual" json:"accrual,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"uploaded_at"`
}

const (
	OrderStatusNew        = "NEW"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusProcessed  = "PROCESSED"
	OrderStatusInvalid    = "INVALID"
)

type OrderRequest struct {
	Number string `json:"number"`
}
