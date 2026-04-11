package handler

import (
	"fmt"
	"loyalty-ledger/internal/config"
	"loyalty-ledger/pkg/auth"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type UserRoutes interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
}

// Структура для всех handler-интерфейсов
type Handlers struct {
	User  UserRoutes
	Order OrderRoutes
}

func NewRouter(h Handlers, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Compress(5))

	// Публичные маршруты
	r.Post("/api/user/register", h.User.Register)
	r.Post("/api/user/login", h.User.Login)

	// Защищенные маршруты
	jwtCfg := auth.DefaultJWTConfigWithSecret(cfg.JWTSecret)

	r.Group(func(protected chi.Router) {
		protected.Use(auth.AuthMiddleware(jwtCfg))
		// Эндпоинты заказов
		protected.Post("/api/user/orders", h.Order.RegisterOrder)
		protected.Get("/api/user/orders", h.Order.GetOrders)
		// Пример защищённого эндпоинта
		protected.Get("/api/protected", func(w http.ResponseWriter, r *http.Request) {
			info, ok := auth.GetAuthInfo(r.Context())
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello, " + info.Login + " (id=" + fmt.Sprint(info.UserID) + ")"))
		})
	})

	return r
}
