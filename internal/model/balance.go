package model

import "time"

// Balance — текущий баланс пользователя
type Balance struct {
	UserID    int64     `db:"user_id" json:"user_id"`
	Current   int64     `db:"current" json:"current"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// BalanceWithWithdrawn — DTO для баланса с суммой списаний
type BalanceWithWithdrawn struct {
	Current   int64 `json:"current"`
	Withdrawn int64 `json:"withdrawn"`
}
