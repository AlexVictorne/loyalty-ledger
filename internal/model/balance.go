package model

import "time"

// Balance — текущий баланс пользователя
type Balance struct {
	UserID    int64     `db:"user_id" json:"user_id"`
	Current   int64     `db:"current" json:"current"` // хранится в сотых долях балла
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// BalanceWithWithdrawn — DTO для баланса с суммой списаний (в сотых долях балла)
type BalanceWithWithdrawn struct {
	Current   int64 `json:"current"`   // сотые доли
	Withdrawn int64 `json:"withdrawn"` // сотые доли
}
