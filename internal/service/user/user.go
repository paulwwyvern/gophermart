package user

import (
	"context"
	"errors"
	"github.com/paulwwyvern/gophermart/internal/model"
)

// после вызова userId внутри user изменяется на присвоенный id
type Repository interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)
}

type TokenCreator interface {
	CreateToken(userId int64) (string, error)
}

type Service struct {
	repo         Repository
	tokenCreator TokenCreator
}

func NewService(repo Repository, tokenCreator TokenCreator) *Service {
	return &Service{
		repo:         repo,
		tokenCreator: tokenCreator,
	}
}

func (s *Service) RegisterUser(ctx context.Context, login string, password string) (string, error) {
	user := &model.User{
		Login:    login,
		Password: password,
	}

	err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return "", err
	}

	token, err := s.tokenCreator.CreateToken(user.UserID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) LoginUser(ctx context.Context, login string, password string) (string, error) {
	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		return "", err
	}
	if user.Password != password {
		return "", errors.New("wrong password")
	}

	token, err := s.tokenCreator.CreateToken(user.UserID)
	if err != nil {
		return "", err
	}

	return token, nil
}
