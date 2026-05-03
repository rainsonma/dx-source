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

type ContentVocabWikiSuite struct {
	suite.Suite
	tests.TestCase
	userID      string
	adminUserID string
}

func TestContentVocabWikiSuite(t *testing.T) {
	suite.Run(t, new(ContentVocabWikiSuite))
}

func (s *ContentVocabWikiSuite) SetupTest() {
	s.userID = s.seedWikiUser()
	s.adminUserID = s.seedAdminUser()
}

func (s *ContentVocabWikiSuite) TearDownTest() {
	q := facades.Orm().Query()
	for _, uid := range []string{s.userID, s.adminUserID} {
		if uid == "" {
			continue
		}
		_, _ = q.Exec(`DELETE FROM content_vocab_edits WHERE editor_user_id = ?`, uid)
		_, _ = q.Exec(`DELETE FROM game_vocabs WHERE game_id IN (SELECT id FROM games WHERE user_id = ?)`, uid)
		_, _ = q.Exec(`DELETE FROM game_levels WHERE game_id IN (SELECT id FROM games WHERE user_id = ?)`, uid)
		_, _ = q.Exec(`DELETE FROM games WHERE user_id = ?`, uid)
		_, _ = q.Exec(`DELETE FROM users WHERE id = ?`, uid)
	}
	_, _ = q.Exec(`DELETE FROM content_vocabs WHERE content_key IN (?,?,?,?)`,
		"fast", "fast!", "斋戒", "wiki-test-word")
	s.userID = ""
	s.adminUserID = ""
}

func (s *ContentVocabWikiSuite) seedWikiUser() string {
	id := uuid.Must(uuid.NewV7()).String()
	u := models.User{
		ID:         id,
		Username:   "test_" + id,
		Grade:      consts.UserGradeLifetime,
		IsActive:   true,
		InviteCode: "w" + id[24:],
		Password:   "x",
	}
	s.Require().NoError(facades.Orm().Query().Create(&u))
	return id
}

func (s *ContentVocabWikiSuite) seedAdminUser() string {
	id := uuid.Must(uuid.NewV7()).String()
	u := models.User{
		ID:         id,
		Username:   "rainson",
		Grade:      consts.UserGradeLifetime,
		IsActive:   true,
		InviteCode: "a" + id[24:],
		Password:   "x",
	}
	s.Require().NoError(facades.Orm().Query().Create(&u))
	return id
}

func (s *ContentVocabWikiSuite) seedVocabGame() (string, string) {
	gameID := uuid.Must(uuid.NewV7()).String()
	g := models.Game{
		ID:       gameID,
		Name:     "wiki_" + gameID,
		UserID:   &s.userID,
		Mode:     consts.GameModeVocabBattle, // no batch-size constraint
		IsActive: true,
		Status:   consts.GameStatusDraft,
	}
	s.Require().NoError(facades.Orm().Query().Create(&g))

	levelID := uuid.Must(uuid.NewV7()).String()
	lv := models.GameLevel{
		ID:       levelID,
		GameID:   gameID,
		Name:     "L1",
		IsActive: true,
		Order:    1000,
	}
	s.Require().NoError(facades.Orm().Query().Create(&lv))
	return gameID, levelID
}

func (s *ContentVocabWikiSuite) countVocabs(contentKey string) int64 {
	var row struct{ N int64 }
	s.Require().NoError(facades.Orm().Query().Raw(
		`SELECT COUNT(*) AS n FROM content_vocabs WHERE content_key = ? AND deleted_at IS NULL`,
		contentKey,
	).Scan(&row))
	return row.N
}

func (s *ContentVocabWikiSuite) countEdits(vocabID, editType string) int64 {
	var row struct{ N int64 }
	s.Require().NoError(facades.Orm().Query().Raw(
		`SELECT COUNT(*) AS n FROM content_vocab_edits WHERE content_vocab_id = ? AND edit_type = ? AND deleted_at IS NULL`,
		vocabID, editType,
	).Scan(&row))
	return row.N
}

func (s *ContentVocabWikiSuite) addVocab(gameID, levelID, word string) api.AddedGameVocab {
	results, err := api.AddVocabsToLevel(s.userID, gameID, levelID, []string{word})
	s.Require().NoError(err)
	s.Require().Len(results, 1)
	return results[0]
}

