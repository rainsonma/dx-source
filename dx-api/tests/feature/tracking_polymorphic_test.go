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

type TrackingPolymorphicSuite struct {
	suite.Suite
	tests.TestCase
	userID      string
	gameID      string
	gameLevelID string
}

func TestTrackingPolymorphicSuite(t *testing.T) {
	suite.Run(t, new(TrackingPolymorphicSuite))
}

func (s *TrackingPolymorphicSuite) SetupTest() {
	s.userID = s.seedTrackingUser()
	s.gameID, s.gameLevelID = s.seedPublishedWSGame()
}

func (s *TrackingPolymorphicSuite) TearDownTest() {
	if s.userID == "" {
		return
	}
	q := facades.Orm().Query()
	_, _ = q.Exec(`DELETE FROM user_masters WHERE user_id = ?`, s.userID)
	_, _ = q.Exec(`DELETE FROM user_unknowns WHERE user_id = ?`, s.userID)
	_, _ = q.Exec(`DELETE FROM user_reviews WHERE user_id = ?`, s.userID)
	_, _ = q.Exec(`DELETE FROM game_vocabs WHERE game_id = ?`, s.gameID)
	_, _ = q.Exec(`DELETE FROM content_vocabs WHERE user_id = ?`, s.userID)
	_, _ = q.Exec(`DELETE FROM content_items WHERE game_id = ?`, s.gameID)
	_, _ = q.Exec(`DELETE FROM game_levels WHERE game_id = ?`, s.gameID)
	_, _ = q.Exec(`DELETE FROM games WHERE id = ?`, s.gameID)
	_, _ = q.Exec(`DELETE FROM users WHERE id = ?`, s.userID)
	s.userID = ""
	s.gameID = ""
	s.gameLevelID = ""
}

func (s *TrackingPolymorphicSuite) seedTrackingUser() string {
	id := uuid.Must(uuid.NewV7()).String()
	u := models.User{
		ID:         id,
		Username:   "test_" + id,
		Grade:      consts.UserGradeLifetime,
		IsActive:   true,
		InviteCode: "t" + id[24:],
		Password:   "x",
	}
	s.Require().NoError(facades.Orm().Query().Create(&u))
	return id
}

func (s *TrackingPolymorphicSuite) seedPublishedWSGame() (string, string) {
	gameID := uuid.Must(uuid.NewV7()).String()
	g := models.Game{
		ID:       gameID,
		Name:     "trk_" + gameID,
		UserID:   &s.userID,
		Mode:     consts.GameModeWordSentence,
		IsActive: true,
		Status:   consts.GameStatusPublished,
	}
	s.Require().NoError(facades.Orm().Query().Create(&g))

	levelID := uuid.Must(uuid.NewV7()).String()
	lv := models.GameLevel{
		ID:       levelID,
		GameID:   gameID,
		Name:     "L1",
		IsActive: true,
		Order:    1,
	}
	s.Require().NoError(facades.Orm().Query().Create(&lv))
	return gameID, levelID
}

func (s *TrackingPolymorphicSuite) seedContentItem(content string) string {
	id := uuid.Must(uuid.NewV7()).String()
	item := models.ContentItem{
		ID:          id,
		GameID:      s.gameID,
		GameLevelID: s.gameLevelID,
		Content:     content,
		ContentType: "word",
		Order:       1000,
	}
	s.Require().NoError(facades.Orm().Query().Create(&item))
	return id
}

func (s *TrackingPolymorphicSuite) seedContentVocab(content string) string {
	key := api.NormalizeVocabContent(content)
	var cv models.ContentVocab
	if err := facades.Orm().Query().Where("user_id", s.userID).Where("content_key", key).First(&cv); err != nil || cv.ID == "" {
		cv = models.ContentVocab{
			ID:         uuid.Must(uuid.NewV7()).String(),
			UserID:     s.userID,
			Content:    content,
			ContentKey: key,
		}
		s.Require().NoError(facades.Orm().Query().Create(&cv))
	}
	return cv.ID
}

func (s *TrackingPolymorphicSuite) countMasters(userID string) int64 {
	n, err := facades.Orm().Query().Model(&models.UserMaster{}).
		Where("user_id", userID).Count()
	s.Require().NoError(err)
	return n
}

func (s *TrackingPolymorphicSuite) countUnknowns(userID string) int64 {
	n, err := facades.Orm().Query().Model(&models.UserUnknown{}).
		Where("user_id", userID).Count()
	s.Require().NoError(err)
	return n
}

