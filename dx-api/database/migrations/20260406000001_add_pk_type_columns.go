package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260406000001AddPkTypeColumns struct{}

func (r *M20260406000001AddPkTypeColumns) Signature() string {
	return "20260406000001_add_pk_type_columns"
}

func (r *M20260406000001AddPkTypeColumns) Up() error {
	_, err := facades.Orm().Query().Exec(
		`ALTER TABLE game_pks
		 ADD COLUMN pk_type text NOT NULL DEFAULT 'random',
		 ADD COLUMN invitation_status text`)
	return err
}

func (r *M20260406000001AddPkTypeColumns) Down() error {
	_, err := facades.Orm().Query().Exec(
		`ALTER TABLE game_pks
		 DROP COLUMN IF EXISTS pk_type,
		 DROP COLUMN IF EXISTS invitation_status`)
	return err
}
