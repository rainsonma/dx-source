package api

import (
	"fmt"
	"math/rand/v2"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"

	"dx-api/app/helpers"
	"dx-api/app/models"
)

var englishFirstNames = []string{
	"james", "emma", "oliver", "sophia", "liam", "ava", "noah", "mia",
	"lucas", "lily", "ethan", "chloe", "mason", "sarah", "logan", "emily",
	"jack", "grace", "henry", "alice", "leo", "ruby", "oscar", "ella",
	"charlie", "hannah", "max", "aria", "sam", "luna", "ben", "zoe",
}

var chineseNames = []string{
	"小明", "小红", "小华", "小丽", "小龙", "小凤", "小杰", "小雨",
	"小雪", "小云", "小星", "小月", "小天", "小海", "小风", "小林",
}

var chineseSurnames = []string{
	"wang", "li", "zhang", "liu", "chen", "yang", "huang", "wu",
}

// FindOrCreateMockUser returns an idle mock user or creates a new one.
func FindOrCreateMockUser() (*models.User, error) {
	var user models.User
	err := facades.Orm().Query().Raw(
		`SELECT u.* FROM users u
		 WHERE u.is_mock = true
		   AND NOT EXISTS (
		     SELECT 1 FROM game_pks gp
		     WHERE gp.opponent_id = u.id AND gp.is_playing = true
		   )
		 ORDER BY RANDOM() LIMIT 1`).Scan(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to query mock users: %w", err)
	}
	if user.ID != "" {
		return &user, nil
	}
	return createMockUser()
}

func createMockUser() (*models.User, error) {
	username := generateMockUsername()
	nickname := generateMockNickname()

	hashedPassword, err := helpers.HashPassword(helpers.GenerateInviteCode(16))
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := models.User{
		ID:         uuid.Must(uuid.NewV7()).String(),
		Username:   username,
		Nickname:   &nickname,
		Password:   hashedPassword,
		IsActive:   true,
		IsMock:     true,
		InviteCode: helpers.GenerateInviteCode(8),
	}

	if err := facades.Orm().Query().Create(&user); err != nil {
		return nil, fmt.Errorf("failed to create mock user: %w", err)
	}
	return &user, nil
}

func generateMockUsername() string {
	name := englishFirstNames[rand.IntN(len(englishFirstNames))]
	suffix := rand.IntN(9000) + 1000
	return fmt.Sprintf("%s%d", name, suffix)
}

func generateMockNickname() string {
	separators := []string{"-", "_", ""}
	sep := separators[rand.IntN(len(separators))]

	switch rand.IntN(4) {
	case 0:
		return englishFirstNames[rand.IntN(len(englishFirstNames))]
	case 1:
		return chineseNames[rand.IntN(len(chineseNames))]
	case 2:
		en := englishFirstNames[rand.IntN(len(englishFirstNames))]
		cn := chineseSurnames[rand.IntN(len(chineseSurnames))]
		return en + sep + cn
	default:
		cn := chineseNames[rand.IntN(len(chineseNames))]
		en := chineseSurnames[rand.IntN(len(chineseSurnames))]
		return cn + sep + en
	}
}
