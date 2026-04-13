package handler

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"loyalty-ledger/internal/model"
	"loyalty-ledger/internal/service"
	"loyalty-ledger/pkg/auth"
	"loyalty-ledger/pkg/points"
	"net/http"
)

type BalanceRoutes interface {
	GetBalance(w http.ResponseWriter, r *http.Request)
	Withdraw(w http.ResponseWriter, r *http.Request)
	GetWithdrawals(w http.ResponseWriter, r *http.Request)
}

type BalanceServiceIface interface {
	GetBalanceWithWithdrawn(ctx context.Context, userID int64) (*model.BalanceWithWithdrawn, error)
	Withdraw(ctx context.Context, userID int64, order string, sum int64) error
	GetWithdrawals(ctx context.Context, userID int64) ([]*model.Withdrawal, error)
}

type BalanceHandler struct {
	service BalanceServiceIface
}

func NewBalanceHandler(service BalanceServiceIface) *BalanceHandler {
	return &BalanceHandler{service: service}
}

// GET /api/user/balance
func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	info, ok := auth.GetAuthInfo(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	bw, err := h.service.GetBalanceWithWithdrawn(r.Context(), info.UserID)
	if err != nil {
		log.Printf("get balance error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	resp := BalanceResponse{
		Current:   points.ToAPI(bw.Current),
		Withdrawn: points.ToAPI(bw.Withdrawn),
	}
	json.NewEncoder(w).Encode(resp)
}

// POST /api/user/balance/withdraw
func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	info, ok := auth.GetAuthInfo(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var req WithdrawRequest
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sumInt := points.ToInternal(req.Sum)
	err = h.service.Withdraw(r.Context(), info.UserID, req.Order, sumInt)
	switch err {
	case nil:
		w.WriteHeader(http.StatusOK)
	case service.ErrInsufficientFunds:
		w.WriteHeader(http.StatusPaymentRequired)
	case service.ErrInvalidSum, service.ErrOrderNumberRequired, service.ErrOrderAlreadyWithdrawn:
		w.WriteHeader(http.StatusUnprocessableEntity)
	default:
		log.Printf("withdraw error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// GET /api/user/withdrawals
func (h *BalanceHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	info, ok := auth.GetAuthInfo(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	list, err := h.service.GetWithdrawals(r.Context(), info.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(list) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	// Преобразуем суммы к float64 для API
	resp := make([]WithdrawalResponse, len(list))
	for i, w := range list {
		resp[i] = WithdrawalResponse{
			Order:       w.OrderNumber,
			Sum:         points.ToAPI(w.Sum),
			ProcessedAt: w.ProcessedAt,
		}
	}
	json.NewEncoder(w).Encode(resp)
}
