package model

import "time"

// Withdrawal — операция списания средств
type Withdrawal struct {
	ID          int64      `db:"id" json:"id"`
	UserID      int64      `db:"user_id" json:"user_id"`
	OrderNumber string     `db:"order_number" json:"order_number"`
	Sum         int64      `db:"sum" json:"sum"` // сотые доли балла
	ProcessedAt *time.Time `db:"processed_at" json:"processed_at,omitempty"`
}
