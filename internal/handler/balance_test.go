package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"loyalty-ledger/internal/model"
	"loyalty-ledger/internal/repository"
	"loyalty-ledger/internal/service"
	"loyalty-ledger/pkg/auth"
	"loyalty-ledger/pkg/points"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// svcWithErr — мок-сервис для теста 500 Internal Server Error
type svcWithErr struct{}

func (s *svcWithErr) GetBalanceWithWithdrawn(ctx context.Context, userID int64) (*model.BalanceWithWithdrawn, error) {
	return nil, assert.AnError
}
func (s *svcWithErr) Withdraw(ctx context.Context, userID int64, order string, sum int64) error {
	return assert.AnError
}
func (s *svcWithErr) GetWithdrawals(ctx context.Context, userID int64) ([]*model.Withdrawal, error) {
	return nil, assert.AnError
}

func TestBalanceHandler_GetBalance(t *testing.T) {
	repo := repository.NewInMemoryBalanceRepository()
	svc := service.NewBalanceService(repo)
	h := NewBalanceHandler(svc)
	userID := int64(42)

	t.Run("unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
		rw := httptest.NewRecorder()
		h.GetBalance(rw, req)
		if rw.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rw.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		// Accrue 1000.50, withdraw 200.25 and 300.25, expect current = 500.00, withdrawn = 500.50
		_ = repo.Accrue(context.Background(), userID, points.ToInternal(1000.50))
		_ = repo.Withdraw(context.Background(), userID, "79927398713", points.ToInternal(200.25))
		_ = repo.Withdraw(context.Background(), userID, "12345678903", points.ToInternal(300.25))
		req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
		ctx := auth.SetAuthInfo(req.Context(), &auth.AuthInfo{UserID: userID, Login: "testuser"})
		req = req.WithContext(ctx)
		rw := httptest.NewRecorder()
		h.GetBalance(rw, req)
		if rw.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rw.Code)
		}
		var resp BalanceResponse
		if err := json.NewDecoder(rw.Body).Decode(&resp); err != nil {
			t.Errorf("decode error: %v", err)
		}
		if resp.Current != 500.00 {
			t.Errorf("expected current 500.00, got %v", resp.Current)
		}
		if resp.Withdrawn != 500.50 {
			t.Errorf("expected withdrawn 500.50, got %v", resp.Withdrawn)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		hErr := NewBalanceHandler(&svcWithErr{}) // Change to use the interface
		req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
		ctx := auth.SetAuthInfo(req.Context(), &auth.AuthInfo{UserID: userID, Login: "testuser"})
		req = req.WithContext(ctx)
		rw := httptest.NewRecorder()
		hErr.GetBalance(rw, req)
		if rw.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", rw.Code)
		}
	})
}

