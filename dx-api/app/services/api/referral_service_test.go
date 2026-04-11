package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateInviteCodeEmptyReturnsFalse(t *testing.T) {
	ok, err := ValidateInviteCode("")
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestValidateInviteCodeFunctionExists(t *testing.T) {
	assert.NotNil(t, ValidateInviteCode)
}
