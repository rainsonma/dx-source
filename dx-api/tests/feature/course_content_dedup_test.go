package feature

import (
	"testing"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/suite"

	"dx-api/app/consts"
	"dx-api/app/models"
	api "dx-api/app/services/api"
	"dx-api/tests"
)

type ContentDedupSuite struct {
	suite.Suite
	tests.TestCase
	userID string
}

func TestContentDedupSuite(t *testing.T) {
	suite.Run(t, new(ContentDedupSuite))
}

// SetupTest creates an isolated VIP user for each test and cleans up after.
func (s *ContentDedupSuite) SetupTest() {
	s.userID = s.seedVipUser()
}

func (s *ContentDedupSuite) TearDownTest() {
	if s.userID == "" {
		return
	}
	// Cascade soft-delete via raw SQL — content tests own all rows under this user.
	q := facades.Orm().Query()
	_, _ = q.Exec(`DELETE FROM game_items WHERE game_id IN (SELECT id FROM games WHERE user_id = ?)`, s.userID)
	_, _ = q.Exec(`DELETE FROM game_metas WHERE game_id IN (SELECT id FROM games WHERE user_id = ?)`, s.userID)
	_, _ = q.Exec(`DELETE FROM content_items WHERE content_meta_id IN (SELECT cm.id FROM content_metas cm JOIN game_metas gm ON gm.content_meta_id = cm.id JOIN games g ON g.id = gm.game_id WHERE g.user_id = ?)`, s.userID)
	_, _ = q.Exec(`DELETE FROM content_metas WHERE id IN (SELECT cm.id FROM content_metas cm LEFT JOIN game_metas gm ON gm.content_meta_id = cm.id LEFT JOIN games g ON g.id = gm.game_id WHERE g.user_id = ? OR g.user_id IS NULL AND cm.created_at > NOW() - INTERVAL '1 hour')`, s.userID)
	_, _ = q.Exec(`DELETE FROM game_levels WHERE game_id IN (SELECT id FROM games WHERE user_id = ?)`, s.userID)
	_, _ = q.Exec(`DELETE FROM games WHERE user_id = ?`, s.userID)
	_, _ = q.Exec(`DELETE FROM users WHERE id = ?`, s.userID)
	s.userID = ""
}

// seedVipUser inserts a lifetime-grade user and returns its id. Username
// and InviteCode use the full UUID so consecutive calls in the same
// millisecond don't collide on the users_username_unique / users_invite_code_unique
// constraints — UUID v7's first 8 hex chars are a timestamp prefix, not random.
func (s *ContentDedupSuite) seedVipUser() string {
	id := uuid.Must(uuid.NewV7()).String()
	user := models.User{
		ID:         id,
		Username:   "test_" + id,
		Grade:      consts.UserGradeLifetime,
		IsActive:   true,
		InviteCode: "c" + id[24:],
		Password:   "x",
	}
	s.Require().NoError(facades.Orm().Query().Create(&user))
	return id
}

// seedGame inserts a draft course game owned by s.userID with the given mode.
func (s *ContentDedupSuite) seedGame(mode string) string {
	id := uuid.Must(uuid.NewV7()).String()
	g := models.Game{
		ID:       id,
		Name:     "test_" + id,
		UserID:   &s.userID,
		Mode:     mode,
		IsActive: true,
		Status:   consts.GameStatusDraft,
	}
	s.Require().NoError(facades.Orm().Query().Create(&g))
	return id
}

// seedLevel inserts a level row for the given game and returns its id.
func (s *ContentDedupSuite) seedLevel(gameID string) string {
	id := uuid.Must(uuid.NewV7()).String()
	lv := models.GameLevel{
		ID:       id,
		GameID:   gameID,
		Name:     "L1",
		IsActive: true,
		Order:    1000,
	}
	s.Require().NoError(facades.Orm().Query().Create(&lv))
	return id
}

