package tests

import (
	"github.com/goravel/framework/testing"

	"dx-api/bootstrap"
)

func init() {
	bootstrap.Boot()
}

type TestCase struct {
	testing.TestCase
}
