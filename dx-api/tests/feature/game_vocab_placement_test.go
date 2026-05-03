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

type GameVocabPlacementSuite struct {
	suite.Suite
	tests.TestCase
	userID string
}

func TestGameVocabPlacementSuite(t *testing.T) {
	suite.Run(t, new(GameVocabPlacementSuite))
}

func (s *GameVocabPlacementSuite) SetupTest() {
	s.userID = s.seedPlacementUser()
}

func (s *GameVocabPlacementSuite) TearDownTest() {
	if s.userID == "" {
		return
	}
	q := facades.Orm().Query()
	_, _ = q.Exec(`DELETE FROM content_vocab_edits WHERE content_vocab_id IN (
		SELECT cv.id FROM content_vocabs cv
		JOIN game_vocabs gv ON gv.content_vocab_id = cv.id
		JOIN games g ON g.id = gv.game_id
		WHERE g.user_id = ?
	)`, s.userID)
	_, _ = q.Exec(`DELETE FROM game_vocabs WHERE game_id IN (SELECT id FROM games WHERE user_id = ?)`, s.userID)
	_, _ = q.Exec(`DELETE FROM game_levels WHERE game_id IN (SELECT id FROM games WHERE user_id = ?)`, s.userID)
	_, _ = q.Exec(`DELETE FROM games WHERE user_id = ?`, s.userID)
	_, _ = q.Exec(`DELETE FROM users WHERE id = ?`, s.userID)
	s.userID = ""
}

func (s *GameVocabPlacementSuite) seedPlacementUser() string {
	id := uuid.Must(uuid.NewV7()).String()
	u := models.User{
		ID:         id,
		Username:   "test_" + id,
		Grade:      consts.UserGradeLifetime,
		IsActive:   true,
		InviteCode: "p" + id[24:],
		Password:   "x",
	}
	s.Require().NoError(facades.Orm().Query().Create(&u))
	return id
}

func (s *GameVocabPlacementSuite) seedGame(mode string) string {
	id := uuid.Must(uuid.NewV7()).String()
	g := models.Game{
		ID:       id,
		Name:     "pl_" + id,
		UserID:   &s.userID,
		Mode:     mode,
		IsActive: true,
		Status:   consts.GameStatusDraft,
	}
	s.Require().NoError(facades.Orm().Query().Create(&g))
	return id
}