// strPtr returns a pointer to a fresh string with the given value. Tests
// use this to pass *string literals to MetadataEntry.Translation without
// introducing a named local per call site.
func strPtr(v string) *string {
	p := new(string)
	*p = v
	return p
}

// countMetasOwnedByUser returns the number of live content_metas reachable
// from the given user via the junction chain.
func (s *ContentDedupSuite) countMetasOwnedByUser(userID string) int64 {
	var n int64
	row := struct{ N int64 }{}
	s.Require().NoError(facades.Orm().Query().Raw(
		`SELECT COUNT(DISTINCT cm.id) AS n
		   FROM content_metas cm
		   JOIN game_metas gm ON gm.content_meta_id = cm.id AND gm.deleted_at IS NULL
		   JOIN game_levels gl ON gl.id = gm.game_level_id AND gl.deleted_at IS NULL
		   JOIN games g ON g.id = gl.game_id AND g.deleted_at IS NULL
		  WHERE cm.deleted_at IS NULL AND g.user_id = ?`,
		userID,
	).Scan(&row))
	n = row.N
	return n
}

// countGameMetasInLevel returns the number of live game_metas in the level.
func (s *ContentDedupSuite) countGameMetasInLevel(levelID string) int64 {
	n, err := facades.Orm().Query().Model(&models.GameMeta{}).
		Where("game_level_id", levelID).Count()
	s.Require().NoError(err)
	return n
}

// countGameItemsInLevel returns the number of live game_items in the level.
func (s *ContentDedupSuite) countGameItemsInLevel(levelID string) int64 {
	n, err := facades.Orm().Query().Model(&models.GameItem{}).
		Where("game_level_id", levelID).Count()
	s.Require().NoError(err)
	return n
}

// Smoke test — verifies the suite boots and the seed helpers work.
func (s *ContentDedupSuite) TestSetup_BootsAndSeeds() {
	s.NotEmpty(s.userID)
	gameID := s.seedGame(consts.GameModeWordSentence)
	levelID := s.seedLevel(gameID)
	s.NotEmpty(gameID)
	s.NotEmpty(levelID)
}

// TestSave_FreshEntries_AllCreated verifies that submitting brand-new metadata
// creates one content_metas row and one game_metas row per entry.
func (s *ContentDedupSuite) TestSave_FreshEntries_AllCreated() {
	gameID := s.seedGame(consts.GameModeWordSentence)
	levelID := s.seedLevel(gameID)

	entries := []api.MetadataEntry{
		{SourceData: "Hello world.", Translation: strPtr("你好世界。"), SourceType: "sentence"},
		{SourceData: "apple", Translation: strPtr("苹果"), SourceType: "vocab"},
	}

	count, err := api.SaveMetadataBatch(s.userID, gameID, levelID, entries, "manual")
	s.Require().NoError(err)
	s.Equal(2, count)

	s.Equal(int64(2), s.countMetasOwnedByUser(s.userID))
	s.Equal(int64(2), s.countGameMetasInLevel(levelID))
}

// TestSave_DedupAcrossGames_ReusesContentMeta verifies that saving the same
// content into a second game owned by the same user reuses the existing
// content_metas row instead of creating a new one.
func (s *ContentDedupSuite) TestSave_DedupAcrossGames_ReusesContentMeta() {
	// Game A — original content
	gameA := s.seedGame(consts.GameModeWordSentence)
	levelA := s.seedLevel(gameA)
	entries := []api.MetadataEntry{
		{SourceData: "shared sentence", Translation: strPtr("共享句子"), SourceType: "sentence"},
	}
	_, err := api.SaveMetadataBatch(s.userID, gameA, levelA, entries, "manual")
	s.Require().NoError(err)
	s.Equal(int64(1), s.countMetasOwnedByUser(s.userID))

	// Game B — same content, different game
	gameB := s.seedGame(consts.GameModeWordSentence)
	levelB := s.seedLevel(gameB)
	_, err = api.SaveMetadataBatch(s.userID, gameB, levelB, entries, "manual")
	s.Require().NoError(err)

	// content_metas row count must still be 1 (deduped), but both levels
	// have one game_metas junction row each.
	s.Equal(int64(1), s.countMetasOwnedByUser(s.userID), "content_metas should be reused")
	s.Equal(int64(1), s.countGameMetasInLevel(levelA))
	s.Equal(int64(1), s.countGameMetasInLevel(levelB))
}

