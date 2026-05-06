package handler

import (
	"loyalty-ledger/internal/config"
	"loyalty-ledger/pkg/auth"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

// Структура для всех handler-интерфейсов
type Handlers struct {
	User    UserRoutes
	Order   OrderRoutes
	Balance BalanceRoutes
}

func NewRouter(h Handlers, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Compress(5))

	// Публичные маршруты
	// Эндпоинты пользователей
	r.Post("/api/user/register", h.User.Register)
	r.Post("/api/user/login", h.User.Login)

	// Защищенные маршруты
	jwtCfg := auth.DefaultJWTConfigWithSecret(cfg.JWTSecret)

	r.Group(func(protected chi.Router) {
		protected.Use(auth.AuthMiddleware(jwtCfg))
		// Эндпоинты заказов
		protected.Post("/api/user/orders", h.Order.RegisterOrder)
		protected.Get("/api/user/orders", h.Order.GetOrders)
		// Эндпоинты баланса
		protected.Get("/api/user/balance", h.Balance.GetBalance)
		protected.Post("/api/user/balance/withdraw", h.Balance.Withdraw)
		protected.Get("/api/user/withdrawals", h.Balance.GetWithdrawals)
	})

	return r
}
