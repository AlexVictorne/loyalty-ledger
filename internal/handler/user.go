package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"loyalty-ledger/internal/model"
	"loyalty-ledger/internal/service"
	"loyalty-ledger/pkg/auth"
	"net/http"
	"strings"
)

type UserHandler struct {
	service *service.UserService
	jwtCfg  auth.JWTConfig
}

func NewUserHandler(service *service.UserService, jwtCfg auth.JWTConfig) *UserHandler {
	return &UserHandler{service: service, jwtCfg: jwtCfg}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req model.UserRequest
	body, err := io.ReadAll(r.Body)
	if err != nil || json.Unmarshal(body, &req) != nil || strings.TrimSpace(req.Login) == "" || strings.TrimSpace(req.Password) == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, err := h.service.Register(context.Background(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserExists):
			w.WriteHeader(http.StatusConflict)
		case errors.Is(err, service.ErrLoginPasswordEmpty):
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if err := auth.IssueJWTAndSetCookie(w, user.ID, user.Login, h.jwtCfg); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req model.UserRequest
	body, err := io.ReadAll(r.Body)
	if err != nil || json.Unmarshal(body, &req) != nil || strings.TrimSpace(req.Login) == "" || strings.TrimSpace(req.Password) == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, err := h.service.Authenticate(context.Background(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLoginPasswordEmpty):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, service.ErrUserNotFound), errors.Is(err, service.ErrInvalidPassword):
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if err := auth.IssueJWTAndSetCookie(w, user.ID, user.Login, h.jwtCfg); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