// TestSave_WithinBatchRepetition_OneMetaTwoJunctions verifies that submitting
// the same entry twice in one batch creates ONE content_meta row but TWO
// game_meta junction rows.
func (s *ContentDedupSuite) TestSave_WithinBatchRepetition_OneMetaTwoJunctions() {
	gameID := s.seedGame(consts.GameModeWordSentence)
	levelID := s.seedLevel(gameID)

	entries := []api.MetadataEntry{
		{SourceData: "repeat me", Translation: strPtr("重复"), SourceType: "vocab"},
		{SourceData: "repeat me", Translation: strPtr("重复"), SourceType: "vocab"},
	}

	count, err := api.SaveMetadataBatch(s.userID, gameID, levelID, entries, "manual")
	s.Require().NoError(err)
	s.Equal(2, count)

	s.Equal(int64(1), s.countMetasOwnedByUser(s.userID), "content_metas deduped within batch")
	s.Equal(int64(2), s.countGameMetasInLevel(levelID), "two junction rows created")
}

// TestSave_NullEqualsEmpty verifies that NULL and "" translations are treated
// as equivalent for dedup purposes.
func (s *ContentDedupSuite) TestSave_NullEqualsEmpty() {
	gameID := s.seedGame(consts.GameModeWordSentence)
	levelID := s.seedLevel(gameID)

	// Save with NULL translation
	_, err := api.SaveMetadataBatch(s.userID, gameID, levelID,
		[]api.MetadataEntry{{SourceData: "fish", Translation: nil, SourceType: "vocab"}},
		"manual")
	s.Require().NoError(err)
	s.Equal(int64(1), s.countMetasOwnedByUser(s.userID))

	// Save with empty-string translation in another level
	level2 := s.seedLevel(gameID)
	emptyStr := ""
	_, err = api.SaveMetadataBatch(s.userID, gameID, level2,
		[]api.MetadataEntry{{SourceData: "fish", Translation: &emptyStr, SourceType: "vocab"}},
		"manual")
	s.Require().NoError(err)

	s.Equal(int64(1), s.countMetasOwnedByUser(s.userID), "NULL and empty translation must dedup")
	s.Equal(int64(1), s.countGameMetasInLevel(levelID))
	s.Equal(int64(1), s.countGameMetasInLevel(level2))
}

// TestSave_DifferentTranslations_DoNotDedup verifies that the same source_data
// with different translations creates two separate content_metas rows.
func (s *ContentDedupSuite) TestSave_DifferentTranslations_DoNotDedup() {
	gameID := s.seedGame(consts.GameModeWordSentence)
	levelID := s.seedLevel(gameID)

	_, err := api.SaveMetadataBatch(s.userID, gameID, levelID,
		[]api.MetadataEntry{{SourceData: "bank", Translation: strPtr("银行"), SourceType: "vocab"}},
		"manual")
	s.Require().NoError(err)

	level2 := s.seedLevel(gameID)
	_, err = api.SaveMetadataBatch(s.userID, gameID, level2,
		[]api.MetadataEntry{{SourceData: "bank", Translation: strPtr("河岸"), SourceType: "vocab"}},
		"manual")
	s.Require().NoError(err)

	s.Equal(int64(2), s.countMetasOwnedByUser(s.userID), "different translations are not deduped")
}

