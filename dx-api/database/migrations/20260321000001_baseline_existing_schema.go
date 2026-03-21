package migrations

import (
	"dx-api/app/facades"
)

type M20260321000001BaselineExistingSchema struct{}

// Signature The unique signature for the migration.
func (r *M20260321000001BaselineExistingSchema) Signature() string {
	return "20260321000001_baseline_existing_schema"
}

// Up is a no-op: all tables already exist in the database, created during
// the Prisma era. This migration serves as the baseline for Goravel's
// migration system. All future schema changes should be Goravel migrations.
//
// Existing tables (48): users, games, game_levels, game_categories, game_presses,
// game_session_totals, game_session_levels, game_stats_totals, game_stats_levels,
// game_records, game_reports, content_metas, content_items, content_seeks,
// user_masters, user_unknowns, user_reviews, user_favorites, user_beans,
// user_redeems, user_referrals, images, audios, notices, feedbacks,
// posts, adm_users, adm_roles, adm_permits, adm_menus, adm_operate_logs,
// game_groups, game_group_items, templates, template_levels, template_items,
// and their supporting tables.
func (r *M20260321000001BaselineExistingSchema) Up() error {
	// Verify connectivity — if this runs, the DB is reachable and tables exist.
	var count int64
	return facades.Orm().Query().Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&count)
}

// Down is intentionally empty — we never want to drop the baseline schema.
func (r *M20260321000001BaselineExistingSchema) Down() error {
	return nil
}
