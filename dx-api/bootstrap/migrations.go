package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20210101000001CreateJobsTable{},
		&migrations.M20260322000001CreateUsersTable{},
		&migrations.M20260322000002CreateAdmUsersTable{},
		&migrations.M20260322000003CreateAdmRolesTable{},
		&migrations.M20260322000004CreateAdmPermitsTable{},
		&migrations.M20260322000005CreateGameCategoriesTable{},
		&migrations.M20260322000006CreateGamePressesTable{},
		&migrations.M20260322000007CreateNoticesTable{},
		&migrations.M20260322000008CreateSettingsTable{},
		&migrations.M20260322000009CreateUserLoginsTable{},
		&migrations.M20260322000010CreateUserBeansTable{},
		&migrations.M20260322000011CreateUserFavoritesTable{},
		&migrations.M20260322000012CreateUserFollowsTable{},
		&migrations.M20260322000013CreateUserRedeemsTable{},
		&migrations.M20260322000014CreateUserReferralsTable{},
		&migrations.M20260322000015CreateUserSettingsTable{},
		&migrations.M20260322000016CreateGamesTable{},
		&migrations.M20260322000019CreateAdmUserRolesTable{},
		&migrations.M20260322000020CreateAdmUserPermitsTable{},
		&migrations.M20260322000021CreateAdmRolePermitsTable{},
		&migrations.M20260322000022CreateAdmMenusTable{},
		&migrations.M20260322000023CreateAdmLoginsTable{},
		&migrations.M20260322000024CreateAdmOperatesTable{},
		&migrations.M20260322000025CreateFeedbacksTable{},
		&migrations.M20260322000026CreateContentSeeksTable{},
		&migrations.M20260322000027CreateGameLevelsTable{},
		&migrations.M20260322000028CreateGameGroupsTable{},
		&migrations.M20260322000029CreateGameGroupMembersTable{},
		&migrations.M20260322000030CreateGameSubgroupsTable{},
		&migrations.M20260322000031CreateGameSubgroupMembersTable{},
		&migrations.M20260322000032CreatePostsTable{},
		&migrations.M20260322000033CreatePostCommentsTable{},
		&migrations.M20260322000034CreatePostLikesTable{},
		&migrations.M20260322000035CreatePostBookmarksTable{},
		&migrations.M20260322000036CreateContentMetasTable{},
		&migrations.M20260322000037CreateContentItemsTable{},
		&migrations.M20260322000043CreateGameReportsTable{},
		&migrations.M20260322000044CreateUserMastersTable{},
		&migrations.M20260322000045CreateUserUnknownsTable{},
		&migrations.M20260322000046CreateUserReviewsTable{},
		&migrations.M20260324000002CreateGameGroupApplicationsTable{},
		&migrations.M20260403000001CreateOrdersTable{},
		&migrations.M20260405000002CreateGameSessionsTable{},
		&migrations.M20260405000003AddGameSessionIndexes{},
		&migrations.M20260405000004CreateGameRecordsTable{},
		&migrations.M20260405000005CreateGamePksTable{},
		&migrations.M20260405000006AddGamePkIndexes{},
		&migrations.M20260414000001CreateGameMetasAndGameItemsTables{},
	}
}
