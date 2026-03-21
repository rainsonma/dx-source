package adm

import (
	"crypto/rand"
	"fmt"
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/oklog/ulid/v2"

	"dx-api/app/facades"
	"dx-api/app/helpers"
	"dx-api/app/models"
)

// AdminSignIn authenticates an admin user via username and password.
// It verifies credentials, issues a JWT token using the "admin" guard,
// and records the login for audit purposes.
func AdminSignIn(ctx contractshttp.Context, username, password string) (string, *models.AdmUser, error) {
	var admUser models.AdmUser
	err := facades.Orm().Query().Where("username", username).First(&admUser)
	if err != nil || admUser.ID == "" {
		return "", nil, ErrAdminNotFound
	}

	if !admUser.IsActive {
		return "", nil, ErrAdminInactive
	}

	if !helpers.CheckPassword(password, admUser.Password) {
		return "", nil, ErrInvalidPassword
	}

	token, err := facades.Auth(ctx).Guard("admin").Login(&admUser)
	if err != nil {
		return "", nil, fmt.Errorf("failed to issue admin token: %w", err)
	}

	ip := ctx.Request().Ip()
	userAgent := ctx.Request().Header("User-Agent", "")
	go RecordAdminLogin(admUser.ID, ip, userAgent)

	return token, &admUser, nil
}

// GetAdminUser retrieves an admin user by ID.
// The password field is excluded from JSON output via the model's json:"-" tag.
func GetAdminUser(userID string) (*models.AdmUser, error) {
	var admUser models.AdmUser
	if err := facades.Orm().Query().Where("id", userID).First(&admUser); err != nil {
		return nil, fmt.Errorf("failed to find admin user: %w", err)
	}
	if admUser.ID == "" {
		return nil, ErrAdminNotFound
	}
	return &admUser, nil
}

// RecordAdminLogin creates an AdmLogin record for audit purposes.
func RecordAdminLogin(admUserID, ip, userAgent string) {
	agent := userAgent
	login := models.AdmLogin{
		ID:        ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
		AdmUserID: admUserID,
		Ip:        ip,
		Agent:     &agent,
	}
	_ = facades.Orm().Query().Create(&login)
}
