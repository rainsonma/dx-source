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
	_, _ = q.Exec(`DELETE FROM game_vocabs WHERE game_id IN (SELECT id FROM games WHERE user_id = ?)`, s.userID)
	_, _ = q.Exec(`DELETE FROM content_vocabs WHERE user_id = ?`, s.userID)
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

// seedVocab creates a content_vocabs row owned by s.userID and returns its ID.
func (s *GameVocabPlacementSuite) seedVocab(content string) string {
	res, err := api.CreateUserVocab(s.userID, api.VocabInput{
		Content:    content,
		Definition: []map[string]string{},
	})
	s.Require().NoError(err)
	return res.Vocab.ID
}

func (s *GameVocabPlacementSuite) countGameVocabs(levelID string) int64 {
	n, err := facades.Orm().Query().Model(&models.GameVocab{}).
		Where("game_level_id", levelID).Count()
	s.Require().NoError(err)
	return n
}

// TestAdd_BatchSize_VocabMatch: adding 4 vocabIDs to vocab-match (batchSize=5) is rejected.
func (s *GameVocabPlacementSuite) TestAdd_BatchSize_VocabMatch() {
	gameID := s.seedGame(consts.GameModeVocabMatch)
	levelID := s.seedLevel(gameID)

	ids4 := make([]string, 4)
	for i := range ids4 {
		ids4[i] = s.seedVocab("matchword" + uuid.Must(uuid.NewV7()).String()[:4])
	}
	_, err := api.AddVocabsToLevel(s.userID, gameID, levelID, ids4)
	s.ErrorIs(err, api.ErrBatchSizeInvalid, "4 vocabs in vocab-match (batchSize=5) must be rejected")

	// Adding exactly 5 must succeed
	ids5 := make([]string, 5)
	for i := range ids5 {
		ids5[i] = s.seedVocab("mwd" + uuid.Must(uuid.NewV7()).String()[:5])
	}
	result, err := api.AddVocabsToLevel(s.userID, gameID, levelID, ids5)
	s.Require().NoError(err)
	s.Len(result, 5)
}

// TestAdd_BatchSize_VocabBattle: any count <= MaxMetasPerLevel works (batchSize=0).
func (s *GameVocabPlacementSuite) TestAdd_BatchSize_VocabBattle() {
	gameID := s.seedGame(consts.GameModeVocabBattle)
	levelID := s.seedLevel(gameID)

	ids := []string{
		s.seedVocab("battleA"),
		s.seedVocab("battleB"),
		s.seedVocab("battleC"),
	}
	result, err := api.AddVocabsToLevel(s.userID, gameID, levelID, ids)
	s.Require().NoError(err)
	s.Len(result, 3)
}

// TestAdd_RejectsOtherUsersVocabs: User B's vocab ID passed by User A → ErrForbidden.
func (s *GameVocabPlacementSuite) TestAdd_RejectsOtherUsersVocabs() {
	// Seed a second user
	otherID := uuid.Must(uuid.NewV7()).String()
	other := models.User{
		ID:         otherID,
		Username:   "other_" + otherID,
		Grade:      consts.UserGradeLifetime,
		IsActive:   true,
		InviteCode: "o" + otherID[24:],
		Password:   "x",
	}
	s.Require().NoError(facades.Orm().Query().Create(&other))
	defer func() {
		_, _ = facades.Orm().Query().Exec(`DELETE FROM content_vocabs WHERE user_id = ?`, otherID)
		_, _ = facades.Orm().Query().Exec(`DELETE FROM users WHERE id = ?`, otherID)
	}()

	// Create a vocab owned by the other user
	otherVocab, err := api.CreateUserVocab(otherID, api.VocabInput{
		Content:    "stolen",
		Definition: []map[string]string{},
	})
	s.Require().NoError(err)

	gameID := s.seedGame(consts.GameModeVocabBattle)
	levelID := s.seedLevel(gameID)

	_, err = api.AddVocabsToLevel(s.userID, gameID, levelID, []string{otherVocab.Vocab.ID})
	s.ErrorIs(err, api.ErrForbidden, "using another user's vocab must be rejected")
}

// TestAdd_AllowsInLevelRepetition: same vocabID twice creates 2 game_vocabs rows.
func (s *GameVocabPlacementSuite) TestAdd_AllowsInLevelRepetition() {
	gameID := s.seedGame(consts.GameModeVocabBattle)
	levelID := s.seedLevel(gameID)

	vocabID := s.seedVocab("repeatme")
	result, err := api.AddVocabsToLevel(s.userID, gameID, levelID, []string{vocabID, vocabID})
	s.Require().NoError(err)
	s.Require().Len(result, 2)

	// Both placement rows exist
	s.Equal(int64(2), s.countGameVocabs(levelID))

	// Both point at the same canonical row
	s.Equal(result[0].ContentVocabID, result[1].ContentVocabID,
		"both game_vocabs must reference the same content_vocab")
}

// TestReorder_ChangesOrder verifies ReorderGameVocab updates the order column.
func (s *GameVocabPlacementSuite) TestReorder_ChangesOrder() {
	gameID := s.seedGame(consts.GameModeVocabBattle)
	levelID := s.seedLevel(gameID)

	vocabID := s.seedVocab("reorderme")
	result, err := api.AddVocabsToLevel(s.userID, gameID, levelID, []string{vocabID})
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
	gameID := s.seedGame(consts.GameModeVocabBattle)
	levelID := s.seedLevel(gameID)

	vocabID := s.seedVocab("deleteplaceme")
	result, err := api.AddVocabsToLevel(s.userID, gameID, levelID, []string{vocabID})
	s.Require().NoError(err)
	s.Require().Len(result, 1)
	gvID := result[0].GameVocabID

	s.Require().NoError(api.DeleteGameVocab(s.userID, gameID, gvID))

	// game_vocabs is soft-deleted (not counted)
	s.Equal(int64(0), s.countGameVocabs(levelID))

	// content_vocabs still alive
	n, err := facades.Orm().Query().Model(&models.ContentVocab{}).Where("id", vocabID).Count()
	s.Require().NoError(err)
	s.Equal(int64(1), n, "content_vocab must survive placement delete")
}

// TestList_ReturnsOrdered verifies GetLevelVocabs returns rows ordered by order ASC.
func (s *GameVocabPlacementSuite) TestList_ReturnsOrdered() {
	gameID := s.seedGame(consts.GameModeVocabBattle)
	levelID := s.seedLevel(gameID)

	alphaID := s.seedVocab("alpha")
	betaID := s.seedVocab("beta")
	gammaID := s.seedVocab("gamma")
	_, err := api.AddVocabsToLevel(s.userID, gameID, levelID, []string{alphaID, betaID, gammaID})
	s.Require().NoError(err)

	// Manually scramble the orders to verify sort
	_, _ = facades.Orm().Query().Exec(
		`UPDATE game_vocabs SET "order" = CASE
		   WHEN content_vocab_id = ? THEN 3000
		   WHEN content_vocab_id = ? THEN 1000
		   WHEN content_vocab_id = ? THEN 2000
		   ELSE "order"
		 END
		 WHERE game_level_id = ?`,
		alphaID, betaID, gammaID, levelID,
	)

	listed, err := api.GetLevelVocabs(s.userID, gameID, levelID)
	s.Require().NoError(err)
	s.Require().Len(listed, 3)
	s.Equal("beta", listed[0].Vocab.Content)
	s.Equal("gamma", listed[1].Vocab.Content)
	s.Equal("alpha", listed[2].Vocab.Content)
}
