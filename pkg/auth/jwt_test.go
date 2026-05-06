package auth_test

import (
	"loyalty-ledger/pkg/auth"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAndParseJWT(t *testing.T) {
	cfg := auth.DefaultJWTConfigWithSecret("test_secret")
	token, err := auth.GenerateJWT(42, "testuser", cfg)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := auth.ParseJWT(token, cfg)
	assert.NoError(t, err)
	assert.Equal(t, float64(42), claims["user_id"])
	assert.Equal(t, "testuser", claims["login"])
}

func TestSetJWTCookie(t *testing.T) {
	cfg := auth.DefaultJWTConfigWithSecret("test_secret")
	rec := httptest.NewRecorder()
	auth.SetJWTCookie(rec, "sometoken", cfg)
	cookie := rec.Result().Cookies()
	assert.Len(t, cookie, 1)
	assert.Equal(t, cfg.CookieName, cookie[0].Name)
	assert.Equal(t, "sometoken", cookie[0].Value)
	assert.True(t, cookie[0].HttpOnly)
	assert.Equal(t, "/", cookie[0].Path)
	assert.InDelta(t, int(cfg.TTL.Seconds()), cookie[0].MaxAge, 1)
}

func TestIssueJWTAndSetCookie(t *testing.T) {
	cfg := auth.DefaultJWTConfigWithSecret("test_secret")
	rec := httptest.NewRecorder()
	err := auth.IssueJWTAndSetCookie(rec, 99, "jwtuser", cfg)
	assert.NoError(t, err)
	cookie := rec.Result().Cookies()
	assert.Len(t, cookie, 1)
	assert.Equal(t, cfg.CookieName, cookie[0].Name)
	assert.NotEmpty(t, cookie[0].Value)

	claims, err := auth.ParseJWT(cookie[0].Value, cfg)
	assert.NoError(t, err)
	assert.Equal(t, float64(99), claims["user_id"])
	assert.Equal(t, "jwtuser", claims["login"])
	assert.True(t, claims["exp"].(float64) > float64(time.Now().Unix()))
}
