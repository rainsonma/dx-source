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

type LevelContentBranchingSuite struct {
	suite.Suite
	tests.TestCase
	userID string
}

func TestLevelContentBranchingSuite(t *testing.T) {
	suite.Run(t, new(LevelContentBranchingSuite))
}

func (s *LevelContentBranchingSuite) SetupTest() {
	s.userID = s.seedBranchUser()
}

func (s *LevelContentBranchingSuite) TearDownTest() {
	if s.userID == "" {
		return
	}
	q := facades.Orm().Query()
	_, _ = q.Exec(`DELETE FROM content_vocab_edits WHERE content_vocab_id IN (
		SELECT cv.id FROM content_vocabs cv
		JOIN game_vocabs gv ON gv.content_vocab_id = cv.id
		JOIN games g ON g.id = gv.game_id WHERE g.user_id = ?
	)`, s.userID)
	_, _ = q.Exec(`DELETE FROM game_vocabs WHERE game_id IN (SELECT id FROM games WHERE user_id = ?)`, s.userID)
	_, _ = q.Exec(`DELETE FROM content_items WHERE game_id IN (SELECT id FROM games WHERE user_id = ?)`, s.userID)
	_, _ = q.Exec(`DELETE FROM game_levels WHERE game_id IN (SELECT id FROM games WHERE user_id = ?)`, s.userID)
	_, _ = q.Exec(`DELETE FROM games WHERE user_id = ?`, s.userID)
	_, _ = q.Exec(`DELETE FROM users WHERE id = ?`, s.userID)
	s.userID = ""
}

func (s *LevelContentBranchingSuite) seedBranchUser() string {
	id := uuid.Must(uuid.NewV7()).String()
	u := models.User{
		ID:         id,
		Username:   "test_" + id,
		Grade:      consts.UserGradeLifetime,
		IsActive:   true,
		InviteCode: "b" + id[24:],
		Password:   "x",
	}
	s.Require().NoError(facades.Orm().Query().Create(&u))
	return id
}

// seedWSGame seeds a published word-sentence game with one level at order 1 (first = free).
func (s *LevelContentBranchingSuite) seedWSGame() (gameID, levelID string) {
	gameID = uuid.Must(uuid.NewV7()).String()
	g := models.Game{
		ID:       gameID,
		Name:     "ws_" + gameID,
		UserID:   &s.userID,
		Mode:     consts.GameModeWordSentence,
		IsActive: true,
		Status:   consts.GameStatusPublished,
	}
	s.Require().NoError(facades.Orm().Query().Create(&g))

	levelID = uuid.Must(uuid.NewV7()).String()
	lv := models.GameLevel{
		ID:       levelID,
		GameID:   gameID,
		Name:     "L1",
		IsActive: true,
		Order:    1,
	}
	s.Require().NoError(facades.Orm().Query().Create(&lv))
	return
}

// seedVocabGame seeds a published vocab-match game with one level at order 1.
func (s *LevelContentBranchingSuite) seedVocabGame() (gameID, levelID string) {
	gameID = uuid.Must(uuid.NewV7()).String()
	g := models.Game{
		ID:       gameID,
		Name:     "vm_" + gameID,
		UserID:   &s.userID,
		Mode:     consts.GameModeVocabMatch,
		IsActive: true,
		Status:   consts.GameStatusPublished,
	}
	s.Require().NoError(facades.Orm().Query().Create(&g))

	levelID = uuid.Must(uuid.NewV7()).String()
	lv := models.GameLevel{
		ID:       levelID,
		GameID:   gameID,
		Name:     "L1",
		IsActive: true,
		Order:    1,
	}
	s.Require().NoError(facades.Orm().Query().Create(&lv))
	return
}

// seedContentItem inserts a content_item row directly for a WS level.
func (s *LevelContentBranchingSuite) seedContentItem(gameID, levelID, content string, order float64) string {
	id := uuid.Must(uuid.NewV7()).String()
	item := models.ContentItem{
		ID:          id,
		GameID:      gameID,
		GameLevelID: levelID,
		Content:     content,
		ContentType: "word",
		Order:       order,
	}
	s.Require().NoError(facades.Orm().Query().Create(&item))
	return id
}

// seedVocabWithCanonical inserts a content_vocabs row and a game_vocabs placement row.
func (s *LevelContentBranchingSuite) seedVocabWithCanonical(gameID, levelID, content string, order float64, ukPhonetic *string) (gameVocabID, contentVocabID string) {
	key := api.NormalizeVocabContent(content)

	var cv models.ContentVocab
	if err := facades.Orm().Query().Where("content_key", key).First(&cv); err != nil || cv.ID == "" {
		cv = models.ContentVocab{
			ID:         uuid.Must(uuid.NewV7()).String(),
			Content:    content,
			ContentKey: key,
			UkPhonetic: ukPhonetic,
		}
		s.Require().NoError(facades.Orm().Query().Create(&cv))
	}
	contentVocabID = cv.ID

	gvID := uuid.Must(uuid.NewV7()).String()
	gv := models.GameVocab{
		ID:             gvID,
		GameID:         gameID,
		GameLevelID:    levelID,
		ContentVocabID: contentVocabID,
		Order:          order,
	}
	s.Require().NoError(facades.Orm().Query().Create(&gv))
	gameVocabID = gvID
	return
}

