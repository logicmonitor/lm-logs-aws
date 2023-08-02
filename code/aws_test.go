package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReturnEmptySecretValueOnEmptyInput(t *testing.T) {
	secret := getSecretValue("")
	assert.Equal(t, secret, "")
}