// TestSave_DifferentSourceTypes_DoNotDedup verifies that the same source_data
// with different source_types creates two separate content_metas rows.
func (s *ContentDedupSuite) TestSave_DifferentSourceTypes_DoNotDedup() {
	gameID := s.seedGame(consts.GameModeWordSentence)
	levelID := s.seedLevel(gameID)

	_, err := api.SaveMetadataBatch(s.userID, gameID, levelID,
		[]api.MetadataEntry{
			{SourceData: "apple", Translation: strPtr("苹果"), SourceType: "vocab"},
			{SourceData: "apple", Translation: strPtr("苹果"), SourceType: "sentence"},
		},
		"manual")
	s.Require().NoError(err)

	s.Equal(int64(2), s.countMetasOwnedByUser(s.userID), "different source_types are not deduped")
	s.Equal(int64(2), s.countGameMetasInLevel(levelID))
}

// TestSave_CrossUserIsolation verifies that User B's identical content does NOT
// reuse User A's content_meta. Each user has a private dedup pool.
func (s *ContentDedupSuite) TestSave_CrossUserIsolation() {
	// User A saves "secret" (using the SetupTest-provided s.userID)
	gameA := s.seedGame(consts.GameModeWordSentence)
	levelA := s.seedLevel(gameA)
	_, err := api.SaveMetadataBatch(s.userID, gameA, levelA,
		[]api.MetadataEntry{{SourceData: "secret", Translation: strPtr("秘密"), SourceType: "vocab"}},
		"manual")
	s.Require().NoError(err)

	// User B saves the same content in their own game. Seed User B directly
	// (can't use s.seedGame / s.seedLevel because those read s.userID).
	userB := s.seedVipUser()
	defer func() {
		q := facades.Orm().Query()
		_, _ = q.Exec(`DELETE FROM game_items WHERE game_id IN (SELECT id FROM games WHERE user_id = ?)`, userB)
		_, _ = q.Exec(`DELETE FROM game_metas WHERE game_id IN (SELECT id FROM games WHERE user_id = ?)`, userB)
		_, _ = q.Exec(`DELETE FROM content_metas WHERE id IN (SELECT cm.id FROM content_metas cm JOIN game_metas gm ON gm.content_meta_id = cm.id JOIN games g ON g.id = gm.game_id WHERE g.user_id = ?)`, userB)
		_, _ = q.Exec(`DELETE FROM game_levels WHERE game_id IN (SELECT id FROM games WHERE user_id = ?)`, userB)
		_, _ = q.Exec(`DELETE FROM games WHERE user_id = ?`, userB)
		_, _ = q.Exec(`DELETE FROM users WHERE id = ?`, userB)
	}()

	gameB := uuid.Must(uuid.NewV7()).String()
	gB := models.Game{ID: gameB, Name: "B", UserID: &userB, Mode: consts.GameModeWordSentence, IsActive: true, Status: consts.GameStatusDraft}
	s.Require().NoError(facades.Orm().Query().Create(&gB))
	levelB := uuid.Must(uuid.NewV7()).String()
	lvB := models.GameLevel{ID: levelB, GameID: gameB, Name: "L1", IsActive: true, Order: 1000}
	s.Require().NoError(facades.Orm().Query().Create(&lvB))

	_, err = api.SaveMetadataBatch(userB, gameB, levelB,
		[]api.MetadataEntry{{SourceData: "secret", Translation: strPtr("秘密"), SourceType: "vocab"}},
		"manual")
	s.Require().NoError(err)

	// Each user should own exactly 1 content_meta — no leakage.
	s.Equal(int64(1), s.countMetasOwnedByUser(s.userID))
	s.Equal(int64(1), s.countMetasOwnedByUser(userB))
}

