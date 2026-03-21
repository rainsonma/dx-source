package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20210101000001CreateJobsTable{},
		&migrations.M20260322000001_CreateUsersTable{},
		&migrations.M20260322000002_CreateAdmUsersTable{},
		&migrations.M20260322000003_CreateAdmRolesTable{},
		&migrations.M20260322000004_CreateAdmPermitsTable{},
		&migrations.M20260322000005_CreateGameCategoriesTable{},
		&migrations.M20260322000006_CreateGamePressesTable{},
		&migrations.M20260322000007_CreateNoticesTable{},
		&migrations.M20260322000008_CreateSettingsTable{},
		&migrations.M20260322000009_CreateUserLoginsTable{},
		&migrations.M20260322000010_CreateUserBeansTable{},
		&migrations.M20260322000011_CreateUserFavoritesTable{},
		&migrations.M20260322000012_CreateUserFollowsTable{},
		&migrations.M20260322000013_CreateUserRedeemsTable{},
		&migrations.M20260322000014_CreateUserReferralsTable{},
		&migrations.M20260322000015_CreateUserSettingsTable{},
		&migrations.M20260322000016_CreateGamesTable{},
		&migrations.M20260322000017_CreateImagesTable{},
		&migrations.M20260322000018_CreateAudiosTable{},
		&migrations.M20260322000019_CreateAdmUserRolesTable{},
		&migrations.M20260322000020_CreateAdmUserPermitsTable{},
		&migrations.M20260322000021_CreateAdmRolePermitsTable{},
		&migrations.M20260322000022_CreateAdmMenusTable{},
		&migrations.M20260322000023_CreateAdmLoginsTable{},
		&migrations.M20260322000024_CreateAdmOperatesTable{},
		&migrations.M20260322000025_CreateFeedbacksTable{},
		&migrations.M20260322000026_CreateContentSeeksTable{},
		&migrations.M20260322000027_CreateGameLevelsTable{},
		&migrations.M20260322000028_CreateGameGroupsTable{},
		&migrations.M20260322000029_CreateGameGroupMembersTable{},
		&migrations.M20260322000030_CreateGameSubgroupsTable{},
		&migrations.M20260322000031_CreateGameSubgroupMembersTable{},
		&migrations.M20260322000032_CreatePostsTable{},
		&migrations.M20260322000033_CreatePostCommentsTable{},
		&migrations.M20260322000034_CreatePostLikesTable{},
		&migrations.M20260322000035_CreatePostBookmarksTable{},
		&migrations.M20260322000036_CreateContentMetasTable{},
		&migrations.M20260322000037_CreateContentItemsTable{},
		&migrations.M20260322000038_CreateGameSessionTotalsTable{},
		&migrations.M20260322000039_CreateGameSessionLevelsTable{},
		&migrations.M20260322000040_CreateGameRecordsTable{},
		&migrations.M20260322000041_CreateGameStatsTotalsTable{},
		&migrations.M20260322000042_CreateGameStatsLevelsTable{},
		&migrations.M20260322000043_CreateGameReportsTable{},
		&migrations.M20260322000044_CreateUserMastersTable{},
		&migrations.M20260322000045_CreateUserUnknownsTable{},
		&migrations.M20260322000046_CreateUserReviewsTable{},
	}
}
