package service

import (
	"context"
	"errors"
	"loyalty-ledger/internal/model"
	"loyalty-ledger/internal/repository"
	"testing"
	"time"
)

func TestOrderService_RegisterOrder(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)

	tests := []struct {
		name          string
		number        string
		preInsert     bool
		preInsertUser int64
		wantStatus    string
		wantErr       error
	}{
		{"empty number", "", false, 0, OrderStatusInvalid, ErrOrderNumberEmpty},
		{"invalid luhn", "79927398710", false, 0, OrderStatusInvalid, ErrOrderInvalid},
		{"success", "79927398713", false, 0, OrderStatusAccepted, nil},
		{"duplicate own", "79927398713", true, 1, OrderStatusDuplicateOwn, nil},
		{"duplicate other", "79927398713", true, 99, OrderStatusDuplicateOther, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := repository.NewInMemoryOrderRepository()
			svc := NewOrderService(repo)
			if tc.preInsert {
				_ = repo.CreateOrder(ctx, &model.Order{Number: "79927398713", UserID: tc.preInsertUser})
			}
			status, err := svc.RegisterOrder(ctx, userID, tc.number)
			if status != tc.wantStatus {
				t.Errorf("expected status %s, got %s", tc.wantStatus, status)
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("expected error %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestOrderService_GetOrdersByUser(t *testing.T) {
	repo := repository.NewInMemoryOrderRepository()
	svc := NewOrderService(repo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("empty list", func(t *testing.T) {
		orders, err := svc.GetOrdersByUser(ctx, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(orders) != 0 {
			t.Errorf("expected 0 orders, got %d", len(orders))
		}
	})

	t.Run("one order", func(t *testing.T) {
		_, _ = svc.RegisterOrder(ctx, userID, "79927398713")
		orders, err := svc.GetOrdersByUser(ctx, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(orders) != 1 {
			t.Errorf("expected 1 order, got %d", len(orders))
		}
	})

	t.Run("orders sorted by CreatedAt desc", func(t *testing.T) {
		repo := repository.NewInMemoryOrderRepository()
		svc := NewOrderService(repo)
		now := time.Now()
		order1 := &model.Order{Number: "1", UserID: userID, CreatedAt: now.Add(-2 * time.Hour)}
		order2 := &model.Order{Number: "2", UserID: userID, CreatedAt: now.Add(-1 * time.Hour)}
		order3 := &model.Order{Number: "3", UserID: userID, CreatedAt: now}
		_ = repo.CreateOrder(ctx, order1)
		_ = repo.CreateOrder(ctx, order2)
		_ = repo.CreateOrder(ctx, order3)
		orders, err := svc.GetOrdersByUser(ctx, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(orders) != 3 {
			t.Fatalf("expected 3 orders, got %d", len(orders))
		}
		if !(orders[0].Number == "3" && orders[1].Number == "2" && orders[2].Number == "1") {
			t.Errorf("orders not sorted by CreatedAt desc: got %v", []string{orders[0].Number, orders[1].Number, orders[2].Number})
		}
	})
}