// TestSave_ReuseBrokenDownMeta_CopiesItemsViaJunction verifies that reusing a
// meta with is_break_done=true creates parallel game_items rows in the new
// level pointing at the existing content_items (no row duplication).
func (s *ContentDedupSuite) TestSave_ReuseBrokenDownMeta_CopiesItemsViaJunction() {
	// Set up Game A with a broken-down meta
	gameA := s.seedGame(consts.GameModeWordSentence)
	levelA := s.seedLevel(gameA)
	_, err := api.SaveMetadataBatch(s.userID, gameA, levelA,
		[]api.MetadataEntry{{SourceData: "broken sentence", Translation: strPtr("已拆解"), SourceType: "sentence"}},
		"manual")
	s.Require().NoError(err)

	// Find the meta and mark it broken-down with two manual content_items.
	var metaID string
	row := struct{ ID string }{}
	s.Require().NoError(facades.Orm().Query().Raw(
		`SELECT cm.id FROM content_metas cm
		   JOIN game_metas gm ON gm.content_meta_id = cm.id
		   JOIN game_levels gl ON gl.id = gm.game_level_id
		   JOIN games g ON g.id = gl.game_id
		  WHERE g.user_id = ? AND cm.source_data = 'broken sentence'`,
		s.userID,
	).Scan(&row))
	metaID = row.ID
	s.Require().NotEmpty(metaID)

	item1 := models.ContentItem{ID: uuid.Must(uuid.NewV7()).String(), ContentMetaID: &metaID, Content: "broken", ContentType: "word"}
	item2 := models.ContentItem{ID: uuid.Must(uuid.NewV7()).String(), ContentMetaID: &metaID, Content: "sentence", ContentType: "word"}
	s.Require().NoError(facades.Orm().Query().Create(&item1))
	s.Require().NoError(facades.Orm().Query().Create(&item2))

	// Link items to level A via game_items junction
	gi1 := models.GameItem{ID: uuid.Must(uuid.NewV7()).String(), GameID: gameA, GameLevelID: levelA, ContentItemID: item1.ID, Order: 1000}
	gi2 := models.GameItem{ID: uuid.Must(uuid.NewV7()).String(), GameID: gameA, GameLevelID: levelA, ContentItemID: item2.ID, Order: 2000}
	s.Require().NoError(facades.Orm().Query().Create(&gi1))
	s.Require().NoError(facades.Orm().Query().Create(&gi2))

	// Mark meta as broken-down
	_, err = facades.Orm().Query().Exec(`UPDATE content_metas SET is_break_done = true WHERE id = ?`, metaID)
	s.Require().NoError(err)

	// Now reuse the meta in a new level via dedup
	gameB := s.seedGame(consts.GameModeWordSentence)
	levelB := s.seedLevel(gameB)
	_, err = api.SaveMetadataBatch(s.userID, gameB, levelB,
		[]api.MetadataEntry{{SourceData: "broken sentence", Translation: strPtr("已拆解"), SourceType: "sentence"}},
		"manual")
	s.Require().NoError(err)

	// content_items table should still have only 2 rows for this meta (no duplication)
	var itemCount int64
	itemCount, err = facades.Orm().Query().Model(&models.ContentItem{}).Where("content_meta_id", metaID).Count()
	s.Require().NoError(err)
	s.Equal(int64(2), itemCount, "content_items should not be duplicated")

	// Level B should now have 2 game_items pointing at the existing content_items
	s.Equal(int64(2), s.countGameItemsInLevel(levelB))

	// Both should reference the original IDs
	var itemIDs []string
	itemIDsRows := []struct{ ContentItemID string }{}
	s.Require().NoError(facades.Orm().Query().Raw(
		`SELECT content_item_id FROM game_items WHERE game_level_id = ? ORDER BY "order"`,
		levelB,
	).Scan(&itemIDsRows))
	for _, r := range itemIDsRows {
		itemIDs = append(itemIDs, r.ContentItemID)
	}
	s.ElementsMatch([]string{item1.ID, item2.ID}, itemIDs)
}

