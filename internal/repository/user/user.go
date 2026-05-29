package user

import (
	"context"
	"errors"
	"github.com/paulwwyvern/gophermart/internal/model"
)

type Storage struct {
	users map[string]string
}

func NewStorage() *Storage {
	return &Storage{
		users: make(map[string]string),
	}
}

func (s *Storage) CreateUser(ctx context.Context, user *model.User) error {
	if _, ok := s.users[user.Login]; ok {
		return errors.New("username already exists")
	}
	s.users[user.Login] = user.Password
	user.UserID = 222

	return nil
}

func (s *Storage) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	if _, ok := s.users[login]; !ok {
		return nil, errors.New("user not found")
	}
	user := &model.User{
		UserID:   222,
		Login:    login,
		Password: s.users[login],
	}

	return user, nil
}
