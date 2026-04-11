package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"loyalty-ledger/internal/model"
	"loyalty-ledger/internal/repository"
	"loyalty-ledger/internal/service"
	"loyalty-ledger/pkg/auth"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOrderHandler_RegisterOrder(t *testing.T) {
	userID := int64(42)

	// Корректный номер по Луну: 79927398713
	// Некорректный по Луну: 79927398710
	tests := []struct {
		name          string
		body          string
		contentType   string
		withAuth      bool
		preInsert     bool
		preInsertUser int64
		wantStatus    int
	}{
		{"unauthorized", "79927398713", "text/plain", false, false, 0, http.StatusUnauthorized},
		{"empty number", "", "text/plain", true, false, 0, http.StatusUnprocessableEntity},
		{"bad content-type", "79927398713", "application/json", true, false, 0, http.StatusBadRequest},
		{"invalid luhn", "79927398710", "text/plain", true, false, 0, http.StatusUnprocessableEntity},
		{"success", "79927398713", "text/plain", true, false, 0, http.StatusAccepted},
		{"duplicate own", "79927398713", "text/plain", true, true, 42, http.StatusOK},
		{"duplicate other", "79927398713", "text/plain", true, true, 99, http.StatusConflict},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := repository.NewInMemoryOrderRepository()
			svc := service.NewOrderService(repo)
			h := NewOrderHandler(svc)
			if tc.preInsert {
				_ = repo.CreateOrder(context.Background(), &model.Order{Number: "79927398713", UserID: tc.preInsertUser})
			}
			req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", tc.contentType)
			if tc.withAuth {
				ctx := auth.SetAuthInfo(req.Context(), &auth.AuthInfo{UserID: userID, Login: "testuser"})
				req = req.WithContext(ctx)
			}
			rw := httptest.NewRecorder()
			h.RegisterOrder(rw, req)
			if rw.Code != tc.wantStatus {
				t.Errorf("expected status %d, got %d", tc.wantStatus, rw.Code)
			}
		})
	}
}

func TestOrderHandler_GetOrders(t *testing.T) {
	repo := repository.NewInMemoryOrderRepository()
	svc := service.NewOrderService(repo)
	h := NewOrderHandler(svc)
	userID := int64(42)

	t.Run("unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
		rw := httptest.NewRecorder()
		h.GetOrders(rw, req)
		if rw.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rw.Code)
		}
	})

	t.Run("no orders", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
		ctx := auth.SetAuthInfo(req.Context(), &auth.AuthInfo{UserID: userID, Login: "testuser"})
		req = req.WithContext(ctx)
		rw := httptest.NewRecorder()
		h.GetOrders(rw, req)
		if rw.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", rw.Code)
		}
	})

	t.Run("one order", func(t *testing.T) {
		_ = repo.CreateOrder(context.Background(), &model.Order{Number: "123", UserID: userID})
		req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
		ctx := auth.SetAuthInfo(req.Context(), &auth.AuthInfo{UserID: userID, Login: "testuser"})
		req = req.WithContext(ctx)
		rw := httptest.NewRecorder()
		h.GetOrders(rw, req)
		if rw.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rw.Code)
		}
		var orders []model.Order
		if err := json.NewDecoder(rw.Body).Decode(&orders); err != nil {
			t.Errorf("response decode error: %v", err)
		}
		if len(orders) != 1 || orders[0].Number != "123" {
			t.Errorf("unexpected orders: %+v", orders)
		}
	})

	t.Run("orders sorted and date RFC3339", func(t *testing.T) {
		repo := repository.NewInMemoryOrderRepository()
		svc := service.NewOrderService(repo)
		h := NewOrderHandler(svc)
		now := time.Now()
		order1 := &model.Order{Number: "1", UserID: userID, CreatedAt: now.Add(-2 * time.Hour)}
		order2 := &model.Order{Number: "2", UserID: userID, CreatedAt: now.Add(-1 * time.Hour)}
		order3 := &model.Order{Number: "3", UserID: userID, CreatedAt: now}
		_ = repo.CreateOrder(context.Background(), order1)
		_ = repo.CreateOrder(context.Background(), order2)
		_ = repo.CreateOrder(context.Background(), order3)
		req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
		ctx := auth.SetAuthInfo(req.Context(), &auth.AuthInfo{UserID: userID, Login: "testuser"})
		req = req.WithContext(ctx)
		rw := httptest.NewRecorder()
		h.GetOrders(rw, req)
		if rw.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rw.Code)
		}
		var orders []map[string]interface{}
		if err := json.NewDecoder(rw.Body).Decode(&orders); err != nil {
			t.Errorf("response decode error: %v", err)
		}
		if len(orders) != 3 {
			t.Fatalf("expected 3 orders, got %d", len(orders))
		}
		if !(orders[0]["number"] == "3" && orders[1]["number"] == "2" && orders[2]["number"] == "1") {
			t.Errorf("orders not sorted by CreatedAt desc: got %v", []interface{}{orders[0]["number"], orders[1]["number"], orders[2]["number"]})
		}
		// Проверка формата даты RFC3339
		for _, o := range orders {
			dateStr, ok := o["uploaded_at"].(string)
			if !ok {
				t.Errorf("uploaded_at not a string: %v", o["uploaded_at"])
				continue
			}
			if _, err := time.Parse(time.RFC3339, dateStr); err != nil {
				t.Errorf("uploaded_at not RFC3339: %s", dateStr)
			}
		}
	})
}
