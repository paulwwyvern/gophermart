package service

import (
	"context"

	"github.com/paulwwyvern/gophermart/internal/model"
	"github.com/paulwwyvern/gophermart/internal/model/errs"
	"github.com/paulwwyvern/gophermart/pkg/passwordhash"
)

// после вызова userId внутри user изменяется на присвоенный id
type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)
}

type TokenCreator interface {
	CreateToken(userId int64) (string, error)
}

type UserService struct {
	userRepo     UserRepository
	tokenCreator TokenCreator
}

func NewUserService(userRepo UserRepository, tokenCreator TokenCreator) *UserService {
	return &UserService{
		userRepo:     userRepo,
		tokenCreator: tokenCreator,
	}
}

func (s *UserService) RegisterUser(ctx context.Context, login string, password string) (string, error) {
	passwordHash, err := passwordhash.HashPassword(password)
	if err != nil {
		return "", err
	}
	user := &model.User{
		Login:    login,
		Password: passwordHash,
	}
	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return "", err
	}

	token, err := s.tokenCreator.CreateToken(user.UserID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserService) LoginUser(ctx context.Context, login string, password string) (string, error) {
	user, err := s.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return "", err
	}

	err = passwordhash.VerifyPassword(user.Password, password)
	if err != nil {
		return "", errs.ErrAuthFailed
	}

	token, err := s.tokenCreator.CreateToken(user.UserID)
	if err != nil {
		return "", err
	}

	return token, nil
}
