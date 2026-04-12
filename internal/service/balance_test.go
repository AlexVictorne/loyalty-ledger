package service

import (
	"context"
	"errors"
	"loyalty-ledger/internal/repository"
	"testing"
	"time"
)

func TestBalanceService_Accrue(t *testing.T) {
	repo := repository.NewInMemoryBalanceRepository()
	svc := NewBalanceService(repo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("negative sum", func(t *testing.T) {
		err := svc.Accrue(ctx, userID, -100)
		if !errors.Is(err, ErrInvalidSum) {
			t.Errorf("expected ErrInvalidSum, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		err := svc.Accrue(ctx, userID, 1000)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		bal, _ := repo.GetBalance(ctx, userID)
		if bal.Current != 1000 {
			t.Errorf("expected 1000, got %d", bal.Current)
		}
	})
}

func TestBalanceService_Withdraw(t *testing.T) {
	repo := repository.NewInMemoryBalanceRepository()
	svc := NewBalanceService(repo)
	ctx := context.Background()
	userID := int64(1)
	_ = repo.Accrue(ctx, userID, 1000)

	t.Run("order not digits", func(t *testing.T) {
		err := svc.Withdraw(ctx, userID, "abc123", 100)
		if !errors.Is(err, ErrOrderNumberRequired) {
			t.Errorf("expected ErrOrderNumberRequired, got %v", err)
		}
	})

	t.Run("order not luhn", func(t *testing.T) {
		err := svc.Withdraw(ctx, userID, "1234567890", 100)
		if !errors.Is(err, ErrOrderNumberRequired) {
			t.Errorf("expected ErrOrderNumberRequired, got %v", err)
		}
	})

	t.Run("negative sum", func(t *testing.T) {
		err := svc.Withdraw(ctx, userID, "123", -100)
		if !errors.Is(err, ErrInvalidSum) {
			t.Errorf("expected ErrInvalidSum, got %v", err)
		}
	})

	t.Run("empty order", func(t *testing.T) {
		err := svc.Withdraw(ctx, userID, "", 100)
		if !errors.Is(err, ErrOrderNumberRequired) {
			t.Errorf("expected ErrOrderNumberRequired, got %v", err)
		}
	})

	t.Run("insufficient funds", func(t *testing.T) {
		err := svc.Withdraw(ctx, userID, "79927398713", 2000)
		if !errors.Is(err, ErrInsufficientFunds) {
			t.Errorf("expected ErrInsufficientFunds, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		err := svc.Withdraw(ctx, userID, "4242424242424242", 500)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		bal, _ := repo.GetBalance(ctx, userID)
		if bal.Current != 500 {
			t.Errorf("expected balance 500, got %d", bal.Current)
		}
	})

	t.Run("accrue and withdraw sequence", func(t *testing.T) {
		repo2 := repository.NewInMemoryBalanceRepository()
		svc2 := NewBalanceService(repo2)
		ctx2 := context.Background()
		userID2 := int64(2)
		// Начисляем 1000
		_ = svc2.Accrue(ctx2, userID2, 1000)
		// Списываем 200
		_ = svc2.Withdraw(ctx2, userID2, "79927398713", 200)
		// Списываем 300
		_ = svc2.Withdraw(ctx2, userID2, "12345678903", 300)
		// Начисляем ещё 500
		_ = svc2.Accrue(ctx2, userID2, 500)
		// Списываем 400
		_ = svc2.Withdraw(ctx2, userID2, "4242424242424242", 400)
		bal, _ := repo2.GetBalance(ctx2, userID2)
		if bal.Current != 600 {
			t.Errorf("expected 600, got %d", bal.Current)
		}
	})

	t.Run("duplicate withdrawal", func(t *testing.T) {
		// Первый раз успешно списываем
		err := svc.Withdraw(ctx, userID, "12345678903", 100)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		// Второй раз по тому же orderNumber — ошибка
		err = svc.Withdraw(ctx, userID, "12345678903", 100)
		if !errors.Is(err, ErrOrderAlreadyWithdrawn) {
			t.Errorf("expected ErrOrderAlreadyWithdrawn, got %v", err)
		}
	})
}

func TestBalanceService_GetWithdrawals(t *testing.T) {
	repo := repository.NewInMemoryBalanceRepository()
	svc := NewBalanceService(repo)
	ctx := context.Background()
	userID := int64(10)

	t.Run("empty list", func(t *testing.T) {
		list, err := svc.GetWithdrawals(ctx, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 0 {
			t.Errorf("expected 0 withdrawals, got %d", len(list))
		}
	})

	t.Run("several withdrawals", func(t *testing.T) {
		_ = svc.Accrue(ctx, userID, 1000)
		// Используем валидные Luhn-номера
		orders := []struct {
			number string
			sum    int64
		}{
			{"79927398713", 200},
			{"12345678903", 300},
			{"4242424242424242", 100},
		}
		for _, o := range orders {
			err := svc.Withdraw(ctx, userID, o.number, o.sum)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}
		list, err := svc.GetWithdrawals(ctx, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 3 {
			t.Errorf("expected 3 withdrawals, got %d", len(list))
		}
		// Ожидаемый порядок: последний — самый новый
		want := []struct {
			number string
			sum    int64
		}{
			{"4242424242424242", 100},
			{"12345678903", 300},
			{"79927398713", 200},
		}
		for i, o := range want {
			if list[i].OrderNumber != o.number {
				t.Errorf("unexpected order number at %d: got %s, want %s", i, list[i].OrderNumber, o.number)
			}
			if list[i].Sum != o.sum {
				t.Errorf("unexpected sum at %d: got %d, want %d", i, list[i].Sum, o.sum)
			}
			if list[i].ProcessedAt == nil {
				t.Errorf("ProcessedAt is nil for withdrawal %+v", list[i])
			}
		}
	})
}

func TestBalanceService_GetBalanceWithWithdrawn(t *testing.T) {
	repo := repository.NewInMemoryBalanceRepository()
	svc := NewBalanceService(repo)
	ctx := context.Background()
	userID := int64(100)
	// Нет операций
	bw, err := svc.GetBalanceWithWithdrawn(ctx, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bw.Current != 0 || bw.Withdrawn != 0 {
		t.Errorf("expected 0/0, got %d/%d", bw.Current, bw.Withdrawn)
	}
	// Начисление и списания
	_ = svc.Accrue(ctx, userID, 1000)
	_ = svc.Withdraw(ctx, userID, "79927398713", 200)
	_ = svc.Withdraw(ctx, userID, "12345678903", 300)
	bw, err = svc.GetBalanceWithWithdrawn(ctx, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bw.Current != 500 {
		t.Errorf("expected current 500, got %d", bw.Current)
	}
	if bw.Withdrawn != 500 {
		t.Errorf("expected withdrawn 500, got %d", bw.Withdrawn)
	}
}

func TestBalanceService_GetWithdrawals_Sorting(t *testing.T) {
	svc := NewBalanceService(repository.NewInMemoryBalanceRepository())
	ctx := context.Background()
	userID := int64(200)
	_ = svc.Accrue(ctx, userID, 1000)
	// Совершаем списания с небольшими паузами для разного времени
	orders := []struct {
		number string
		sum    int64
	}{
		{"79927398713", 100},
		{"4242424242424242", 200},
		{"12345678903", 300},
	}
	for _, o := range orders {
		_ = svc.Withdraw(ctx, userID, o.number, o.sum)
		time.Sleep(10 * time.Millisecond)
	}
	list, err := svc.GetWithdrawals(ctx, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 withdrawals, got %d", len(list))
	}
	// Должно быть: последний — самый новый
	want := []string{"12345678903", "4242424242424242", "79927398713"}
	for i, o := range want {
		if list[i].OrderNumber != o {
			t.Errorf("unexpected order at %d: got %s, want %s", i, list[i].OrderNumber, o)
		}
	}
}
