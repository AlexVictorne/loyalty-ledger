package accrual

import (
	"context"
	"log"
	"sync"
	"time"
)

// Обрабатывает все PROCESSING заказы через внешнюю accrual систему
func StartProcessingAccrualWorker(ctx context.Context, client Client, accrualService OrderAccrualService, interval time.Duration, numWorkers int) {
	if interval == 0 {
		interval = 5 * time.Second
	}
	if numWorkers <= 0 {
		numWorkers = 4
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Получаем заказы со статусом PROCESSING
				processingOrders, err := accrualService.GetProcessingOrders(ctx)
				if err != nil {
					log.Printf("accrual processing worker: get pending orders: %v", err)
					continue
				}
				if len(processingOrders) == 0 {
					continue
				}
				const (
					maxAttempts = 10
					minJitter   = 50 * time.Millisecond
					maxJitter   = 250 * time.Millisecond
				)
				type task struct {
					order   string
					attempt int
				}
				// Буферизованный канал: исходные задачи + все возможные ретраи
				tasks := make(chan task, len(processingOrders)*(maxAttempts+1))
				// Отслеживаем все задачи (включая ретраи), канал закрывается по завершении
				var taskWg sync.WaitGroup
				for _, order := range processingOrders {
					taskWg.Add(1)
					tasks <- task{order: order, attempt: 0}
				}
				go func() {
					taskWg.Wait()
					close(tasks)
				}()
				var wg sync.WaitGroup
				for i := 0; i < numWorkers; i++ {
					wg.Go(func() {
						for t := range tasks {
							enqueueRetry := func() {
								if t.attempt < maxAttempts {
									taskWg.Add(1)
									tasks <- task{order: t.order, attempt: t.attempt + 1}
								} else {
									log.Printf("accrual processing worker: order %s exceeded max attempts", t.order)
								}
							}
							// Добавляем задержку между запросами
							jitter := minJitter + time.Duration(int64(maxJitter-minJitter)*int64(time.Now().UnixNano()%1000)/1000)
							select {
							case <-ctx.Done():
								taskWg.Done()
								return
							case <-time.After(jitter):
							}
							// Получаем информацию о заказе из accrual системы
							info, err := client.GetOrderInfo(ctx, t.order)
							if err != nil {
								if tooMany, ok := err.(*TooManyRequestsError); ok {
									log.Printf("accrual processing worker: 429, retry after %d", tooMany.RetryAfter)
									time.Sleep(time.Duration(tooMany.RetryAfter) * time.Second)
									enqueueRetry()
								} else {
									log.Printf("accrual processing worker: get order %s: %v (retry %d)", t.order, err, t.attempt+1)
									enqueueRetry()
								}
								taskWg.Done()
								continue
							}
							// Обновляем информацию о заказе и балансе
							if info.Status == StatusProcessed || info.Status == StatusInvalid {
								internalStatus, mapErr := MapToInternalStatus(info.Status)
								if mapErr != nil {
									log.Printf("accrual processing worker: unknown status %q for order %s", info.Status, t.order)
									taskWg.Done()
									continue
								}
								err = accrualService.UpdateOrderAndBalanceFromAccrual(ctx, t.order, internalStatus, info.Accrual)
								if err != nil {
									log.Printf("accrual processing worker: update order %s: %v (retry %d)", t.order, err, t.attempt+1)
									enqueueRetry()
								} else {
									switch info.Status {
									case StatusProcessed:
										if info.Accrual != nil {
											log.Printf("accrual processing worker: order %s PROCESSED, accrual=%.2f", t.order, *info.Accrual)
										} else {
											log.Printf("accrual processing worker: order %s PROCESSED", t.order)
										}
									case StatusInvalid:
										log.Printf("accrual processing worker: order %s INVALID", t.order)
									}
								}
							} else {
								log.Printf("accrual processing worker: order %s external status is %s, not updating", t.order, info.Status)
							}
							taskWg.Done()
						}
					})
				}
				wg.Wait()
			}
		}
	}()
}
