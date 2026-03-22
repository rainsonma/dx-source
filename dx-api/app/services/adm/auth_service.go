package adm

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/helpers"
	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

type AuthResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func issueAdminSession(userID string) (*AuthResult, error) {
	authID := uuid.Must(uuid.NewV7()).String()

	ttl := time.Duration(facades.Config().GetInt("refresh_token.ttl", 10080)) * time.Minute
	if err := helpers.RedisSet(fmt.Sprintf("user_auth:%s:admin", userID), authID, ttl); err != nil {
		return nil, fmt.Errorf("failed to store auth_id: %w", err)
	}

	_ = helpers.DeleteUserRefreshTokens(userID, "admin")

	accessToken, err := helpers.IssueAccessToken(userID, authID)
	if err != nil {
		return nil, fmt.Errorf("failed to issue access token: %w", err)
	}

	refreshToken, err := helpers.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	if err := helpers.StoreRefreshToken(refreshToken, userID, "admin", authID); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &AuthResult{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

// AdminSignIn authenticates an admin user via username and password.
// It verifies credentials, issues a JWT token using the "admin" guard,
// and records the login for audit purposes.
func AdminSignIn(ctx contractshttp.Context, username, password string) (*AuthResult, *models.AdmUser, error) {
	var admUser models.AdmUser
	err := facades.Orm().Query().Where("username", username).First(&admUser)
	if err != nil || admUser.ID == "" {
		return nil, nil, ErrAdminNotFound
	}

	if !admUser.IsActive {
		return nil, nil, ErrAdminInactive
	}

	if !helpers.CheckPassword(password, admUser.Password) {
		return nil, nil, ErrInvalidPassword
	}

	result, err := issueAdminSession(admUser.ID)
	if err != nil {
		return nil, nil, err
	}

	ip := ctx.Request().Ip()
	userAgent := ctx.Request().Header("User-Agent", "")
	go RecordAdminLogin(admUser.ID, ip, userAgent)

	return result, &admUser, nil
}

// RefreshToken rotates the admin refresh token and issues a new access token.
func RefreshToken(ctx contractshttp.Context, oldRefreshToken string) (*AuthResult, error) {
	data, err := helpers.LookupRefreshToken(oldRefreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}
	if data.Guard != "admin" {
		return nil, ErrInvalidRefreshToken
	}

	// Check if this session is still the active one
	currentAuthID, err := helpers.RedisGet(fmt.Sprintf("user_auth:%s:admin", data.UserID))
	if err != nil || currentAuthID != data.AuthID {
		return nil, ErrSessionReplaced
	}

	var admUser models.AdmUser
	if err := facades.Orm().Query().Where("id", data.UserID).First(&admUser); err != nil || admUser.ID == "" {
		return nil, ErrAdminNotFound
	}

	accessToken, err := helpers.IssueAccessToken(data.UserID, data.AuthID)
	if err != nil {
		return nil, fmt.Errorf("failed to issue access token: %w", err)
	}

	_ = helpers.DeleteRefreshToken(oldRefreshToken, data.UserID, "admin")

	newRefreshToken, err := helpers.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	if err := helpers.StoreRefreshToken(newRefreshToken, data.UserID, "admin", data.AuthID); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &AuthResult{AccessToken: accessToken, RefreshToken: newRefreshToken}, nil
}

// Logout invalidates the admin refresh token.
func Logout(refreshToken string) error {
	data, err := helpers.LookupRefreshToken(refreshToken)
	if err != nil {
		return nil
	}
	if data.Guard != "admin" {
		return nil
	}
	key := fmt.Sprintf("user_auth:%s:%s", data.UserID, data.Guard)
	currentAuthID, _ := helpers.RedisGet(key)
	if currentAuthID == data.AuthID {
		_ = helpers.RedisDel(key)
	}
	return helpers.DeleteRefreshToken(refreshToken, data.UserID, data.Guard)
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
		ID:        uuid.Must(uuid.NewV7()).String(),
		AdmUserID: admUserID,
		Ip:        ip,
		Agent:     &agent,
	}
	_ = facades.Orm().Query().Create(&login)
}
