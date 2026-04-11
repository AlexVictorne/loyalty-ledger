package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"loyalty-ledger/internal/service"
	"loyalty-ledger/pkg/auth"
	"net/http"
)

type OrderRoutes interface {
	RegisterOrder(w http.ResponseWriter, r *http.Request)
	GetOrders(w http.ResponseWriter, r *http.Request)
}

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(service *service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) RegisterOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("Content-Type") != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	info, ok := auth.GetAuthInfo(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	number := string(bytes.TrimSpace(body))
	status, err := h.service.RegisterOrder(context.Background(), info.UserID, number)
	switch status {
	case service.OrderStatusInvalid:
		w.WriteHeader(http.StatusUnprocessableEntity) // 422
	case service.OrderStatusDuplicateOwn:
		w.WriteHeader(http.StatusOK)
	case service.OrderStatusDuplicateOther:
		w.WriteHeader(http.StatusConflict)
	case service.OrderStatusAccepted:
		w.WriteHeader(http.StatusAccepted)
	default:
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	info, ok := auth.GetAuthInfo(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	orders, err := h.service.GetOrdersByUser(context.Background(), info.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(orders)
}
