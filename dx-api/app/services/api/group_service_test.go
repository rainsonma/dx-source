package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateGroupFunctionExists(t *testing.T) {
	assert.NotNil(t, CreateGroup)
}

func TestListGroupsFunctionExists(t *testing.T) {
	assert.NotNil(t, ListGroups)
}

func TestGetGroupDetailFunctionExists(t *testing.T) {
	assert.NotNil(t, GetGroupDetail)
}

func TestUpdateGroupFunctionExists(t *testing.T) {
	assert.NotNil(t, UpdateGroup)
}

func TestDismissGroupFunctionExists(t *testing.T) {
	assert.NotNil(t, DismissGroup)
}
