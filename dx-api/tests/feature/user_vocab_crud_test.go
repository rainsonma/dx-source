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

type UserVocabCrudSuite struct {
	suite.Suite
	tests.TestCase
	userID  string
	userBID string
}

func TestUserVocabCrudSuite(t *testing.T) {
	suite.Run(t, new(UserVocabCrudSuite))
}

func (s *UserVocabCrudSuite) SetupTest() {
	s.userID = s.seedVocabUser("A")
	s.userBID = s.seedVocabUser("B")
}

func (s *UserVocabCrudSuite) TearDownTest() {
	q := facades.Orm().Query()
	for _, uid := range []string{s.userID, s.userBID} {
		if uid == "" {
			continue
		}
		_, _ = q.Exec(`DELETE FROM game_vocabs WHERE game_id IN (SELECT id FROM games WHERE user_id = ?)`, uid)
		_, _ = q.Exec(`DELETE FROM game_levels WHERE game_id IN (SELECT id FROM games WHERE user_id = ?)`, uid)
		_, _ = q.Exec(`DELETE FROM games WHERE user_id = ?`, uid)
		_, _ = q.Exec(`DELETE FROM content_vocabs WHERE user_id = ?`, uid)
		_, _ = q.Exec(`DELETE FROM users WHERE id = ?`, uid)
	}
	s.userID = ""
	s.userBID = ""
}

func (s *UserVocabCrudSuite) seedVocabUser(suffix string) string {
	id := uuid.Must(uuid.NewV7()).String()
	u := models.User{
		ID:         id,
		Username:   "vcrud_" + suffix + "_" + id[:8],
		Grade:      consts.UserGradeLifetime,
		IsActive:   true,
		InviteCode: suffix + id[24:],
		Password:   "x",
	}
	s.Require().NoError(facades.Orm().Query().Create(&u))
	return id
}

func (s *UserVocabCrudSuite) makeInput(word string) api.VocabInput {
	return api.VocabInput{
		Content:    word,
		Definition: []map[string]string{{"adj": "快的"}},
	}
}

// TestCreate_NewVocab_StoredWithUser verifies a created vocab has user_id set correctly.
func (s *UserVocabCrudSuite) TestCreate_NewVocab_StoredWithUser() {
	res, err := api.CreateUserVocab(s.userID, s.makeInput("swift"))
	s.Require().NoError(err)
	s.False(res.WasReused)
	s.NotEmpty(res.Vocab.ID)
	s.Equal("swift", res.Vocab.Content)

	var v models.ContentVocab
	s.Require().NoError(facades.Orm().Query().Where("id", res.Vocab.ID).First(&v))
	s.Equal(s.userID, v.UserID, "user_id must be set to the creating user")
}

// TestCreate_DuplicateContentKey_ReturnsExisting verifies idempotent create.
func (s *UserVocabCrudSuite) TestCreate_DuplicateContentKey_ReturnsExisting() {
	res1, err := api.CreateUserVocab(s.userID, s.makeInput("fast"))
	s.Require().NoError(err)
	s.False(res1.WasReused)

	// Case-insensitive match on same user
	res2, err := api.CreateUserVocab(s.userID, s.makeInput("FAST"))
	s.Require().NoError(err)
	s.True(res2.WasReused, "second create must return WasReused=true")
	s.Equal(res1.Vocab.ID, res2.Vocab.ID, "must return the same row ID")

	// Only one row in DB
	var row struct{ N int64 }
	s.Require().NoError(facades.Orm().Query().Raw(
		`SELECT COUNT(*) AS n FROM content_vocabs WHERE user_id = ? AND content_key = 'fast' AND deleted_at IS NULL`,
		s.userID,
	).Scan(&row))
	s.Equal(int64(1), row.N)
}

// TestCreate_SameContentKey_AcrossUsers_TwoRows verifies two users can each have "fast".
func (s *UserVocabCrudSuite) TestCreate_SameContentKey_AcrossUsers_TwoRows() {
	resA, err := api.CreateUserVocab(s.userID, s.makeInput("fast"))
	s.Require().NoError(err)

	resB, err := api.CreateUserVocab(s.userBID, s.makeInput("fast"))
	s.Require().NoError(err)

	s.NotEqual(resA.Vocab.ID, resB.Vocab.ID, "two users must have distinct vocab rows")

	var row struct{ N int64 }
	s.Require().NoError(facades.Orm().Query().Raw(
		`SELECT COUNT(*) AS n FROM content_vocabs WHERE content_key = 'fast' AND deleted_at IS NULL`,
	).Scan(&row))
	s.GreaterOrEqual(row.N, int64(2), "at least 2 rows — one per user")
}