func (s *TrackingPolymorphicSuite) countReviews(userID string) int64 {
	n, err := facades.Orm().Query().Model(&models.UserReview{}).
		Where("user_id", userID).Count()
	s.Require().NoError(err)
	return n
}

// ---- MarkAsMastered tests ----

// TestMark_Item_StoresContentItemID verifies item variant populates content_item_id
// and leaves content_vocab_id NULL.
func (s *TrackingPolymorphicSuite) TestMark_Item_StoresContentItemID() {
	itemID := s.seedContentItem("hello world")

	err := api.MarkAsMastered(s.userID, &itemID, nil, s.gameID, s.gameLevelID)
	s.Require().NoError(err)

	var m models.UserMaster
	s.Require().NoError(facades.Orm().Query().Where("user_id", s.userID).Where("content_item_id", itemID).First(&m))
	s.Require().NotEmpty(m.ID)
	s.Equal(itemID, *m.ContentItemID)
	s.Nil(m.ContentVocabID)
}

// TestMark_Vocab_StoresContentVocabID verifies vocab variant populates content_vocab_id
// and leaves content_item_id NULL.
func (s *TrackingPolymorphicSuite) TestMark_Vocab_StoresContentVocabID() {
	cvID := s.seedContentVocab("swift")

	err := api.MarkAsMastered(s.userID, nil, &cvID, s.gameID, s.gameLevelID)
	s.Require().NoError(err)

	var m models.UserMaster
	s.Require().NoError(facades.Orm().Query().Where("user_id", s.userID).Where("content_vocab_id", cvID).First(&m))
	s.Require().NotEmpty(m.ID)
	s.Equal(cvID, *m.ContentVocabID)
	s.Nil(m.ContentItemID)
}

// TestMark_Both_NullErrors: passing both nil returns XOR error.
func (s *TrackingPolymorphicSuite) TestMark_Both_NullErrors() {
	err := api.MarkAsMastered(s.userID, nil, nil, s.gameID, s.gameLevelID)
	s.Error(err)
	s.Contains(err.Error(), "exactly one")
}

// TestMark_Both_SetErrors: passing both non-nil returns XOR error.
func (s *TrackingPolymorphicSuite) TestMark_Both_SetErrors() {
	itemID := s.seedContentItem("both set")
	cvID := s.seedContentVocab("bothset")

	err := api.MarkAsMastered(s.userID, &itemID, &cvID, s.gameID, s.gameLevelID)
	s.Error(err)
	s.Contains(err.Error(), "exactly one")
}

// TestMark_Idempotent_ItemTwice verifies calling MarkAsMastered twice for the same
// (user, content_item_id) does NOT create a second row (ON CONFLICT DO NOTHING).
func (s *TrackingPolymorphicSuite) TestMark_Idempotent_ItemTwice() {
	itemID := s.seedContentItem("idempotent item")

	s.Require().NoError(api.MarkAsMastered(s.userID, &itemID, nil, s.gameID, s.gameLevelID))
	s.Require().NoError(api.MarkAsMastered(s.userID, &itemID, nil, s.gameID, s.gameLevelID))

	var row struct{ N int64 }
	s.Require().NoError(facades.Orm().Query().Raw(
		`SELECT COUNT(*) AS n FROM user_masters
		   WHERE user_id = ? AND content_item_id = ? AND deleted_at IS NULL`,
		s.userID, itemID,
	).Scan(&row))
	s.Equal(int64(1), row.N, "idempotent: second mark must not create duplicate")
}

// TestMark_AfterUnmark_NewRow verifies Mark → DeleteMastered → Mark creates a fresh row.
func (s *TrackingPolymorphicSuite) TestMark_AfterUnmark_NewRow() {
	itemID := s.seedContentItem("unmark me")

	s.Require().NoError(api.MarkAsMastered(s.userID, &itemID, nil, s.gameID, s.gameLevelID))

	// Find the master row to delete it
	var m models.UserMaster
	s.Require().NoError(facades.Orm().Query().Where("user_id", s.userID).
		Where("content_item_id", itemID).First(&m))
	s.Require().NotEmpty(m.ID)

	s.Require().NoError(api.DeleteMastered(s.userID, m.ID))

	// Mark again — should create a NEW row (soft-deleted row excluded from partial unique)
	s.Require().NoError(api.MarkAsMastered(s.userID, &itemID, nil, s.gameID, s.gameLevelID))

	var row struct{ N int64 }
	s.Require().NoError(facades.Orm().Query().Raw(
		`SELECT COUNT(*) AS n FROM user_masters
		   WHERE user_id = ? AND content_item_id = ? AND deleted_at IS NULL`,
		s.userID, itemID,
	).Scan(&row))
	s.Equal(int64(1), row.N, "fresh row must exist after unmark + re-mark")
}

