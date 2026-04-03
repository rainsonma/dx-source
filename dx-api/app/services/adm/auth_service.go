package adm

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/helpers"
	"dx-api/app/models"
)

// issueAdminSession generates a JWT via Goravel and stores login timestamp in Redis.
func issueAdminSession(ctx contractshttp.Context, userID string) (string, error) {
	token, err := facades.Auth(ctx).Guard("admin").LoginUsingID(userID)
	if err != nil {
		return "", fmt.Errorf("failed to issue token: %w", err)
	}

	loginTs := strconv.FormatInt(time.Now().Unix(), 10)
	ttl := time.Duration(facades.Config().GetInt("jwt.refresh_ttl", 20160)) * time.Minute
	if err := helpers.RedisSet(fmt.Sprintf("user_auth:%s:admin", userID), loginTs, ttl); err != nil {
		return "", fmt.Errorf("failed to store login timestamp: %w", err)
	}

	return token, nil
}

// AdminSignIn authenticates an admin user via username and password.
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

	token, err := issueAdminSession(ctx, admUser.ID)
	if err != nil {
		return "", nil, err
	}

	ip := ctx.Request().Ip()
	userAgent := ctx.Request().Header("User-Agent", "")
	go RecordAdminLogin(admUser.ID, ip, userAgent)

	return token, &admUser, nil
}

// Logout blacklists the current JWT and deletes the login timestamp from Redis.
func Logout(ctx contractshttp.Context) {
	userID, err := facades.Auth(ctx).Guard("admin").ID()
	if err == nil && userID != "" {
		_ = facades.Auth(ctx).Guard("admin").Logout()
		_ = helpers.RedisDel(fmt.Sprintf("user_auth:%s:admin", userID))
	}
}

// GetAdminUser retrieves an admin user by ID.
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
		ID:        uuid.Must(uuid.NewV7()).String(),
		AdmUserID: admUserID,
		Ip:        ip,
		Agent:     &agent,
	}
	_ = facades.Orm().Query().Create(&login)
}
