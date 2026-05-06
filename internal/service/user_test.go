package service

import (
	"context"
	"testing"

	"loyalty-ledger/internal/model"
	"loyalty-ledger/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserService_Register(t *testing.T) {
	type testCase struct {
		name      string
		req       *model.UserRequest
		before    func(svc *UserService)
		wantErr   error
		wantLogin string
	}
	cases := []testCase{
		{
			name:      "success",
			req:       &model.UserRequest{Login: "user1", Password: "pass1"},
			wantErr:   nil,
			wantLogin: "user1",
		},
		{
			name: "duplicate",
			req:  &model.UserRequest{Login: "user1", Password: "pass1"},
			before: func(svc *UserService) {
				_, _ = svc.Register(context.Background(), &model.UserRequest{Login: "user1", Password: "pass1"})
			},
			wantErr: ErrUserExists,
		},
		{
			name:    "empty login/password",
			req:     &model.UserRequest{},
			wantErr: ErrLoginPasswordEmpty,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := repository.NewInMemoryUserRepository()
			svc := NewUserService(repo)
			if tc.before != nil {
				tc.before(svc)
			}
			user, err := svc.Register(context.Background(), tc.req)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantLogin, user.Login)
			}
		})
	}
}

func TestUserService_Authenticate(t *testing.T) {
	type testCase struct {
		name    string
		req     *model.UserRequest
		before  func(svc *UserService)
		wantErr error
	}
	cases := []testCase{
		{
			name: "success",
			req:  &model.UserRequest{Login: "user1", Password: "pass1"},
			before: func(svc *UserService) {
				_, _ = svc.Register(context.Background(), &model.UserRequest{Login: "user1", Password: "pass1"})
			},
			wantErr: nil,
		},
		{
			name: "invalid password",
			req:  &model.UserRequest{Login: "user1", Password: "wrong"},
			before: func(svc *UserService) {
				_, _ = svc.Register(context.Background(), &model.UserRequest{Login: "user1", Password: "pass1"})
			},
			wantErr: ErrInvalidPassword,
		},
		{
			name:    "user not found",
			req:     &model.UserRequest{Login: "nouser", Password: "pass"},
			wantErr: ErrUserNotFound,
		},
		{
			name:    "empty login/password",
			req:     &model.UserRequest{},
			wantErr: ErrLoginPasswordEmpty,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := repository.NewInMemoryUserRepository()
			svc := NewUserService(repo)
			if tc.before != nil {
				tc.before(svc)
			}
			_, err := svc.Authenticate(context.Background(), tc.req)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
