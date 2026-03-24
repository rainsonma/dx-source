package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListGroupMembersFunctionExists(t *testing.T) {
	assert.NotNil(t, ListGroupMembers)
}

func TestKickMemberFunctionExists(t *testing.T) {
	assert.NotNil(t, KickMember)
}

func TestLeaveGroupFunctionExists(t *testing.T) {
	assert.NotNil(t, LeaveGroup)
}

func TestJoinByCodeFunctionExists(t *testing.T) {
	assert.NotNil(t, JoinByCode)
}
