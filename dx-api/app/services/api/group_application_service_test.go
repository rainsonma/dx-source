package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyToGroupFunctionExists(t *testing.T) {
	assert.NotNil(t, ApplyToGroup)
}

func TestCancelApplicationFunctionExists(t *testing.T) {
	assert.NotNil(t, CancelApplication)
}

func TestListApplicationsFunctionExists(t *testing.T) {
	assert.NotNil(t, ListApplications)
}

func TestHandleApplicationFunctionExists(t *testing.T) {
	assert.NotNil(t, HandleApplication)
}
