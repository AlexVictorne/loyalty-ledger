package auth

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct {
	Secret     string
	CookieName string
	TTL        time.Duration
}

func DefaultJWTConfig() JWTConfig {
	return JWTConfig{
		Secret:     "dev_secret",
		CookieName: "auth",
		TTL:        24 * time.Hour,
	}
}

func DefaultJWTConfigWithSecret(secret string) JWTConfig {
	cfg := DefaultJWTConfig()
	cfg.Secret = secret
	return cfg
}

func GenerateJWT(userID int64, login string, cfg JWTConfig) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"login":   login,
		"exp":     time.Now().Add(cfg.TTL).Unix(),
	})
	return token.SignedString([]byte(cfg.Secret))
}

func ParseJWT(tokenString string, cfg JWTConfig) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}

func SetJWTCookie(w http.ResponseWriter, token string, cfg JWTConfig) {
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(cfg.TTL.Seconds()),
	})
}

// Сгенерировать JWT и установить его в cookie
func IssueJWTAndSetCookie(w http.ResponseWriter, userID int64, login string, cfg JWTConfig) error {
	token, err := GenerateJWT(userID, login, cfg)
	if err != nil {
		return err
	}
	SetJWTCookie(w, token, cfg)
	return nil
}
