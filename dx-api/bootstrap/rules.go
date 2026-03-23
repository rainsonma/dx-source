package bootstrap

import (
	contractsvalidation "github.com/goravel/framework/contracts/validation"

	"dx-api/app/rules"
)

func Rules() []contractsvalidation.Rule {
	return []contractsvalidation.Rule{
		&rules.StrongPassword{},
	}
}
