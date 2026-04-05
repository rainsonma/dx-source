package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestForceEndGroupLosersFunctionExists(t *testing.T) {
	var fn func(string, string, string) error = ForceEndGroupLosers
	assert.NotNil(t, fn)
}

func TestForceEndGroupLosersExceptTeamFunctionExists(t *testing.T) {
	var fn func(string, string, string) error = ForceEndGroupLosersExceptTeam
	assert.NotNil(t, fn)
}
