package accrual

import (
	"context"
	"log"
	"time"
)

// Переводит все NEW заказы в PROCESSING
func StartNewToProcessingWorker(ctx context.Context, accrualService OrderAccrualService, interval time.Duration) {
	const (
		maxPerTick = 100 // лимит заказов за тик
		maxJitter  = 500 * time.Millisecond
	)
	if interval == 0 {
		interval = 5 * time.Second
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Добавляем задержку
				jitter := time.Duration(int64(maxJitter) * int64(time.Now().UnixNano()%1000) / 1000)
				select {
				case <-ctx.Done():
					return
				case <-time.After(jitter):
				}
				// Получаем заказы со статусом NEW
				newOrders, err := accrualService.GetNewOrders(ctx)
				if err != nil {
					log.Printf("accrual new->processing worker: get new orders: %v", err)
					continue
				}
				if len(newOrders) == 0 {
					continue
				}
				// Устанавливаем лимит на количество обрабатываемых заказов
				if len(newOrders) > maxPerTick {
					newOrders = newOrders[:maxPerTick]
				}
				// Переводим заказы в PROCESSING
				updated, err := accrualService.BatchSetOrdersProcessing(ctx, newOrders)
				if err != nil {
					log.Printf("accrual new->processing worker: batch update error: %v", err)
					continue
				}
				if len(updated) == 0 {
					continue
				}
				// Логируем обработанные заказы
				for _, order := range updated {
					log.Printf("accrual new->processing worker: set PROCESSING for order %s", order)
				}
			}
		}
	}()
}
