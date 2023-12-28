package main

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReturnEmptySecretValueOnEmptyInput(t *testing.T) {
	secret := getSecretValue("")
	assert.Equal(t, secret, "")
}

func TestAccessSecret(t *testing.T) {
	t.Setenv("LOGICMONITOR_ACCESS_KEY", "arn:aws:secretsmanager:region:key")
	t.Setenv("LOGICMONITOR_ACCESS_ID", "arn:aws:secretsmanager:region:id")
	t.Setenv("USE_SECRET_MANAGER", "true")

	useSecretManager := os.Getenv("USE_SECRET_MANAGER")
	assert.Equal(t, "true", useSecretManager)
	assert.True(t, strings.Contains(os.Getenv("LOGICMONITOR_ACCESS_KEY"), "arn"))
	assert.True(t, strings.Contains(os.Getenv("LOGICMONITOR_ACCESS_ID"), "arn"))

}
