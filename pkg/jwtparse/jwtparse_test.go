package jwtparse

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateToken(t *testing.T) {
	parser := NewParser("abcde", 10*time.Millisecond)
	token, err := parser.CreateToken(0)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestParseToken(t *testing.T) {
	parser := NewParser("abcde", time.Second)
	token, err := parser.CreateToken(222)
	require.NoError(t, err)
	userId, err := parser.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, int64(222), userId)
}

func TestParseInvalidToken(t *testing.T) {
	parser := NewParser("abcde", 10*time.Millisecond)
	_, err := parser.ValidateToken("abcde")
	assert.Error(t, err)
}

func TestParseExpiredToken(t *testing.T) {
	parser := NewParser("abcde", 10*time.Millisecond)
	token, err := parser.CreateToken(222)
	require.NoError(t, err)
	time.Sleep(200 * time.Millisecond)
	_, err = parser.ValidateToken(token)
	assert.Error(t, err)
}
