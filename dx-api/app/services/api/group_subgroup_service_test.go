package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSubgroupFunctionExists(t *testing.T) {
	assert.NotNil(t, CreateSubgroup)
}

func TestListSubgroupsFunctionExists(t *testing.T) {
	assert.NotNil(t, ListSubgroups)
}

func TestUpdateSubgroupFunctionExists(t *testing.T) {
	assert.NotNil(t, UpdateSubgroup)
}

func TestDeleteSubgroupFunctionExists(t *testing.T) {
	assert.NotNil(t, DeleteSubgroup)
}

func TestListSubgroupMembersFunctionExists(t *testing.T) {
	assert.NotNil(t, ListSubgroupMembers)
}

func TestAssignSubgroupMembersFunctionExists(t *testing.T) {
	assert.NotNil(t, AssignSubgroupMembers)
}

func TestRemoveSubgroupMemberFunctionExists(t *testing.T) {
	assert.NotNil(t, RemoveSubgroupMember)
}
