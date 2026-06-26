package service

import (
	"context"
	"testing"

	"github.com/paulwwyvern/gophermart/internal/model"
	"github.com/paulwwyvern/gophermart/pkg/passwordhash"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -source=user.go -destination=mock_user_test.go -package=service

func TestUserService_RegisterUser(t *testing.T) {
	userId := int64(123456)
	login := "abcdef"
	password := "qwerty"

	token := "token"

	ctrl := gomock.NewController(t)
	tokenCreator := NewMockTokenCreator(ctrl)
	userRepo := NewMockUserRepository(ctrl)
	gomock.InOrder(
		userRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(userId, nil),
		tokenCreator.EXPECT().CreateToken(userId).Return(token, nil),
	)

	service := NewUserService(userRepo, tokenCreator)
	res, err := service.RegisterUser(context.Background(), login, password)
	assert.NoError(t, err)
	assert.Equal(t, token, res)
}

func TestUserService_LoginUser(t *testing.T) {
	userId := int64(123456)
	login := "abcdef"
	password := "qwerty"

	token := "token"

	passwordHash, _ := passwordhash.HashPassword(password)
	user := &model.User{
		UserID:   userId,
		Login:    login,
		Password: passwordHash,
	}

	ctrl := gomock.NewController(t)
	tokenCreator := NewMockTokenCreator(ctrl)
	userRepo := NewMockUserRepository(ctrl)
	gomock.InOrder(
		userRepo.EXPECT().GetUserByLogin(gomock.Any(), login).Return(user, nil),
		tokenCreator.EXPECT().CreateToken(userId).Return(token, nil),
	)

	service := NewUserService(userRepo, tokenCreator)
	res, err := service.LoginUser(context.Background(), login, password)
	assert.NoError(t, err)
	assert.Equal(t, token, res)
}
