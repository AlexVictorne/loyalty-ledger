package accrual

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakeClient struct {
	calls   []string
	results map[string]OrderInfo
	errs    map[string]error
}

func (c *fakeClient) GetOrderInfo(ctx context.Context, orderNumber string) (OrderInfo, error) {
	c.calls = append(c.calls, orderNumber)
	if err, ok := c.errs[orderNumber]; ok {
		return OrderInfo{}, err
	}
	return c.results[orderNumber], nil
}

type fakeAccrualService struct {
	newOrders        []string
	processingOrders []string
	update           map[string]struct {
		status  string
		accrual *float64
	}
	mu         sync.Mutex
	calls      []string
	failUpdate map[string]error
}

func (s *fakeAccrualService) GetNewOrders(ctx context.Context) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Return a copy to avoid race
	out := make([]string, len(s.newOrders))
	copy(out, s.newOrders)
	return out, nil
}
func (s *fakeAccrualService) GetProcessingOrders(ctx context.Context) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.processingOrders))
	copy(out, s.processingOrders)
	return out, nil
}
func (s *fakeAccrualService) UpdateOrderAndBalanceFromAccrual(ctx context.Context, orderNumber string, status string, accrual *float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, orderNumber)
	if s.failUpdate != nil {
		if err, ok := s.failUpdate[orderNumber]; ok {
			return err
		}
	}
	// Simulate status transitions
	switch status {
	case "PROCESSING":
		// Remove from newOrders, add to processingOrders if not already there
		s.newOrders = removeString(s.newOrders, orderNumber)
		if !containsString(s.processingOrders, orderNumber) {
			s.processingOrders = append(s.processingOrders, orderNumber)
		}
	case "PROCESSED":
		// Remove from processingOrders
		s.processingOrders = removeString(s.processingOrders, orderNumber)
	}
	s.update[orderNumber] = struct {
		status  string
		accrual *float64
	}{status, accrual}
	return nil
}

// BatchSetOrdersProcessing переводит заказы из NEW в PROCESSING батчем, возвращает реально обновлённые номера
func (s *fakeAccrualService) BatchSetOrdersProcessing(ctx context.Context, orderNumbers []string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var updated []string
	for _, num := range orderNumbers {
		// только если заказ есть в newOrders
		found := false
		for i, n := range s.newOrders {
			if n == num {
				// удалить из newOrders
				s.newOrders = append(s.newOrders[:i], s.newOrders[i+1:]...)
				found = true
				break
			}
		}
		if found {
			// добавить в processingOrders, если ещё нет
			if !containsString(s.processingOrders, num) {
				s.processingOrders = append(s.processingOrders, num)
			}
			s.update[num] = struct {
				status  string
				accrual *float64
			}{"PROCESSING", nil}
			updated = append(updated, num)
		}
	}
	return updated, nil
}

// Helpers for slice operations
func removeString(slice []string, val string) []string {
	out := slice[:0]
	for _, s := range slice {
		if s != val {
			out = append(out, s)
		}
	}
	return out
}
func containsString(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

func TestWorker_HappyPath(t *testing.T) {
	client := &fakeClient{
		results: map[string]OrderInfo{
			"1": {Order: "1", Status: "PROCESSED", Accrual: floatPtr(5)},
		},
	}
	service := &fakeAccrualService{
		newOrders:        []string{"1"},
		processingOrders: []string{},
		update: make(map[string]struct {
			status  string
			accrual *float64
		}),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	StartNewToProcessingWorker(ctx, service, 10*time.Millisecond)
	StartProcessingAccrualWorker(ctx, client, service, 10*time.Millisecond, 2)
	// Wait up to 500ms for the order to reach PROCESSED
	deadline := time.Now().Add(500 * time.Millisecond)
	var accrual *float64
	for {
		service.mu.Lock()
		upd, ok := service.update["1"]
		processed := ok && upd.status == "PROCESSED"
		if processed {
			accrual = upd.accrual
		}
		service.mu.Unlock()
		if processed {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("timeout waiting for order to be PROCESSED")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if accrual == nil {
		t.Fatal("accrual should not be nil")
	}
	assert.Equal(t, 5.0, *accrual)
}

func TestWorker_ParallelOrders(t *testing.T) {
	client := &fakeClient{
		results: map[string]OrderInfo{
			"1": {Order: "1", Status: "PROCESSED", Accrual: floatPtr(5)},
			"2": {Order: "2", Status: "PROCESSED", Accrual: floatPtr(10)},
			"3": {Order: "3", Status: "PROCESSED", Accrual: floatPtr(15)},
		},
	}
	service := &fakeAccrualService{
		newOrders:        []string{"1", "2", "3"},
		processingOrders: []string{},
		update: make(map[string]struct {
			status  string
			accrual *float64
		}),
	}
	ctx := t.Context()
	StartNewToProcessingWorker(ctx, service, 10*time.Millisecond)
	StartProcessingAccrualWorker(ctx, client, service, 10*time.Millisecond, 3)
	// Wait up to 500ms for all orders to reach PROCESSED
	deadline := time.Now().Add(500 * time.Millisecond)
	for {
		service.mu.Lock()
		allProcessed := true
		for _, n := range []string{"1", "2", "3"} {
			upd, ok := service.update[n]
			if !ok || upd.status != "PROCESSED" {
				allProcessed = false
				break
			}
		}
		service.mu.Unlock()
		if allProcessed {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("timeout waiting for all orders to be PROCESSED")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestWorker_TooManyRequests(t *testing.T) {
	client := &fakeClient{
		errs: map[string]error{
			"2": &TooManyRequestsError{RetryAfter: 1},
		},
	}
	service := &fakeAccrualService{
		newOrders:        []string{"2"},
		processingOrders: []string{},
		update: make(map[string]struct {
			status  string
			accrual *float64
		}),
	}
	ctx := t.Context()
	StartNewToProcessingWorker(ctx, service, 10*time.Millisecond)
	StartProcessingAccrualWorker(ctx, client, service, 10*time.Millisecond, 4)
	// Wait up to 200ms for the status to be set to PROCESSING (but not PROCESSED)
	deadline := time.Now().Add(200 * time.Millisecond)
	for {
		service.mu.Lock()
		upd, ok := service.update["2"]
		isProcessing := ok && upd.status == "PROCESSING" && upd.accrual == nil
		service.mu.Unlock()
		if isProcessing {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("timeout waiting for order to be PROCESSING with nil accrual")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