func (s *GameVocabPlacementSuite) seedLevel(gameID string) string {
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

func (s *GameVocabPlacementSuite) countGameVocabs(levelID string) int64 {
	n, err := facades.Orm().Query().Model(&models.GameVocab{}).
		Where("game_level_id", levelID).Count()
	s.Require().NoError(err)
	return n
}

// TestAdd_BatchSize_VocabMatch: adding 4 entries to vocab-match (batchSize=5) is rejected.
func (s *GameVocabPlacementSuite) TestAdd_BatchSize_VocabMatch() {
	gameID := s.seedGame(consts.GameModeVocabMatch)
	levelID := s.seedLevel(gameID)

	words4 := make([]string, 4)
	for i := range words4 {
		words4[i] = "word" + uuid.Must(uuid.NewV7()).String()[:4]
	}
	_, err := api.AddVocabsToLevel(s.userID, gameID, levelID, words4)
	s.ErrorIs(err, api.ErrBatchSizeInvalid, "4 vocabs in vocab-match (batchSize=5) must be rejected")

	// Adding exactly 5 must succeed
	words5 := make([]string, 5)
	for i := range words5 {
		words5[i] = "wrd" + uuid.Must(uuid.NewV7()).String()[:5]
	}
	result, err := api.AddVocabsToLevel(s.userID, gameID, levelID, words5)
	s.Require().NoError(err)
	s.Len(result, 5)
}

// TestAdd_BatchSize_VocabBattle: any count <= MaxMetasPerLevel works (batchSize=0).
func (s *GameVocabPlacementSuite) TestAdd_BatchSize_VocabBattle() {
	gameID := s.seedGame(consts.GameModeVocabBattle)
	levelID := s.seedLevel(gameID)

	// 3 words — no batch constraint for vocab-battle
	words := []string{"battleA", "battleB", "battleC"}
	result, err := api.AddVocabsToLevel(s.userID, gameID, levelID, words)
	s.Require().NoError(err)
	s.Len(result, 3)
}

// TestAdd_RejectsInvalidContent: punctuation and empty strings are rejected.
func (s *GameVocabPlacementSuite) TestAdd_RejectsInvalidContent() {
	gameID := s.seedGame(consts.GameModeVocabBattle)
	levelID := s.seedLevel(gameID)

	_, err := api.AddVocabsToLevel(s.userID, gameID, levelID, []string{"fast!"})
	s.ErrorIs(err, api.ErrVocabContentInvalid)

	_, err = api.AddVocabsToLevel(s.userID, gameID, levelID, []string{""})
	s.ErrorIs(err, api.ErrVocabContentEmpty)
}

// TestAdd_AllowsInLevelRepetition: same word twice creates 2 game_vocabs rows
// pointing at the same canonical content_vocabs row.
func (s *GameVocabPlacementSuite) TestAdd_AllowsInLevelRepetition() {
	gameID := s.seedGame(consts.GameModeVocabBattle)
	levelID := s.seedLevel(gameID)

	// Clean up any prior canonical row so we can count reliably
	_, _ = facades.Orm().Query().Exec(`DELETE FROM content_vocabs WHERE content_key = 'repeatme'`)

	result, err := api.AddVocabsToLevel(s.userID, gameID, levelID, []string{"repeatme", "repeatme"})
	s.Require().NoError(err)
	s.Require().Len(result, 2)

	// Both placement rows exist
	s.Equal(int64(2), s.countGameVocabs(levelID))

	// Both point at the same canonical row
	s.Equal(result[0].ContentVocabID, result[1].ContentVocabID,
		"both game_vocabs must reference the same canonical content_vocab")

	// Only one canonical row exists
	var row struct{ N int64 }
	s.Require().NoError(facades.Orm().Query().Raw(
		`SELECT COUNT(*) AS n FROM content_vocabs WHERE content_key = 'repeatme' AND deleted_at IS NULL`,
	).Scan(&row))
	s.Equal(int64(1), row.N, "canonical row must not be duplicated")
}

// TestReorder_ChangesOrder verifies ReorderGameVocab updates the order column.
func (s *GameVocabPlacementSuite) TestReorder_ChangesOrder() {
	gameID := s.seedGame(consts.GameModeVocabBattle)
	levelID := s.seedLevel(gameID)

	result, err := api.AddVocabsToLevel(s.userID, gameID, levelID, []string{"reorderme"})
	s.Require().NoError(err)
	s.Require().Len(result, 1)
	gvID := result[0].GameVocabID

	s.Require().NoError(api.ReorderGameVocab(s.userID, gameID, gvID, 9999))

	var gv models.GameVocab
	s.Require().NoError(facades.Orm().Query().Where("id", gvID).First(&gv))
	s.Equal(float64(9999), gv.Order)
}

// TestDelete_SoftDeletesPlacementOnly verifies DeleteGameVocab soft-deletes the
// game_vocabs row but leaves the content_vocabs row intact.
func (s *GameVocabPlacementSuite) TestDelete_SoftDeletesPlacementOnly() {
	_, _ = facades.Orm().Query().Exec(`DELETE FROM content_vocabs WHERE content_key = 'deleteplaceme'`)
	gameID := s.seedGame(consts.GameModeVocabBattle)
	levelID := s.seedLevel(gameID)

	result, err := api.AddVocabsToLevel(s.userID, gameID, levelID, []string{"deleteplaceme"})
	s.Require().NoError(err)
	s.Require().Len(result, 1)
	gvID := result[0].GameVocabID
	cvID := result[0].ContentVocabID

	s.Require().NoError(api.DeleteGameVocab(s.userID, gameID, gvID))

	// game_vocabs is soft-deleted (not counted)
	s.Equal(int64(0), s.countGameVocabs(levelID))

	// content_vocabs still alive
	n, err := facades.Orm().Query().Model(&models.ContentVocab{}).Where("id", cvID).Count()
	s.Require().NoError(err)
	s.Equal(int64(1), n, "canonical content_vocab must survive placement delete")
}

// TestList_ReturnsOrdered verifies GetLevelVocabs returns rows ordered by order ASC.
func (s *GameVocabPlacementSuite) TestList_ReturnsOrdered() {
	gameID := s.seedGame(consts.GameModeVocabBattle)
	levelID := s.seedLevel(gameID)

	// Add 3 words — order is assigned 1000, 2000, 3000 by the service
	words := []string{"alpha", "beta", "gamma"}
	_, err := api.AddVocabsToLevel(s.userID, gameID, levelID, words)
	s.Require().NoError(err)

	// Manually scramble the orders to verify sort
	_, _ = facades.Orm().Query().Exec(
		`UPDATE game_vocabs SET "order" = CASE
		   WHEN content_vocab_id = (SELECT id FROM content_vocabs WHERE content_key='alpha' AND deleted_at IS NULL LIMIT 1) THEN 3000
		   WHEN content_vocab_id = (SELECT id FROM content_vocabs WHERE content_key='beta' AND deleted_at IS NULL LIMIT 1) THEN 1000
		   WHEN content_vocab_id = (SELECT id FROM content_vocabs WHERE content_key='gamma' AND deleted_at IS NULL LIMIT 1) THEN 2000
		   ELSE "order"
		 END
		 WHERE game_level_id = ?`,
		levelID,
	)

	listed, err := api.GetLevelVocabs(s.userID, gameID, levelID)
	s.Require().NoError(err)
	s.Require().Len(listed, 3)
	s.Equal("beta", listed[0].Vocab.Content)
	s.Equal("gamma", listed[1].Vocab.Content)
	s.Equal("alpha", listed[2].Vocab.Content)
}
