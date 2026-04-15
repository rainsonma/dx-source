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

// seedVipUser inserts a lifetime-grade user and returns its id.
func (s *ContentDedupSuite) seedVipUser() string {
	id := uuid.Must(uuid.NewV7()).String()
	user := models.User{
		ID:         id,
		Username:   "test_" + id[:8],
		Grade:      consts.UserGradeLifetime,
		IsActive:   true,
		InviteCode: id[:8],
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
