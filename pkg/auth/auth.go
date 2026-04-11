package auth

import (
	"context"
	"net/http"
	"strings"
)

type authInfoKey struct{}

type AuthInfo struct {
	UserID int64
	Login  string
}

func AuthMiddleware(cfg JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cfg.CookieName)
			if err != nil || strings.TrimSpace(cookie.Value) == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			claims, err := ParseJWT(cookie.Value, cfg)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			userID, ok := claims["user_id"].(float64)
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			login, _ := claims["login"].(string)
			info := &AuthInfo{UserID: int64(userID), Login: login}
			ctx := context.WithValue(r.Context(), authInfoKey{}, info)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetAuthInfo(ctx context.Context) (*AuthInfo, bool) {
	info, ok := ctx.Value(authInfoKey{}).(*AuthInfo)
	return info, ok
}
