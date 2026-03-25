package migrations

import "github.com/goravel/framework/facades"

type M20260325000001AddGroupPlayColumns struct{}

func (r *M20260325000001AddGroupPlayColumns) Signature() string {
	return "20260325000001_add_group_play_columns"
}

func (r *M20260325000001AddGroupPlayColumns) Up() error {
	if _, err := facades.Orm().Query().Exec(
		`ALTER TABLE game_groups
		 ADD COLUMN IF NOT EXISTS answer_time_limit INTEGER NOT NULL DEFAULT 10,
		 ADD COLUMN IF NOT EXISTS is_playing BOOLEAN NOT NULL DEFAULT false`); err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Exec(
		`ALTER TABLE game_group_members
		 ADD COLUMN IF NOT EXISTS last_won_at TIMESTAMP`); err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Exec(
		`ALTER TABLE game_subgroups
		 ADD COLUMN IF NOT EXISTS last_won_at TIMESTAMP`); err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Exec(
		`ALTER TABLE game_session_totals
		 ADD COLUMN IF NOT EXISTS game_group_id UUID REFERENCES game_groups(id) ON DELETE SET NULL,
		 ADD COLUMN IF NOT EXISTS game_subgroup_id UUID REFERENCES game_subgroups(id) ON DELETE SET NULL`); err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Exec(
		`ALTER TABLE game_session_levels
		 ADD COLUMN IF NOT EXISTS game_group_id UUID REFERENCES game_groups(id) ON DELETE SET NULL,
		 ADD COLUMN IF NOT EXISTS game_subgroup_id UUID REFERENCES game_subgroups(id) ON DELETE SET NULL`); err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Exec(
		`CREATE INDEX IF NOT EXISTS idx_game_session_totals_group
		 ON game_session_totals (game_group_id)
		 WHERE game_group_id IS NOT NULL`); err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Exec(
		`CREATE INDEX IF NOT EXISTS idx_game_session_levels_group_level
		 ON game_session_levels (game_group_id, game_level_id)
		 WHERE game_group_id IS NOT NULL`); err != nil {
		return err
	}

	return nil
}

func (r *M20260325000001AddGroupPlayColumns) Down() error {
	facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_session_levels_group_level`)
	facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_session_totals_group`)
	facades.Orm().Query().Exec(`ALTER TABLE game_session_levels DROP COLUMN IF EXISTS game_subgroup_id, DROP COLUMN IF EXISTS game_group_id`)
	facades.Orm().Query().Exec(`ALTER TABLE game_session_totals DROP COLUMN IF EXISTS game_subgroup_id, DROP COLUMN IF EXISTS game_group_id`)
	facades.Orm().Query().Exec(`ALTER TABLE game_subgroups DROP COLUMN IF EXISTS last_won_at`)
	facades.Orm().Query().Exec(`ALTER TABLE game_group_members DROP COLUMN IF EXISTS last_won_at`)
	facades.Orm().Query().Exec(`ALTER TABLE game_groups DROP COLUMN IF EXISTS is_playing, DROP COLUMN IF EXISTS answer_time_limit`)
	return nil
}
