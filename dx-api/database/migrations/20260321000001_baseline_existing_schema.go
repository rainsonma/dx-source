package migrations

import "dx-api/app/facades"

type M20260321000001BaselineExistingSchema struct{}

func (r *M20260321000001BaselineExistingSchema) Signature() string {
	return "20260321000001_baseline_existing_schema"
}

func (r *M20260321000001BaselineExistingSchema) Up() error {
	statements := []string{
		// 1. users
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR PRIMARY KEY,
			grade TEXT NOT NULL DEFAULT '',
			username TEXT NOT NULL DEFAULT '',
			nickname TEXT,
			email TEXT,
			phone TEXT,
			password TEXT NOT NULL DEFAULT '',
			avatar_id TEXT,
			city TEXT,
			introduction TEXT,
			is_active BOOLEAN NOT NULL DEFAULT true,
			beans INTEGER NOT NULL DEFAULT 0,
			granted_beans INTEGER NOT NULL DEFAULT 0,
			exp INTEGER NOT NULL DEFAULT 0,
			invite_code TEXT NOT NULL DEFAULT '',
			current_play_streak INTEGER NOT NULL DEFAULT 0,
			max_play_streak INTEGER NOT NULL DEFAULT 0,
			last_played_at TIMESTAMPTZ,
			vip_due_at TIMESTAMPTZ,
			last_read_notice_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 2. user_logins
		`CREATE TABLE IF NOT EXISTS user_logins (
			id VARCHAR PRIMARY KEY,
			user_id TEXT NOT NULL,
			ip TEXT NOT NULL DEFAULT '',
			agent TEXT,
			country TEXT,
			province TEXT,
			city TEXT,
			isp TEXT,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 3. user_masters
		`CREATE TABLE IF NOT EXISTS user_masters (
			id VARCHAR PRIMARY KEY,
			user_id TEXT NOT NULL,
			content_item_id TEXT NOT NULL,
			game_id TEXT NOT NULL,
			game_level_id TEXT NOT NULL,
			mastered_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 4. user_unknowns
		`CREATE TABLE IF NOT EXISTS user_unknowns (
			id VARCHAR PRIMARY KEY,
			user_id TEXT NOT NULL,
			content_item_id TEXT NOT NULL,
			game_id TEXT NOT NULL,
			game_level_id TEXT NOT NULL,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 5. user_reviews
		`CREATE TABLE IF NOT EXISTS user_reviews (
			id VARCHAR PRIMARY KEY,
			user_id TEXT NOT NULL,
			content_item_id TEXT NOT NULL,
			game_id TEXT NOT NULL,
			game_level_id TEXT NOT NULL,
			last_review_at TIMESTAMPTZ,
			next_review_at TIMESTAMPTZ,
			review_count INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 6. user_beans
		`CREATE TABLE IF NOT EXISTS user_beans (
			id VARCHAR PRIMARY KEY,
			user_id TEXT NOT NULL,
			beans INTEGER NOT NULL DEFAULT 0,
			origin INTEGER NOT NULL DEFAULT 0,
			result INTEGER NOT NULL DEFAULT 0,
			slug TEXT NOT NULL DEFAULT '',
			reason TEXT NOT NULL DEFAULT '',
			data TEXT,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 7. user_favorites
		`CREATE TABLE IF NOT EXISTS user_favorites (
			id VARCHAR PRIMARY KEY,
			user_id TEXT NOT NULL,
			game_id TEXT NOT NULL,
			created_at TIMESTAMPTZ
		)`,

		// 8. user_follows
		`CREATE TABLE IF NOT EXISTS user_follows (
			id VARCHAR PRIMARY KEY,
			follower_id TEXT NOT NULL,
			following_id TEXT NOT NULL,
			created_at TIMESTAMPTZ
		)`,

		// 9. user_redeems
		`CREATE TABLE IF NOT EXISTS user_redeems (
			id VARCHAR PRIMARY KEY,
			code TEXT NOT NULL DEFAULT '',
			grade TEXT NOT NULL DEFAULT '',
			user_id TEXT,
			redeemed_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 10. user_referrals
		`CREATE TABLE IF NOT EXISTS user_referrals (
			id VARCHAR PRIMARY KEY,
			referrer_id TEXT NOT NULL,
			invitee_id TEXT,
			status TEXT NOT NULL DEFAULT '',
			reward_amount DOUBLE PRECISION NOT NULL DEFAULT 0,
			rewarded_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 11. user_settings
		`CREATE TABLE IF NOT EXISTS user_settings (
			id VARCHAR PRIMARY KEY,
			user_id TEXT NOT NULL,
			"group" TEXT NOT NULL DEFAULT '',
			key TEXT NOT NULL DEFAULT '',
			value TEXT NOT NULL DEFAULT '',
			value_type TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 12. games
		`CREATE TABLE IF NOT EXISTS games (
			id VARCHAR PRIMARY KEY,
			name TEXT NOT NULL DEFAULT '',
			description TEXT,
			user_id TEXT,
			mode TEXT NOT NULL DEFAULT '',
			game_category_id TEXT,
			game_press_id TEXT,
			icon TEXT,
			cover_id TEXT,
			"order" DOUBLE PRECISION NOT NULL DEFAULT 0,
			is_active BOOLEAN NOT NULL DEFAULT true,
			status TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 13. game_categories
		`CREATE TABLE IF NOT EXISTS game_categories (
			id VARCHAR PRIMARY KEY,
			parent_id TEXT,
			cover_id TEXT,
			name TEXT NOT NULL DEFAULT '',
			alias TEXT,
			description TEXT,
			"order" DOUBLE PRECISION NOT NULL DEFAULT 0,
			is_enabled BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 14. game_levels
		`CREATE TABLE IF NOT EXISTS game_levels (
			id VARCHAR PRIMARY KEY,
			game_id TEXT NOT NULL,
			name TEXT NOT NULL DEFAULT '',
			description TEXT,
			"order" DOUBLE PRECISION NOT NULL DEFAULT 0,
			passing_score INTEGER NOT NULL DEFAULT 0,
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 15. game_groups
		`CREATE TABLE IF NOT EXISTS game_groups (
			id VARCHAR PRIMARY KEY,
			name TEXT NOT NULL DEFAULT '',
			description TEXT,
			owner_id TEXT NOT NULL,
			cover_id TEXT,
			current_game_id TEXT,
			invite_code TEXT NOT NULL DEFAULT '',
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 16. game_group_members
		`CREATE TABLE IF NOT EXISTS game_group_members (
			id VARCHAR PRIMARY KEY,
			game_group_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 17. game_subgroups
		`CREATE TABLE IF NOT EXISTS game_subgroups (
			id VARCHAR PRIMARY KEY,
			game_group_id TEXT NOT NULL,
			name TEXT NOT NULL DEFAULT '',
			description TEXT,
			"order" DOUBLE PRECISION NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 18. game_subgroup_members
		`CREATE TABLE IF NOT EXISTS game_subgroup_members (
			id VARCHAR PRIMARY KEY,
			game_subgroup_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 19. game_presses
		`CREATE TABLE IF NOT EXISTS game_presses (
			id VARCHAR PRIMARY KEY,
			name TEXT NOT NULL DEFAULT '',
			alias TEXT,
			cover_id TEXT,
			"order" DOUBLE PRECISION NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 20. game_records
		`CREATE TABLE IF NOT EXISTS game_records (
			id VARCHAR PRIMARY KEY,
			user_id TEXT NOT NULL,
			game_session_total_id TEXT NOT NULL,
			game_session_level_id TEXT NOT NULL,
			game_level_id TEXT NOT NULL,
			content_item_id TEXT NOT NULL,
			is_correct BOOLEAN NOT NULL DEFAULT false,
			source_answer TEXT NOT NULL DEFAULT '',
			user_answer TEXT NOT NULL DEFAULT '',
			base_score INTEGER NOT NULL DEFAULT 0,
			combo_score INTEGER NOT NULL DEFAULT 0,
			duration INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 21. game_reports
		`CREATE TABLE IF NOT EXISTS game_reports (
			id VARCHAR PRIMARY KEY,
			user_id TEXT NOT NULL,
			game_id TEXT NOT NULL,
			game_level_id TEXT NOT NULL,
			content_item_id TEXT NOT NULL,
			reason TEXT NOT NULL DEFAULT '',
			note TEXT,
			count INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 22. game_session_totals
		`CREATE TABLE IF NOT EXISTS game_session_totals (
			id VARCHAR PRIMARY KEY,
			user_id TEXT NOT NULL,
			game_id TEXT NOT NULL,
			degree TEXT NOT NULL DEFAULT '',
			pattern TEXT,
			current_level_id TEXT NOT NULL DEFAULT '',
			current_content_item_id TEXT,
			started_at TIMESTAMPTZ NOT NULL,
			last_played_at TIMESTAMPTZ NOT NULL,
			ended_at TIMESTAMPTZ,
			score INTEGER NOT NULL DEFAULT 0,
			exp INTEGER NOT NULL DEFAULT 0,
			max_combo INTEGER NOT NULL DEFAULT 0,
			correct_count INTEGER NOT NULL DEFAULT 0,
			wrong_count INTEGER NOT NULL DEFAULT 0,
			skip_count INTEGER NOT NULL DEFAULT 0,
			play_time INTEGER NOT NULL DEFAULT 0,
			total_levels_count INTEGER NOT NULL DEFAULT 0,
			played_levels_count INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 23. game_session_levels
		`CREATE TABLE IF NOT EXISTS game_session_levels (
			id VARCHAR PRIMARY KEY,
			game_session_total_id TEXT NOT NULL,
			game_level_id TEXT NOT NULL,
			current_content_item_id TEXT,
			degree TEXT NOT NULL DEFAULT '',
			pattern TEXT,
			started_at TIMESTAMPTZ NOT NULL,
			last_played_at TIMESTAMPTZ NOT NULL,
			ended_at TIMESTAMPTZ,
			score INTEGER NOT NULL DEFAULT 0,
			exp INTEGER NOT NULL DEFAULT 0,
			max_combo INTEGER NOT NULL DEFAULT 0,
			correct_count INTEGER NOT NULL DEFAULT 0,
			wrong_count INTEGER NOT NULL DEFAULT 0,
			skip_count INTEGER NOT NULL DEFAULT 0,
			play_time INTEGER NOT NULL DEFAULT 0,
			total_items_count INTEGER NOT NULL DEFAULT 0,
			played_items_count INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 24. game_stats_totals
		`CREATE TABLE IF NOT EXISTS game_stats_totals (
			id VARCHAR PRIMARY KEY,
			user_id TEXT NOT NULL,
			game_id TEXT NOT NULL,
			total_sessions INTEGER NOT NULL DEFAULT 0,
			total_exp INTEGER NOT NULL DEFAULT 0,
			highest_score INTEGER NOT NULL DEFAULT 0,
			total_scores INTEGER NOT NULL DEFAULT 0,
			total_play_time INTEGER NOT NULL DEFAULT 0,
			first_played_at TIMESTAMPTZ NOT NULL,
			last_played_at TIMESTAMPTZ NOT NULL,
			first_completed_at TIMESTAMPTZ,
			last_completed_at TIMESTAMPTZ,
			completion_count INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 25. game_stats_levels
		`CREATE TABLE IF NOT EXISTS game_stats_levels (
			id VARCHAR PRIMARY KEY,
			user_id TEXT NOT NULL,
			game_level_id TEXT NOT NULL,
			highest_score INTEGER NOT NULL DEFAULT 0,
			total_scores INTEGER NOT NULL DEFAULT 0,
			total_play_time INTEGER NOT NULL DEFAULT 0,
			first_played_at TIMESTAMPTZ NOT NULL,
			last_played_at TIMESTAMPTZ NOT NULL,
			first_completed_at TIMESTAMPTZ,
			last_completed_at TIMESTAMPTZ,
			completion_count INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 26. content_items
		`CREATE TABLE IF NOT EXISTS content_items (
			id VARCHAR PRIMARY KEY,
			game_level_id TEXT NOT NULL,
			content_meta_id TEXT,
			content TEXT NOT NULL DEFAULT '',
			content_type TEXT NOT NULL DEFAULT '',
			uk_audio_id TEXT,
			us_audio_id TEXT,
			definition TEXT,
			translation TEXT,
			explanation TEXT,
			items JSONB,
			structure JSONB,
			"order" DOUBLE PRECISION NOT NULL DEFAULT 0,
			tags TEXT[],
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 27. content_metas
		`CREATE TABLE IF NOT EXISTS content_metas (
			id VARCHAR PRIMARY KEY,
			game_level_id TEXT NOT NULL,
			source_from TEXT NOT NULL DEFAULT '',
			source_type TEXT NOT NULL DEFAULT '',
			source_data TEXT NOT NULL DEFAULT '',
			translation TEXT,
			is_break_done BOOLEAN NOT NULL DEFAULT false,
			"order" DOUBLE PRECISION NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 28. content_seeks
		`CREATE TABLE IF NOT EXISTS content_seeks (
			id VARCHAR PRIMARY KEY,
			user_id TEXT NOT NULL,
			course_name TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			disk_url TEXT NOT NULL DEFAULT '',
			count INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 29. audios
		`CREATE TABLE IF NOT EXISTS audios (
			id VARCHAR PRIMARY KEY,
			adm_user_id TEXT,
			user_id TEXT,
			url TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL DEFAULT '',
			mime TEXT NOT NULL DEFAULT '',
			size INTEGER NOT NULL DEFAULT 0,
			duration INTEGER NOT NULL DEFAULT 0,
			role TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 30. images
		`CREATE TABLE IF NOT EXISTS images (
			id VARCHAR PRIMARY KEY,
			adm_user_id TEXT,
			user_id TEXT,
			url TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL DEFAULT '',
			mime TEXT NOT NULL DEFAULT '',
			size INTEGER NOT NULL DEFAULT 0,
			role TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 31. posts
		`CREATE TABLE IF NOT EXISTS posts (
			id VARCHAR PRIMARY KEY,
			user_id TEXT NOT NULL,
			content TEXT NOT NULL DEFAULT '',
			image_id TEXT,
			tags TEXT[],
			like_count INTEGER NOT NULL DEFAULT 0,
			comment_count INTEGER NOT NULL DEFAULT 0,
			share_count INTEGER NOT NULL DEFAULT 0,
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 32. post_comments
		`CREATE TABLE IF NOT EXISTS post_comments (
			id VARCHAR PRIMARY KEY,
			post_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			content TEXT NOT NULL DEFAULT '',
			like_count INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 33. post_likes
		`CREATE TABLE IF NOT EXISTS post_likes (
			id VARCHAR PRIMARY KEY,
			post_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			created_at TIMESTAMPTZ
		)`,

		// 34. post_bookmarks
		`CREATE TABLE IF NOT EXISTS post_bookmarks (
			id VARCHAR PRIMARY KEY,
			post_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			created_at TIMESTAMPTZ
		)`,

		// 35. feedbacks
		`CREATE TABLE IF NOT EXISTS feedbacks (
			id VARCHAR PRIMARY KEY,
			user_id TEXT NOT NULL,
			type TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			count INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 36. notices
		`CREATE TABLE IF NOT EXISTS notices (
			id VARCHAR PRIMARY KEY,
			title TEXT NOT NULL DEFAULT '',
			content TEXT,
			icon TEXT,
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 37. settings
		`CREATE TABLE IF NOT EXISTS settings (
			id VARCHAR PRIMARY KEY,
			"group" TEXT NOT NULL DEFAULT '',
			label TEXT,
			key TEXT NOT NULL DEFAULT '',
			value TEXT NOT NULL DEFAULT '',
			value_type TEXT NOT NULL DEFAULT '',
			value_from TEXT NOT NULL DEFAULT '',
			value_options JSONB NOT NULL DEFAULT '{}',
			description TEXT NOT NULL DEFAULT '',
			"order" DOUBLE PRECISION NOT NULL DEFAULT 0,
			is_enabled BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 38. adm_users
		`CREATE TABLE IF NOT EXISTS adm_users (
			id VARCHAR PRIMARY KEY,
			username TEXT NOT NULL DEFAULT '',
			nickname TEXT,
			password TEXT NOT NULL DEFAULT '',
			avatar_id TEXT,
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 39. adm_roles
		`CREATE TABLE IF NOT EXISTS adm_roles (
			id VARCHAR PRIMARY KEY,
			slug TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 40. adm_permits
		`CREATE TABLE IF NOT EXISTS adm_permits (
			id VARCHAR PRIMARY KEY,
			slug TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL DEFAULT '',
			http_methods TEXT[],
			http_paths TEXT[],
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 41. adm_user_roles
		`CREATE TABLE IF NOT EXISTS adm_user_roles (
			id VARCHAR PRIMARY KEY,
			adm_role_id TEXT NOT NULL,
			adm_user_id TEXT NOT NULL,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 42. adm_user_permits
		`CREATE TABLE IF NOT EXISTS adm_user_permits (
			id VARCHAR PRIMARY KEY,
			adm_user_id TEXT NOT NULL,
			adm_permit_id TEXT NOT NULL,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 43. adm_role_permits
		`CREATE TABLE IF NOT EXISTS adm_role_permits (
			id VARCHAR PRIMARY KEY,
			adm_role_id TEXT NOT NULL,
			adm_permit_id TEXT NOT NULL,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 44. adm_menus
		`CREATE TABLE IF NOT EXISTS adm_menus (
			id VARCHAR PRIMARY KEY,
			parent_id TEXT,
			name TEXT NOT NULL DEFAULT '',
			alias TEXT,
			icon TEXT,
			uri TEXT,
			"order" DOUBLE PRECISION NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 45. adm_logins
		`CREATE TABLE IF NOT EXISTS adm_logins (
			id VARCHAR PRIMARY KEY,
			adm_user_id TEXT NOT NULL,
			ip TEXT NOT NULL DEFAULT '',
			agent TEXT,
			country TEXT,
			province TEXT,
			city TEXT,
			isp TEXT,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,

		// 46. adm_operates
		`CREATE TABLE IF NOT EXISTS adm_operates (
			id VARCHAR PRIMARY KEY,
			adm_user_id TEXT NOT NULL,
			path TEXT NOT NULL DEFAULT '',
			method TEXT NOT NULL DEFAULT '',
			ip TEXT NOT NULL DEFAULT '',
			input TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,
	}

	for _, stmt := range statements {
		if _, err := facades.Orm().Query().Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

// Down is intentionally empty — we never want to drop the baseline schema.
func (r *M20260321000001BaselineExistingSchema) Down() error {
	return nil
}
