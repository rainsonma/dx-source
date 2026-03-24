package migrations

import "github.com/goravel/framework/facades"

type M20260324000001AlterGroupTables struct{}

func (r *M20260324000001AlterGroupTables) Signature() string {
	return "20260324000001_alter_group_tables"
}

func (r *M20260324000001AlterGroupTables) Up() error {
	if facades.Schema().HasColumn("game_group_members", "role") {
		if _, err := facades.Orm().Query().Exec(`ALTER TABLE game_group_members DROP COLUMN role`); err != nil {
			return err
		}
	}
	if facades.Schema().HasColumn("game_subgroup_members", "role") {
		if _, err := facades.Orm().Query().Exec(`ALTER TABLE game_subgroup_members DROP COLUMN role`); err != nil {
			return err
		}
	}
	if !facades.Schema().HasColumn("game_groups", "member_count") {
		if _, err := facades.Orm().Query().Exec(`ALTER TABLE game_groups ADD COLUMN member_count integer NOT NULL DEFAULT 0`); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260324000001AlterGroupTables) Down() error {
	_, _ = facades.Orm().Query().Exec(`ALTER TABLE game_group_members ADD COLUMN role text NOT NULL DEFAULT ''`)
	_, _ = facades.Orm().Query().Exec(`ALTER TABLE game_subgroup_members ADD COLUMN role text NOT NULL DEFAULT ''`)
	_, _ = facades.Orm().Query().Exec(`ALTER TABLE game_groups DROP COLUMN IF EXISTS member_count`)
	return nil
}