// TestList_ReturnsOnlyOwnerVocabs verifies list query never includes other users' vocabs.
func (s *UserVocabCrudSuite) TestList_ReturnsOnlyOwnerVocabs() {
	_, err := api.CreateUserVocab(s.userID, s.makeInput("exclusive"))
	s.Require().NoError(err)
	_, err = api.CreateUserVocab(s.userBID, s.makeInput("exclusive"))
	s.Require().NoError(err)

	items, _, _, err := api.ListUserVocabs(s.userID, "", "", 50)
	s.Require().NoError(err)
	for _, item := range items {
		// Verify none belong to userB by checking DB
		var v models.ContentVocab
		s.Require().NoError(facades.Orm().Query().Where("id", item.ID).First(&v))
		s.Equal(s.userID, v.UserID, "listed vocab must belong to the requesting user")
	}
}

// TestUpdate_OwnVocab_Succeeds verifies a user can update their own vocab.
func (s *UserVocabCrudSuite) TestUpdate_OwnVocab_Succeeds() {
	res, err := api.CreateUserVocab(s.userID, s.makeInput("slow"))
	s.Require().NoError(err)

	updated, err := api.UpdateUserVocab(s.userID, res.Vocab.ID, api.VocabInput{
		Content:    "slower",
		Definition: []map[string]string{{"adj": "更慢的"}},
	})
	s.Require().NoError(err)
	s.Equal("slower", updated.Content)

	var v models.ContentVocab
	s.Require().NoError(facades.Orm().Query().Where("id", res.Vocab.ID).First(&v))
	s.Equal("slower", v.Content)
	s.Equal("slower", v.ContentKey)
}

// TestUpdate_OthersVocab_ErrVocabNotFound verifies updating someone else's vocab returns ErrVocabNotFound.
func (s *UserVocabCrudSuite) TestUpdate_OthersVocab_ErrVocabNotFound() {
	resB, err := api.CreateUserVocab(s.userBID, s.makeInput("forbidden"))
	s.Require().NoError(err)

	_, err = api.UpdateUserVocab(s.userID, resB.Vocab.ID, s.makeInput("forbidden"))
	s.ErrorIs(err, api.ErrVocabNotFound, "updating another user's vocab must return ErrVocabNotFound")
}

// TestDelete_SoftDelete verifies DeleteUserVocab sets deleted_at and excludes from list.
func (s *UserVocabCrudSuite) TestDelete_SoftDelete() {
	res, err := api.CreateUserVocab(s.userID, s.makeInput("ephemeral"))
	s.Require().NoError(err)
	vocabID := res.Vocab.ID

	s.Require().NoError(api.DeleteUserVocab(s.userID, vocabID))

	// Row still in DB with deleted_at set
	var v models.ContentVocab
	s.Require().NoError(facades.Orm().Query().Where("id", vocabID).WithTrashed().First(&v))
	s.NotNil(v.DeletedAt, "deleted_at must be set after soft delete")

	// Not in list
	items, _, _, err := api.ListUserVocabs(s.userID, "", "ephemeral", 50)
	s.Require().NoError(err)
	for _, item := range items {
		s.NotEqual(vocabID, item.ID, "deleted vocab must not appear in list")
	}
}

// TestDelete_PlacementsAlsoExcludedFromPlay verifies that after vocab soft-delete, GetLevelContent
// excludes it from the synthesized envelope.
func (s *UserVocabCrudSuite) TestDelete_PlacementsAlsoExcludedFromPlay() {
	// Seed a published vocab-match game
	gameID := uuid.Must(uuid.NewV7()).String()
	g := models.Game{
		ID:       gameID,
		Name:     "vcrud_game_" + gameID[:8],
		UserID:   &s.userID,
		Mode:     consts.GameModeVocabMatch,
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

	// Seed vocab directly (bypass VIP check for the placed vocab)
	cv := models.ContentVocab{
		ID:         uuid.Must(uuid.NewV7()).String(),
		UserID:     s.userID,
		Content:    "deleteme",
		ContentKey: "deleteme",
	}
	s.Require().NoError(facades.Orm().Query().Create(&cv))

	gv := models.GameVocab{
		ID:             uuid.Must(uuid.NewV7()).String(),
		GameID:         gameID,
		GameLevelID:    levelID,
		ContentVocabID: cv.ID,
		Order:          1000,
	}
	s.Require().NoError(facades.Orm().Query().Create(&gv))

	// Before delete — vocab appears in level content
	items, err := api.GetLevelContent(s.userID, levelID, "")
	s.Require().NoError(err)
	s.Require().Len(items, 1, "before delete: one item expected")

	// Soft-delete the vocab
	s.Require().NoError(api.DeleteUserVocab(s.userID, cv.ID))

	// After delete — GetLevelContent should return empty (soft-deleted content_vocab is excluded)
	itemsAfter, err := api.GetLevelContent(s.userID, levelID, "")
	s.Require().NoError(err)
	s.Empty(itemsAfter, "after soft-delete: vocab must not appear in level content")
}