// TestGetLevelContent_WordSentence_ReturnsContentItems verifies WS mode returns
// ContentItemData from content_items.
func (s *LevelContentBranchingSuite) TestGetLevelContent_WordSentence_ReturnsContentItems() {
	gameID, levelID := s.seedWSGame()
	itemID := s.seedContentItem(gameID, levelID, "hello", 1000)

	items, err := api.GetLevelContent(s.userID, levelID, "")
	s.Require().NoError(err)
	s.Require().Len(items, 1)
	s.Equal(itemID, items[0].ID)
	s.Equal("hello", items[0].Content)
	s.Equal("word", items[0].ContentType)
}

// TestGetLevelContent_VocabMode_SynthesizesEnvelope verifies vocab mode returns
// ContentItemData with id=game_vocab_id, contentType="vocab", and items synthesized.
func (s *LevelContentBranchingSuite) TestGetLevelContent_VocabMode_SynthesizesEnvelope() {
	// Ensure no stale canonical row
	_, _ = facades.Orm().Query().Exec(`DELETE FROM content_vocabs WHERE content_key = 'bright'`)

	gameID, levelID := s.seedVocabGame()
	uk := "/braɪt/"
	gvID, _ := s.seedVocabWithCanonical(gameID, levelID, "bright", 1000, &uk)

	// Set definition so gloss is populated
	_, _ = facades.Orm().Query().Exec(
		`UPDATE content_vocabs SET definition = '[{"adj":"明亮的"}]'
		   WHERE content_key = 'bright' AND deleted_at IS NULL`,
	)

	items, err := api.GetLevelContent(s.userID, levelID, "")
	s.Require().NoError(err)
	s.Require().Len(items, 1)

	item := items[0]
	s.Equal(gvID, item.ID, "id must equal game_vocab_id")
	s.Equal("bright", item.Content)
	s.Equal("vocab", item.ContentType)
	s.NotNil(item.Definition, "definition should be joined gloss")
	s.Contains(*item.Definition, "明亮的")
	s.NotNil(item.Items, "items must be synthesized")
	s.Contains(*item.Items, "bright")
}

// TestGetLevelContent_VocabMode_OrderingByGameVocabsOrder verifies ordering follows
// game_vocabs.order ASC.
func (s *LevelContentBranchingSuite) TestGetLevelContent_VocabMode_OrderingByGameVocabsOrder() {
	for _, key := range []string{"apple", "banana", "cherry"} {
		_, _ = facades.Orm().Query().Exec(`DELETE FROM content_vocabs WHERE content_key = ?`, key)
	}

	gameID, levelID := s.seedVocabGame()
	// Insert out of order: cherry 1000, apple 2000, banana 3000 — then scramble
	s.seedVocabWithCanonical(gameID, levelID, "cherry", 1000, nil)
	s.seedVocabWithCanonical(gameID, levelID, "apple", 2000, nil)
	s.seedVocabWithCanonical(gameID, levelID, "banana", 3000, nil)

	items, err := api.GetLevelContent(s.userID, levelID, "")
	s.Require().NoError(err)
	s.Require().Len(items, 3)
	s.Equal("cherry", items[0].Content)
	s.Equal("apple", items[1].Content)
	s.Equal("banana", items[2].Content)
}

// TestGetLevelContent_VocabMode_NullPhonetic_StillReturnsItems verifies that null
// phonetic/definition still produces a single-element items array.
func (s *LevelContentBranchingSuite) TestGetLevelContent_VocabMode_NullPhonetic_StillReturnsItems() {
	_, _ = facades.Orm().Query().Exec(`DELETE FROM content_vocabs WHERE content_key = 'sparrow'`)

	gameID, levelID := s.seedVocabGame()
	// Seed with no phonetic, no definition
	s.seedVocabWithCanonical(gameID, levelID, "sparrow", 1000, nil)

	items, err := api.GetLevelContent(s.userID, levelID, "")
	s.Require().NoError(err)
	s.Require().Len(items, 1)

	item := items[0]
	s.Equal("sparrow", item.Content)
	s.Equal("vocab", item.ContentType)
	s.Nil(item.Definition, "no definition means nil field")
	// items JSON still synthesized with empty strings
	s.NotNil(item.Items, "items must not be nil even with null phonetic/definition")
	s.Contains(*item.Items, "sparrow")
}