// TestDeleteMetadata_PreservesSharedMeta verifies that deleting a meta from
// one level leaves a shared meta intact in another level.
func (s *ContentDedupSuite) TestDeleteMetadata_PreservesSharedMeta() {
	gameA := s.seedGame(consts.GameModeWordSentence)
	levelA := s.seedLevel(gameA)
	gameB := s.seedGame(consts.GameModeWordSentence)
	levelB := s.seedLevel(gameB)

	entries := []api.MetadataEntry{
		{SourceData: "shared", Translation: strPtr("共享"), SourceType: "vocab"},
	}
	_, err := api.SaveMetadataBatch(s.userID, gameA, levelA, entries, "manual")
	s.Require().NoError(err)
	_, err = api.SaveMetadataBatch(s.userID, gameB, levelB, entries, "manual")
	s.Require().NoError(err)

	// Find the shared meta ID
	var metaID string
	row := struct{ ID string }{}
	s.Require().NoError(facades.Orm().Query().Raw(
		`SELECT id FROM content_metas WHERE source_data = 'shared' AND deleted_at IS NULL LIMIT 1`,
	).Scan(&row))
	metaID = row.ID

	// Delete it from level A
	s.Require().NoError(api.DeleteMetadata(s.userID, gameA, levelA, metaID))

	// content_metas row should still exist (level B references it)
	s.Equal(int64(1), s.countMetasOwnedByUser(s.userID))
	// Level A junction is gone
	s.Equal(int64(0), s.countGameMetasInLevel(levelA))
	// Level B junction still present
	s.Equal(int64(1), s.countGameMetasInLevel(levelB))
}

// TestDeleteMetadata_LastReferenceSoftDeletesUnderlying verifies that deleting
// the only reference to a meta DOES soft-delete the underlying content_metas.
func (s *ContentDedupSuite) TestDeleteMetadata_LastReferenceSoftDeletesUnderlying() {
	gameID := s.seedGame(consts.GameModeWordSentence)
	levelID := s.seedLevel(gameID)
	entries := []api.MetadataEntry{
		{SourceData: "lonely", Translation: strPtr("孤独"), SourceType: "vocab"},
	}
	_, err := api.SaveMetadataBatch(s.userID, gameID, levelID, entries, "manual")
	s.Require().NoError(err)

	var metaID string
	row := struct{ ID string }{}
	s.Require().NoError(facades.Orm().Query().Raw(
		`SELECT id FROM content_metas WHERE source_data = 'lonely' AND deleted_at IS NULL LIMIT 1`,
	).Scan(&row))
	metaID = row.ID

	s.Require().NoError(api.DeleteMetadata(s.userID, gameID, levelID, metaID))

	s.Equal(int64(0), s.countMetasOwnedByUser(s.userID), "underlying should be soft-deleted")

	// Verify it's soft-deleted not hard-deleted
	dtRow := struct {
		DeletedAt *string `gorm:"column:deleted_at"`
	}{}
	s.Require().NoError(facades.Orm().Query().Raw(
		`SELECT deleted_at FROM content_metas WHERE id = ?`, metaID,
	).Scan(&dtRow))
	s.NotNil(dtRow.DeletedAt, "soft delete sets deleted_at")
}

// TestSave_CapacityCountsDedupedEntries verifies that deduped entries STILL
// count toward the level's capacity limit (capacity reflects displayed items,
// not unique underlying rows).
func (s *ContentDedupSuite) TestSave_CapacityCountsDedupedEntries() {
	gameID := s.seedGame(consts.GameModeVocabMatch)
	levelID := s.seedLevel(gameID)

	// Pre-fill the level to one short of MaxMetasPerLevel.
	preEntries := make([]api.MetadataEntry, consts.MaxMetasPerLevel-1)
	for i := range preEntries {
		preEntries[i] = api.MetadataEntry{
			SourceData:  uuid.Must(uuid.NewV7()).String(),
			Translation: strPtr("t"),
			SourceType:  "vocab",
		}
	}
	_, err := api.SaveMetadataBatch(s.userID, gameID, levelID, preEntries, "manual")
	s.Require().NoError(err)

	// Now try to add 2 entries that would dedup to 0 new content_metas — but
	// they are 2 displayable items, pushing the level over capacity.
	entries := []api.MetadataEntry{
		preEntries[0], // dedups
		preEntries[1], // dedups
	}
	_, err = api.SaveMetadataBatch(s.userID, gameID, levelID, entries, "manual")
	s.Error(err, "capacity check must consider total entries, not just net new metas")
}
