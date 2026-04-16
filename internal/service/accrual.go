package service

import (
	"context"
	"loyalty-ledger/internal/model"
	"loyalty-ledger/internal/repository"
	"loyalty-ledger/pkg/points"
)

type AccrualService struct {
	orderRepo   repository.OrderRepository
	balanceRepo repository.BalanceRepository
}

func NewAccrualService(orderRepo repository.OrderRepository, balanceRepo repository.BalanceRepository) *AccrualService {
	return &AccrualService{orderRepo: orderRepo, balanceRepo: balanceRepo}
}

func (s *AccrualService) GetNewOrders(ctx context.Context) ([]string, error) {
	return s.orderRepo.GetOrdersByStatus(ctx, model.OrderStatusNew)
}

func (s *AccrualService) GetProcessingOrders(ctx context.Context) ([]string, error) {
	return s.orderRepo.GetOrdersByStatus(ctx, model.OrderStatusProcessing)
}

func (s *AccrualService) UpdateOrderAndBalanceFromAccrual(ctx context.Context, orderNumber string, status string, accrual *float64) error {
	order, err := s.orderRepo.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		return err
	}
	if order == nil {
		return model.ErrOrderNotFound
	}
	accrualInt := order.Accrual
	if accrual != nil {
		accrualInt = points.ToInternal(*accrual)
	}
	if order.Status == status && order.Accrual == accrualInt {
		return nil
	}
	// Обновить заказ
	if err := s.orderRepo.UpdateOrderStatus(ctx, orderNumber, status, &accrualInt); err != nil {
		return err
	}
	// Если заказ PROCESSED и есть начисление - начисляем баллы на баланс
	if status == model.OrderStatusProcessed && accrual != nil && accrualInt > 0 {
		return s.balanceRepo.Accrue(ctx, order.UserID, accrualInt)
	}
	return nil
}

// BatchSetOrdersProcessing переводит заказы из NEW в PROCESSING батчем, возвращает реально обновлённые номера
func (s *AccrualService) BatchSetOrdersProcessing(ctx context.Context, orderNumbers []string) ([]string, error) {
	return s.orderRepo.BatchUpdateOrderStatus(ctx, orderNumbers, model.OrderStatusNew, model.OrderStatusProcessing)
}
