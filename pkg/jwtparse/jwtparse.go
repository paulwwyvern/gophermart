package jwtparse

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type Claims struct {
	jwt.RegisteredClaims
	UserId int64
}

type Parser struct {
	secret   []byte
	tokenTTL time.Duration
}

func NewParser(secret string, tokenTTL time.Duration) *Parser {
	return &Parser{
		secret:   []byte(secret),
		tokenTTL: tokenTTL,
	}
}

func (p *Parser) CreateToken(userId int64) (string, error) {
	now := time.Now()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(p.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	})

	tokenString, err := token.SignedString(p.secret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (p *Parser) ValidateToken(tokenString string) (int64, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return p.secret, nil
	})
	if err != nil {
		return -1, err
	}

	if !token.Valid {
		return -1, fmt.Errorf("invalid token")
	}

	return claims.UserId, nil
}
