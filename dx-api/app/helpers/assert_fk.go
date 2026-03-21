package helpers

import (
	"fmt"

	"github.com/goravel/framework/contracts/database/orm"
)

// AssertFK verifies that a record exists in the given table by ID.
// Use this in transactions to ensure foreign key references are valid.
func AssertFK(query orm.Query, table string, id string) error {
	count, err := query.Table(table).Where("id", id).Count()
	if err != nil {
		return fmt.Errorf("failed to verify %s record: %w", table, err)
	}
	if count == 0 {
		return fmt.Errorf("%s record not found: %s", table, id)
	}
	return nil
}
