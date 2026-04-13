package handler

import "time"

// BalanceResponse — структура ответа для GET /api/user/balance
//
//	{
//	  "current": 500.0,
//	  "withdrawn": 500.5
//	}
type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// WithdrawalResponse — структура ответа для GET /api/user/withdrawals
// [
//
//	{"order": "12345678903", "sum": 300.75, "processed_at": "2023-03-10T10:00:00Z"},
//	...
//
// ]
type WithdrawalResponse struct {
	Order       string     `json:"order"`
	Sum         float64    `json:"sum"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}

// WithdrawRequest — структура запроса для POST /api/user/balance/withdraw
type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}