// TestCreate_NewCanonical_WritesEdit verifies AddVocabsToLevel creates a content_vocabs
// row and a content_vocab_edits('create') row.
func (s *ContentVocabWikiSuite) TestCreate_NewCanonical_WritesEdit() {
	_, _ = facades.Orm().Query().Exec(`DELETE FROM content_vocabs WHERE content_key = 'fast'`)
	gameID, levelID := s.seedVocabGame()

	result := s.addVocab(gameID, levelID, "fast")

	s.Equal(int64(1), s.countVocabs("fast"))
	s.Equal(int64(1), s.countEdits(result.ContentVocabID, "create"))
	s.False(result.WasReused)
}

// TestAdd_ReusesCanonical_NoNewRow verifies case-insensitive reuse across games.
func (s *ContentVocabWikiSuite) TestAdd_ReusesCanonical_NoNewRow() {
	_, _ = facades.Orm().Query().Exec(`DELETE FROM content_vocabs WHERE content_key = 'fast'`)

	// Game A — original
	gameA, levelA := s.seedVocabGame()
	s.addVocab(gameA, levelA, "fast")
	s.Equal(int64(1), s.countVocabs("fast"))

	// Game B — reuse (different case)
	gameB, levelB := s.seedVocabGame()
	result, err := api.AddVocabsToLevel(s.userID, gameB, levelB, []string{"FAST"})
	s.Require().NoError(err)
	s.Require().Len(result, 1)

	s.True(result[0].WasReused)
	s.Equal(int64(1), s.countVocabs("fast"), "content_vocabs count must stay at 1")
}

// TestComplement_AdditiveMerge verifies complement merges new POS keys and
// drops existing ones silently.
func (s *ContentVocabWikiSuite) TestComplement_AdditiveMerge() {
	_, _ = facades.Orm().Query().Exec(`DELETE FROM content_vocabs WHERE content_key = 'fast'`)
	gameID, levelID := s.seedVocabGame()
	res := s.addVocab(gameID, levelID, "fast")
	vocabID := res.ContentVocabID

	// Seed initial definition: adj:快的
	_, err := facades.Orm().Query().Exec(
		`UPDATE content_vocabs SET definition = '[{"adj":"快的"}]' WHERE id = ?`, vocabID,
	)
	s.Require().NoError(err)

	// Complement with v:斋戒 — new POS, should be appended
	patch := api.VocabComplementPatch{
		Definition: []map[string]string{{"v": "斋戒"}},
	}
	updated, err := api.ComplementContentVocab(s.userID, vocabID, patch)
	s.Require().NoError(err)
	s.NotNil(updated.Definition)
	s.Contains(*updated.Definition, "adj")
	s.Contains(*updated.Definition, "v")

	// Complement with adj:错 — existing POS, should be silently dropped
	patch2 := api.VocabComplementPatch{
		Definition: []map[string]string{{"adj": "错"}},
	}
	updated2, err := api.ComplementContentVocab(s.userID, vocabID, patch2)
	s.Require().NoError(err)
	s.NotNil(updated2.Definition)
	s.Contains(*updated2.Definition, `"adj":"快的"`, "existing adj gloss must be preserved")
	s.NotContains(*updated2.Definition, `"adj":"错"`)
}

// TestComplement_PhoneticOnlyIfNull verifies complement sets ukPhonetic only
// when currently null and preserves existing values.
func (s *ContentVocabWikiSuite) TestComplement_PhoneticOnlyIfNull() {
	_, _ = facades.Orm().Query().Exec(`DELETE FROM content_vocabs WHERE content_key = 'fast'`)
	gameID, levelID := s.seedVocabGame()
	res := s.addVocab(gameID, levelID, "fast")
	vocabID := res.ContentVocabID

	ukPhonetic := "/fɑːst/"
	patch := api.VocabComplementPatch{UkPhonetic: &ukPhonetic}
	_, err := api.ComplementContentVocab(s.userID, vocabID, patch)
	s.Require().NoError(err)

	var v models.ContentVocab
	s.Require().NoError(facades.Orm().Query().Where("id", vocabID).First(&v))
	s.Require().NotNil(v.UkPhonetic)
	s.Equal("/fɑːst/", *v.UkPhonetic)

	// Second complement with different value — existing wins
	newPhonetic := "/different/"
	patch2 := api.VocabComplementPatch{UkPhonetic: &newPhonetic}
	_, err = api.ComplementContentVocab(s.userID, vocabID, patch2)
	s.Require().NoError(err)

	var v2 models.ContentVocab
	s.Require().NoError(facades.Orm().Query().Where("id", vocabID).First(&v2))
	s.Equal("/fɑːst/", *v2.UkPhonetic, "existing phonetic must not be overwritten")
}