// TestList_MixedItemAndVocab verifies ListMastered returns both kinds with
// content properly resolved.
func (s *TrackingPolymorphicSuite) TestList_MixedItemAndVocab() {
	itemID := s.seedContentItem("mix item")
	cvID := s.seedContentVocab("mixvocab")

	// Set definition for vocab so gloss is resolvable
	_, _ = facades.Orm().Query().Exec(
		`UPDATE content_vocabs SET definition = '[{"n":"混合词汇"}]' WHERE id = ?`, cvID,
	)

	s.Require().NoError(api.MarkAsMastered(s.userID, &itemID, nil, s.gameID, s.gameLevelID))
	s.Require().NoError(api.MarkAsMastered(s.userID, nil, &cvID, s.gameID, s.gameLevelID))

	items, _, _, err := api.ListMastered(s.userID, "", 20)
	s.Require().NoError(err)
	s.Len(items, 2)

	// Collect content types from the list
	types := make(map[string]bool)
	for _, item := range items {
		if item.ContentItem != nil {
			types[item.ContentItem.ContentType] = true
		}
	}
	s.True(types["word"], "item-type entry must be in list")
	s.True(types["vocab"], "vocab-type entry must be in list")
}

// ---- MarkAsUnknown tests (abbreviated parallel cases) ----

// TestMarkUnknown_Item_Idempotent verifies item-variant idempotency.
func (s *TrackingPolymorphicSuite) TestMarkUnknown_Item_Idempotent() {
	itemID := s.seedContentItem("unknown item")

	s.Require().NoError(api.MarkAsUnknown(s.userID, &itemID, nil, s.gameID, s.gameLevelID))
	s.Require().NoError(api.MarkAsUnknown(s.userID, &itemID, nil, s.gameID, s.gameLevelID))

	var row struct{ N int64 }
	s.Require().NoError(facades.Orm().Query().Raw(
		`SELECT COUNT(*) AS n FROM user_unknowns
		   WHERE user_id = ? AND content_item_id = ? AND deleted_at IS NULL`,
		s.userID, itemID,
	).Scan(&row))
	s.Equal(int64(1), row.N, "idempotent: second unknown mark must not duplicate")
}

// TestMarkUnknown_Vocab_StoresVocabID verifies vocab variant populates content_vocab_id.
func (s *TrackingPolymorphicSuite) TestMarkUnknown_Vocab_StoresVocabID() {
	cvID := s.seedContentVocab("unknownvocab")

	s.Require().NoError(api.MarkAsUnknown(s.userID, nil, &cvID, s.gameID, s.gameLevelID))

	var u models.UserUnknown
	s.Require().NoError(facades.Orm().Query().Where("user_id", s.userID).
		Where("content_vocab_id", cvID).First(&u))
	s.Require().NotEmpty(u.ID)
	s.Equal(cvID, *u.ContentVocabID)
	s.Nil(u.ContentItemID)
}

// ---- MarkAsReview tests (abbreviated parallel cases) ----

// TestMarkReview_Item_Idempotent verifies item-variant idempotency.
func (s *TrackingPolymorphicSuite) TestMarkReview_Item_Idempotent() {
	itemID := s.seedContentItem("review item")

	s.Require().NoError(api.MarkAsReview(s.userID, &itemID, nil, s.gameID, s.gameLevelID))
	s.Require().NoError(api.MarkAsReview(s.userID, &itemID, nil, s.gameID, s.gameLevelID))

	var row struct{ N int64 }
	s.Require().NoError(facades.Orm().Query().Raw(
		`SELECT COUNT(*) AS n FROM user_reviews
		   WHERE user_id = ? AND content_item_id = ? AND deleted_at IS NULL`,
		s.userID, itemID,
	).Scan(&row))
	s.Equal(int64(1), row.N, "idempotent: second review mark must not duplicate")
}

// TestMarkReview_Vocab_StoresVocabID verifies vocab variant populates content_vocab_id.
func (s *TrackingPolymorphicSuite) TestMarkReview_Vocab_StoresVocabID() {
	cvID := s.seedContentVocab("reviewvocab")

	s.Require().NoError(api.MarkAsReview(s.userID, nil, &cvID, s.gameID, s.gameLevelID))

	var r models.UserReview
	s.Require().NoError(facades.Orm().Query().Where("user_id", s.userID).
		Where("content_vocab_id", cvID).First(&r))
	s.Require().NotEmpty(r.ID)
	s.Equal(cvID, *r.ContentVocabID)
	s.Nil(r.ContentItemID)
}
