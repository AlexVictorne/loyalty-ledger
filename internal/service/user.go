package service

import (
	"context"
	"errors"

	"loyalty-ledger/internal/model"
	"loyalty-ledger/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = model.ErrUserExists
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrLoginPasswordEmpty = errors.New("login and password required")
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(ctx context.Context, req *model.UserRequest) (*model.User, error) {
	if req.Login == "" || req.Password == "" {
		return nil, ErrLoginPasswordEmpty
	}
	user, err := s.repo.GetByLogin(ctx, req.Login)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return nil, ErrUserExists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	newUser := &model.User{
		Login:    req.Login,
		Password: string(hash),
	}
	if err := s.repo.CreateUser(ctx, newUser); err != nil {
		return nil, err
	}
	return newUser, nil
}

func (s *UserService) Authenticate(ctx context.Context, req *model.UserRequest) (*model.User, error) {
	if req.Login == "" || req.Password == "" {
		return nil, ErrLoginPasswordEmpty
	}
	user, err := s.repo.GetByLogin(ctx, req.Login)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		return nil, ErrInvalidPassword
	}
	return user, nil
}