// TestReplace_GatedToCreator verifies original creator can replace; a non-creator
// non-admin user is blocked once the 24h window has passed (old row).
func (s *ContentVocabWikiSuite) TestReplace_GatedToCreator() {
	_, _ = facades.Orm().Query().Exec(`DELETE FROM content_vocabs WHERE content_key = 'fast'`)
	_, _ = facades.Orm().Query().Exec(`DELETE FROM content_vocabs WHERE content_key = 'oldfastword'`)

	gameID, levelID := s.seedVocabGame()
	res := s.addVocab(gameID, levelID, "fast")
	vocabID := res.ContentVocabID

	patch := api.VocabReplacePatch{
		Content:    "fast",
		Definition: []map[string]string{{"adj": "迅速的"}},
	}

	// Creator can always replace
	_, err := api.ReplaceContentVocab(s.userID, vocabID, patch)
	s.Require().NoError(err)

	// Another user is blocked on an OLD unverified row (past 24h window)
	otherUser := s.seedWikiUser()
	defer func() {
		_, _ = facades.Orm().Query().Exec(`DELETE FROM users WHERE id = ?`, otherUser)
	}()

	oldVocabID := uuid.Must(uuid.NewV7()).String()
	_, err = facades.Orm().Query().Exec(
		`INSERT INTO content_vocabs (id, content, content_key, is_verified, created_by, created_at, updated_at)
		 VALUES (?, 'oldfastword', 'oldfastword', false, ?, now() - interval '25 hours', now())`,
		oldVocabID, s.userID,
	)
	s.Require().NoError(err)
	defer func() {
		_, _ = facades.Orm().Query().Exec(`DELETE FROM content_vocab_edits WHERE content_vocab_id = ?`, oldVocabID)
		_, _ = facades.Orm().Query().Exec(`DELETE FROM content_vocabs WHERE id = ?`, oldVocabID)
	}()

	// otherUser is neither creator nor admin, and 24h window has expired
	oldPatch := api.VocabReplacePatch{
		Content:    "oldfastword",
		Definition: []map[string]string{{"adj": "旧词"}},
	}
	_, err = api.ReplaceContentVocab(otherUser, oldVocabID, oldPatch)
	s.ErrorIs(err, api.ErrVocabNotEditable)
}

// TestReplace_GatedToAdmin verifies admin can replace any vocab.
func (s *ContentVocabWikiSuite) TestReplace_GatedToAdmin() {
	_, _ = facades.Orm().Query().Exec(`DELETE FROM content_vocabs WHERE content_key = 'fast'`)
	gameID, levelID := s.seedVocabGame()
	res := s.addVocab(gameID, levelID, "fast")
	vocabID := res.ContentVocabID

	// Set created_by to some other user (not adminUserID) so it's not the creator
	otherUser := s.seedWikiUser()
	defer func() {
		_, _ = facades.Orm().Query().Exec(`DELETE FROM users WHERE id = ?`, otherUser)
	}()
	_, _ = facades.Orm().Query().Exec(
		`UPDATE content_vocabs SET created_by = ? WHERE id = ?`, otherUser, vocabID,
	)

	patch := api.VocabReplacePatch{
		Content:    "fast",
		Definition: []map[string]string{{"adj": "迅速的"}},
	}

	// Admin can replace regardless of creator
	_, err := api.ReplaceContentVocab(s.adminUserID, vocabID, patch)
	s.Require().NoError(err)
}

// TestReplace_BlockedAfter24hAndVerified verifies that a verified row is admin-only
// and an old unverified row is gated after the 24h window.
func (s *ContentVocabWikiSuite) TestReplace_BlockedAfter24hAndVerified() {
	_, _ = facades.Orm().Query().Exec(`DELETE FROM content_vocabs WHERE content_key = 'wiki-test-word'`)

	// Use a third user as the original creator — not s.userID, not adminUserID.
	// This ensures neither "creator" nor "admin" bypasses CanReplaceVocab.
	thirdUser := s.seedWikiUser()
	defer func() {
		_, _ = facades.Orm().Query().Exec(`DELETE FROM users WHERE id = ?`, thirdUser)
	}()

	// Insert a row > 24h old, not verified, created by thirdUser
	oldVocabID := uuid.Must(uuid.NewV7()).String()
	_, err := facades.Orm().Query().Exec(
		`INSERT INTO content_vocabs (id, content, content_key, is_verified, created_by, created_at, updated_at)
		 VALUES (?, 'wiki-test-word', 'wiki-test-word', false, ?, now() - interval '25 hours', now())`,
		oldVocabID, thirdUser,
	)
	s.Require().NoError(err)
	defer func() {
		_, _ = facades.Orm().Query().Exec(`DELETE FROM content_vocab_edits WHERE content_vocab_id = ?`, oldVocabID)
		_, _ = facades.Orm().Query().Exec(`DELETE FROM content_vocabs WHERE id = ?`, oldVocabID)
	}()

	patch := api.VocabReplacePatch{
		Content:    "wiki-test-word",
		Definition: []map[string]string{{"n": "测试词"}},
	}

	// s.userID is neither creator nor admin — old unverified row (>24h) must be gated
	_, err = api.ReplaceContentVocab(s.userID, oldVocabID, patch)
	s.ErrorIs(err, api.ErrVocabNotEditable, "non-creator non-admin must be blocked after 24h")

	// Now verify it — even s.userID (non-creator) is blocked by is_verified flag
	_, err = facades.Orm().Query().Exec(
		`UPDATE content_vocabs SET is_verified = true WHERE id = ?`, oldVocabID,
	)
	s.Require().NoError(err)
	_, err = api.ReplaceContentVocab(s.userID, oldVocabID, patch)
	s.ErrorIs(err, api.ErrVocabNotEditable, "verified row must block non-admin")

	// Admin can replace verified row
	_, err = api.ReplaceContentVocab(s.adminUserID, oldVocabID, patch)
	s.Require().NoError(err)
}

