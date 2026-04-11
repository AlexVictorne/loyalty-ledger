package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"loyalty-ledger/internal/model"
	"loyalty-ledger/internal/repository"
	"loyalty-ledger/internal/service"

	"loyalty-ledger/pkg/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeUserHandlerWithRepo() (*UserHandler, *repository.InMemoryUserRepository) {
	repo := repository.NewInMemoryUserRepository()
	svc := service.NewUserService(repo)
	jwtCfg := auth.DefaultJWTConfig()
	return NewUserHandler(svc, jwtCfg), repo
}

func TestUserHandler_Register(t *testing.T) {
	h, _ := makeUserHandlerWithRepo()
	ts := httptest.NewServer(http.HandlerFunc(h.Register))
	defer ts.Close()

	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
	}{
		{"success", model.UserRequest{Login: "user1", Password: "pass1"}, http.StatusOK},
		{"duplicate", model.UserRequest{Login: "user1", Password: "pass1"}, http.StatusConflict},
		{"empty login", model.UserRequest{Login: "", Password: "pass1"}, http.StatusBadRequest},
		{"empty password", model.UserRequest{Login: "user2", Password: ""}, http.StatusBadRequest},
		{"empty", model.UserRequest{}, http.StatusBadRequest},
		{"nil", nil, http.StatusBadRequest},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != nil {
				b, _ := json.Marshal(tc.body)
				body = bytes.NewReader(b)
			}
			resp, err := http.Post(ts.URL, "application/json", body)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tc.wantStatus, resp.StatusCode, "case %d", i)

			// Проверка cookie и токена только для успешного кейса
			if tc.wantStatus == http.StatusOK {
				cookies := resp.Cookies()
				assert.NotEmpty(t, cookies)
				var jwtCookie *http.Cookie
				for _, c := range cookies {
					if c.Name == h.jwtCfg.CookieName {
						jwtCookie = c
					}
				}
				if assert.NotNil(t, jwtCookie, "JWT cookie must be set") {
					claims, err := auth.ParseJWT(jwtCookie.Value, h.jwtCfg)
					assert.NoError(t, err)
					reqObj, _ := tc.body.(model.UserRequest)
					assert.Equal(t, reqObj.Login, claims["login"])
				}
			} else {
				// Для неуспешных кейсов cookie не должно быть
				for _, c := range resp.Cookies() {
					assert.NotEqual(t, h.jwtCfg.CookieName, c.Name)
				}
			}
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	h, _ := makeUserHandlerWithRepo()
	mux := http.NewServeMux()
	mux.HandleFunc("/register", h.Register)
	mux.HandleFunc("/login", h.Login)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	regBody, _ := json.Marshal(model.UserRequest{Login: "user1", Password: "pass1"})
	resp, err := http.Post(ts.URL+"/register", "application/json", bytes.NewReader(regBody))
	require.NoError(t, err)
	resp.Body.Close()

	tests := []struct {
		name       string
		body       any
		wantStatus int
	}{
		{"success", model.UserRequest{Login: "user1", Password: "pass1"}, http.StatusOK},
		{"wrong password", model.UserRequest{Login: "user1", Password: "wrong"}, http.StatusUnauthorized},
		{"not found", model.UserRequest{Login: "nouser", Password: "pass"}, http.StatusUnauthorized},
		{"empty login", model.UserRequest{Login: "", Password: "pass1"}, http.StatusBadRequest},
		{"empty password", model.UserRequest{Login: "user1", Password: ""}, http.StatusBadRequest},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != nil {
				b, _ := json.Marshal(tc.body)
				body = bytes.NewReader(b)
			}
			resp, err := http.Post(ts.URL+"/login", "application/json", body)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tc.wantStatus, resp.StatusCode, "case %d", i)

			if tc.wantStatus == http.StatusOK {
				cookies := resp.Cookies()
				assert.NotEmpty(t, cookies)
				var jwtCookie *http.Cookie
				for _, c := range cookies {
					if c.Name == h.jwtCfg.CookieName {
						jwtCookie = c
					}
				}
				if assert.NotNil(t, jwtCookie, "JWT cookie must be set") {
					claims, err := auth.ParseJWT(jwtCookie.Value, h.jwtCfg)
					assert.NoError(t, err)
					reqObj, _ := tc.body.(model.UserRequest)
					assert.Equal(t, reqObj.Login, claims["login"])
				}
			} else {
				for _, c := range resp.Cookies() {
					assert.NotEqual(t, h.jwtCfg.CookieName, c.Name)
				}
			}
		})
	}
}