func TestBalanceHandler_Withdraw(t *testing.T) {
	repo := repository.NewInMemoryBalanceRepository()
	svc := service.NewBalanceService(repo)
	h := NewBalanceHandler(svc)
	userID := int64(42)
	_ = repo.Accrue(context.Background(), userID, points.ToInternal(1000.00))

	t.Run("unauthorized", func(t *testing.T) {
		body := bytes.NewBufferString(`{"order":"79927398713","sum":100}`)
		req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", body)
		req.Header.Set("Content-Type", "application/json")
		rw := httptest.NewRecorder()
		h.Withdraw(rw, req)
		if rw.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rw.Code)
		}
	})

	t.Run("insufficient funds", func(t *testing.T) {
		body := bytes.NewBufferString(`{"order":"79927398713","sum":2000}`)
		req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", body)
		req.Header.Set("Content-Type", "application/json")
		ctx := auth.SetAuthInfo(req.Context(), &auth.AuthInfo{UserID: userID, Login: "testuser"})
		req = req.WithContext(ctx)
		rw := httptest.NewRecorder()
		h.Withdraw(rw, req)
		if rw.Code != http.StatusPaymentRequired {
			t.Errorf("expected 402, got %d", rw.Code)
		}
	})

	t.Run("success_float", func(t *testing.T) {
		body := bytes.NewBufferString(`{"order":"4242424242424242","sum":500.25}`)
		req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", body)
		req.Header.Set("Content-Type", "application/json")
		ctx := auth.SetAuthInfo(req.Context(), &auth.AuthInfo{UserID: userID, Login: "testuser"})
		req = req.WithContext(ctx)
		rw := httptest.NewRecorder()
		h.Withdraw(rw, req)
		if rw.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rw.Code)
		}
		bal, _ := repo.GetBalance(context.Background(), userID)
		expected := points.ToInternal(1000.00 - 500.25)
		if bal.Current != expected {
			t.Errorf("expected balance %.2f, got %d", 1000.00-500.25, bal.Current)
		}
	})

	t.Run("success_int", func(t *testing.T) {
		body := bytes.NewBufferString(`{"order":"79927398713","sum":100}`)
		req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", body)
		req.Header.Set("Content-Type", "application/json")
		ctx := auth.SetAuthInfo(req.Context(), &auth.AuthInfo{UserID: userID, Login: "testuser"})
		req = req.WithContext(ctx)
		rw := httptest.NewRecorder()
		h.Withdraw(rw, req)
		if rw.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rw.Code)
		}
		bal, _ := repo.GetBalance(context.Background(), userID)
		expected := points.ToInternal(1000.00 - 500.25 - 100)
		if bal.Current != expected {
			t.Errorf("expected balance %.2f, got %d", 1000.00-500.25-100, bal.Current)
		}
	})

	t.Run("duplicate withdrawal", func(t *testing.T) {
		body := bytes.NewBufferString(`{"order":"12345678903","sum":100}`)
		req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", body)
		req.Header.Set("Content-Type", "application/json")
		ctx := auth.SetAuthInfo(req.Context(), &auth.AuthInfo{UserID: userID, Login: "testuser"})
		req = req.WithContext(ctx)
		rw := httptest.NewRecorder()
		h.Withdraw(rw, req)
		if rw.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rw.Code)
		}
		// Второй раз по тому же orderNumber — ошибка
		body2 := bytes.NewBufferString(`{"order":"12345678903","sum":100}`)
		req2 := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", body2)
		req2.Header.Set("Content-Type", "application/json")
		req2 = req2.WithContext(ctx)
		rw2 := httptest.NewRecorder()
		h.Withdraw(rw2, req2)
		if rw2.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected 422, got %d", rw2.Code)
		}
	})

	t.Run("invalid order number", func(t *testing.T) {
		body := bytes.NewBufferString(`{"order":"notaluhn","sum":100}`)
		req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", body)
		req.Header.Set("Content-Type", "application/json")
		ctx := auth.SetAuthInfo(req.Context(), &auth.AuthInfo{UserID: userID, Login: "testuser"})
		req = req.WithContext(ctx)
		rw := httptest.NewRecorder()
		h.Withdraw(rw, req)
		if rw.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected 422, got %d", rw.Code)
		}
	})

	t.Run("bad content-type", func(t *testing.T) {
		body := bytes.NewBufferString(`{"order":"79927398713","sum":100}`)
		req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", body)
		// Не устанавливаем Content-Type
		ctx := auth.SetAuthInfo(req.Context(), &auth.AuthInfo{UserID: userID, Login: "testuser"})
		req = req.WithContext(ctx)
		rw := httptest.NewRecorder()
		h.Withdraw(rw, req)
		if rw.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rw.Code)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		body := bytes.NewBufferString(`not a json`)
		req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", body)
		req.Header.Set("Content-Type", "application/json")
		ctx := auth.SetAuthInfo(req.Context(), &auth.AuthInfo{UserID: userID, Login: "testuser"})
		req = req.WithContext(ctx)
		rw := httptest.NewRecorder()
		h.Withdraw(rw, req)
		if rw.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rw.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		hErr := NewBalanceHandler(&svcWithErr{})
		body := bytes.NewBufferString(`{"order":"79927398713","sum":100}`)
		req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", body)
		req.Header.Set("Content-Type", "application/json")
		ctx := auth.SetAuthInfo(req.Context(), &auth.AuthInfo{UserID: userID, Login: "testuser"})
		req = req.WithContext(ctx)
		rw := httptest.NewRecorder()
		hErr.Withdraw(rw, req)
		if rw.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", rw.Code)
		}
	})
}