// TestVerify_AdminOnly verifies non-admin gets ErrVocabAdminOnly.
func (s *ContentVocabWikiSuite) TestVerify_AdminOnly() {
	_, _ = facades.Orm().Query().Exec(`DELETE FROM content_vocabs WHERE content_key = 'fast'`)
	gameID, levelID := s.seedVocabGame()
	res := s.addVocab(gameID, levelID, "fast")

	_, err := api.VerifyContentVocab(s.userID, res.ContentVocabID, true)
	s.ErrorIs(err, api.ErrVocabAdminOnly)
}

// TestVerify_LocksFutureReplaces verifies that after admin verify, a non-creator
// non-admin user is blocked (is_verified=true removes the open-window path).
// Note: the creator retains replace rights per CanReplaceVocab semantics.
func (s *ContentVocabWikiSuite) TestVerify_LocksFutureReplaces() {
	_, _ = facades.Orm().Query().Exec(`DELETE FROM content_vocabs WHERE content_key = 'fast'`)
	gameID, levelID := s.seedVocabGame()
	res := s.addVocab(gameID, levelID, "fast")
	vocabID := res.ContentVocabID

	_, err := api.VerifyContentVocab(s.adminUserID, vocabID, true)
	s.Require().NoError(err)

	// Creator can still replace even after verify (CanReplaceVocab: creator always allowed)
	patch := api.VocabReplacePatch{
		Content:    "fast",
		Definition: []map[string]string{{"adj": "快速的"}},
	}
	_, err = api.ReplaceContentVocab(s.userID, vocabID, patch)
	s.Require().NoError(err, "creator retains replace right even after verify")

	// Non-creator non-admin is blocked by is_verified (no open-window path)
	otherUser := s.seedWikiUser()
	defer func() {
		_, _ = facades.Orm().Query().Exec(`DELETE FROM users WHERE id = ?`, otherUser)
	}()
	_, err = api.ReplaceContentVocab(otherUser, vocabID, patch)
	s.ErrorIs(err, api.ErrVocabNotEditable, "verified vocab must block non-creator non-admin")
}

// TestEdits_AppendOnEachOp verifies each operation writes one content_vocab_edits row.
func (s *ContentVocabWikiSuite) TestEdits_AppendOnEachOp() {
	_, _ = facades.Orm().Query().Exec(`DELETE FROM content_vocabs WHERE content_key = 'fast'`)
	gameID, levelID := s.seedVocabGame()
	res := s.addVocab(gameID, levelID, "fast")
	vocabID := res.ContentVocabID

	// create edit already written by AddVocabsToLevel
	s.Equal(int64(1), s.countEdits(vocabID, "create"))

	// complement
	uk := "/fɑːst/"
	_, err := api.ComplementContentVocab(s.userID, vocabID, api.VocabComplementPatch{UkPhonetic: &uk})
	s.Require().NoError(err)
	s.Equal(int64(1), s.countEdits(vocabID, "complement"))

	// replace
	_, err = api.ReplaceContentVocab(s.userID, vocabID, api.VocabReplacePatch{
		Content:    "fast",
		Definition: []map[string]string{{"adj": "迅速的"}},
	})
	s.Require().NoError(err)
	s.Equal(int64(1), s.countEdits(vocabID, "replace"))

	// verify
	_, err = api.VerifyContentVocab(s.adminUserID, vocabID, true)
	s.Require().NoError(err)
	s.Equal(int64(1), s.countEdits(vocabID, "verify"))
}
