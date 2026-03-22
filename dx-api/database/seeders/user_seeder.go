package seeders

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/goravel/framework/facades"
	"dx-api/app/helpers"
	"dx-api/app/models"
)

type UserSeeder struct{}

func (s *UserSeeder) Signature() string {
	return "UserSeeder"
}

func (s *UserSeeder) Run() error {
	hashedPw, err := helpers.HashPassword("Password123")
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := facades.Orm().Query()

	namedUsers := []struct {
		Username string
		Nickname string
		Grade    string
		Email    string
	}{
		{"rainson", "Rainson", "lifetime", "rainsonma@gmail.com"},
		{"june", "June", "lifetime", ""},
	}

	for _, u := range namedUsers {
		nickname := u.Nickname
		var email *string
		if u.Email != "" {
			email = &u.Email
		}

		var existing models.User
		if err := query.Where("username", u.Username).First(&existing); err != nil || existing.ID == "" {
			if err := query.Create(&models.User{
				ID:         ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
				Username:   u.Username,
				Nickname:   &nickname,
				Grade:      u.Grade,
				Email:      email,
				Password:   hashedPw,
				InviteCode: helpers.GenerateInviteCode(8),
				IsActive:   true,
			}); err != nil {
				return fmt.Errorf("failed to create user %s: %w", u.Username, err)
			}
		} else {
			if _, err := query.Where("username", u.Username).Update(&models.User{
				Nickname: &nickname,
				Grade:    u.Grade,
				Email:    email,
				Password: hashedPw,
				IsActive: true,
			}); err != nil {
				return fmt.Errorf("failed to update user %s: %w", u.Username, err)
			}
		}
	}

	for i := 3; i <= 100; i++ {
		username := fmt.Sprintf("user%03d", i)
		nickname := fmt.Sprintf("用户%03d", i)

		var existing models.User
		if err := query.Where("username", username).First(&existing); err != nil || existing.ID == "" {
			if err := query.Create(&models.User{
				ID:         ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
				Username:   username,
				Nickname:   &nickname,
				Password:   hashedPw,
				InviteCode: helpers.GenerateInviteCode(8),
				IsActive:   true,
			}); err != nil {
				return fmt.Errorf("failed to create user %s: %w", username, err)
			}
		} else {
			if _, err := query.Where("username", username).Update(&models.User{
				Nickname: &nickname,
				Password: hashedPw,
				IsActive: true,
			}); err != nil {
				return fmt.Errorf("failed to update user %s: %w", username, err)
			}
		}
	}

	log.Println("Seeded 100 users")
	return nil
}
