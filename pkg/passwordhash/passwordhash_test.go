package passwordhash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("password")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestVerifyPassword(t *testing.T) {
	hash, _ := HashPassword("password")
	err := VerifyPassword(hash, "password")
	assert.NoError(t, err)
}

func TestVerifyPassword_Wrong(t *testing.T) {
	err := VerifyPassword("hash:)", "password")
	assert.Error(t, err)
}
