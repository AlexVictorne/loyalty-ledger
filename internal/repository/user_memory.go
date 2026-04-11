package repository

import (
	"context"
	"loyalty-ledger/internal/model"
	"sync"
)

type InMemoryUserRepository struct {
	mu     sync.RWMutex
	users  map[string]*model.User
	nextID int64
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:  make(map[string]*model.User),
		nextID: 1,
	}
}

func (r *InMemoryUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[user.Login]; exists {
		return model.ErrUserExists
	}
	user.ID = r.nextID
	r.nextID++
	userCopy := *user
	r.users[user.Login] = &userCopy
	return nil
}

func (r *InMemoryUserRepository) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, exists := r.users[login]
	if !exists {
		return nil, nil
	}
	userCopy := *user
	return &userCopy, nil
}