func TestBalanceHandler_GetWithdrawals(t *testing.T) {
	t.Run("internal error", func(t *testing.T) {
		hErr := NewBalanceHandler(&svcWithErr{})
		userID := int64(42)
		req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
		ctx := auth.SetAuthInfo(req.Context(), &auth.AuthInfo{UserID: userID, Login: "testuser"})
		req = req.WithContext(ctx)
		rw := httptest.NewRecorder()
		hErr.GetWithdrawals(rw, req)
		if rw.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", rw.Code)
		}
	})

	t.Run("no withdrawals", func(t *testing.T) {
		repo := repository.NewInMemoryBalanceRepository()
		svc := service.NewBalanceService(repo)
		h := NewBalanceHandler(svc)
		userID := int64(42)
		req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
		ctx := auth.SetAuthInfo(req.Context(), &auth.AuthInfo{UserID: userID, Login: "testuser"})
		req = req.WithContext(ctx)
		rw := httptest.NewRecorder()
		h.GetWithdrawals(rw, req)
		if rw.Code != http.StatusNoContent {
			t.Errorf("expected 204, got %d", rw.Code)
		}
	})
	t.Run("sorted order", func(t *testing.T) {
		repo := repository.NewInMemoryBalanceRepository()
		svc := service.NewBalanceService(repo)
		h := NewBalanceHandler(svc)
		userID := int64(42)
		_ = svc.Accrue(context.Background(), userID, 100000)
		orders := []struct {
			number string
			sum    float64
		}{
			{"79927398713", 100.25},
			{"4242424242424242", 200.50},
			{"12345678903", 300.75},
		}
		for _, o := range orders {
			_ = svc.Withdraw(context.Background(), userID, o.number, points.ToInternal(o.sum))
			time.Sleep(10 * time.Millisecond)
		}
		req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
		ctx := auth.SetAuthInfo(req.Context(), &auth.AuthInfo{UserID: userID, Login: "testuser"})
		req = req.WithContext(ctx)
		rw := httptest.NewRecorder()
		h.GetWithdrawals(rw, req)
		if rw.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rw.Code)
		}
		var list []WithdrawalResponse
		if err := json.NewDecoder(rw.Body).Decode(&list); err != nil {
			t.Errorf("decode error: %v", err)
		}
		want := []struct {
			order string
			sum   float64
		}{
			{"12345678903", 300.75},
			{"4242424242424242", 200.50},
			{"79927398713", 100.25},
		}
		for i, o := range want {
			if list[i].Order != o.order {
				t.Errorf("unexpected order at %d: got %s, want %s", i, list[i].Order, o.order)
			}
			if list[i].Sum != o.sum {
				t.Errorf("unexpected sum at %d: got %v, want %v", i, list[i].Sum, o.sum)
			}
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		repo := repository.NewInMemoryBalanceRepository()
		svc := service.NewBalanceService(repo)
		h := NewBalanceHandler(svc)
		req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
		rw := httptest.NewRecorder()
		h.GetWithdrawals(rw, req)
		if rw.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rw.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		repo := repository.NewInMemoryBalanceRepository()
		svc := service.NewBalanceService(repo)
		h := NewBalanceHandler(svc)
		userID := int64(42)
		_ = repo.Accrue(context.Background(), userID, 1000)
		_ = repo.Withdraw(context.Background(), userID, "79927398713", 200)
		_ = repo.Withdraw(context.Background(), userID, "12345678903", 300)
		req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
		ctx := auth.SetAuthInfo(req.Context(), &auth.AuthInfo{UserID: userID, Login: "testuser"})
		req = req.WithContext(ctx)
		rw := httptest.NewRecorder()
		h.GetWithdrawals(rw, req)
		if rw.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rw.Code)
		}
		var list []struct {
			Order       string     `json:"order"`
			Sum         float64    `json:"sum"`
			ProcessedAt *time.Time `json:"processed_at,omitempty"`
		}
		if err := json.NewDecoder(rw.Body).Decode(&list); err != nil {
			t.Errorf("decode error: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("expected 2 withdrawals, got %d", len(list))
		}
		if list[0].Order != "12345678903" || list[1].Order != "79927398713" {
			t.Errorf("unexpected withdrawals: %+v", list)
		}
		if list[0].Sum != 3.0 || list[1].Sum != 2.0 {
			t.Errorf("unexpected sums: %+v", list)
		}
		if list[0].ProcessedAt == nil || list[1].ProcessedAt == nil {
			t.Errorf("expected processed_at to be set")
		}
	})
}
