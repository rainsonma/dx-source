package api

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/oklog/ulid/v2"

	"github.com/goravel/framework/facades"
	"dx-api/app/helpers"
	"dx-api/app/models"
	"dx-api/app/services/com"
)

// SendSignUpCode generates and sends a signup verification code to the given email.
func SendSignUpCode(email string) error {
	key := fmt.Sprintf("signup_code:%s", email)

	allowed, err := helpers.CheckRateLimit(fmt.Sprintf("rate:signup_code:%s", email), 1, 60)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	if !allowed {
		return ErrRateLimited
	}

	code := helpers.GenerateCode(6)
	if err := helpers.RedisSet(key, code, 300*time.Second); err != nil {
		return fmt.Errorf("failed to store verification code: %w", err)
	}

	if err := com.SendVerificationEmail(email, code); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

// AuthResult holds the tokens returned after login/signup/refresh.
type AuthResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// SignUp registers a new user with the given email, verification code, username, and password.
func SignUp(ctx contractshttp.Context, email, code, username, password string) (*AuthResult, *models.User, error) {
	// Verify code
	key := fmt.Sprintf("signup_code:%s", email)
	storedCode, err := helpers.RedisGet(key)
	if err != nil || storedCode != code {
		return nil, nil, ErrInvalidCode
	}
	_ = helpers.RedisDel(key)

	// Check duplicate email
	var existing models.User
	err = facades.Orm().Query().Where("email", email).First(&existing)
	if err == nil && existing.ID != "" {
		return nil, nil, ErrDuplicateEmail
	}

	// Derive username from email prefix if empty
	if username == "" {
		username = strings.Split(email, "@")[0]
	}

	// Check duplicate username
	err = facades.Orm().Query().Where("username", username).First(&existing)
	if err == nil && existing.ID != "" {
		return nil, nil, ErrDuplicateUsername
	}

	// Auto-generate password if empty
	if password == "" {
		password = helpers.GenerateInviteCode(16)
	}

	hashedPassword, err := helpers.HashPassword(password)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	emailStr := email
	user := models.User{
		ID:         ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
		Username:   username,
		Email:      &emailStr,
		Password:   hashedPassword,
		IsActive:   true,
		InviteCode: helpers.GenerateInviteCode(8),
	}

	if err := facades.Orm().Query().Create(&user); err != nil {
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := facades.Auth(ctx).Guard("user").Login(&user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to issue token: %w", err)
	}

	refreshToken, err := helpers.GenerateRefreshToken()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	if err := helpers.StoreRefreshToken(refreshToken, user.ID, "user"); err != nil {
		return nil, nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &AuthResult{AccessToken: token, RefreshToken: refreshToken}, &user, nil
}

// SendSignInCode generates and sends a signin verification code to the given email.
func SendSignInCode(email string) error {
	key := fmt.Sprintf("signin_code:%s", email)

	allowed, err := helpers.CheckRateLimit(fmt.Sprintf("rate:signin_code:%s", email), 1, 60)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	if !allowed {
		return ErrRateLimited
	}

	code := helpers.GenerateCode(6)
	if err := helpers.RedisSet(key, code, 300*time.Second); err != nil {
		return fmt.Errorf("failed to store verification code: %w", err)
	}

	if err := com.SendVerificationEmail(email, code); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

// SignInByEmail authenticates a user via email and verification code.
// If the user does not exist, a new account is created automatically.
func SignInByEmail(ctx contractshttp.Context, email, code string) (*AuthResult, *models.User, error) {
	// Verify code
	key := fmt.Sprintf("signin_code:%s", email)
	storedCode, err := helpers.RedisGet(key)
	if err != nil || storedCode != code {
		return nil, nil, ErrInvalidCode
	}
	_ = helpers.RedisDel(key)

	// Find user by email
	var user models.User
	err = facades.Orm().Query().Where("email", email).First(&user)
	if err != nil || user.ID == "" {
		// Auto-register
		username := strings.Split(email, "@")[0]

		// Ensure unique username
		var existingUser models.User
		if checkErr := facades.Orm().Query().Where("username", username).First(&existingUser); checkErr == nil && existingUser.ID != "" {
			username = fmt.Sprintf("%s_%s", username, helpers.GenerateCode(4))
		}

		pw := helpers.GenerateInviteCode(16)
		hashedPw, hashErr := helpers.HashPassword(pw)
		if hashErr != nil {
			return nil, nil, fmt.Errorf("failed to hash password: %w", hashErr)
		}

		emailStr := email
		user = models.User{
			ID:         ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
			Username:   username,
			Email:      &emailStr,
			Password:   hashedPw,
			IsActive:   true,
			InviteCode: helpers.GenerateInviteCode(8),
		}

		if createErr := facades.Orm().Query().Create(&user); createErr != nil {
			return nil, nil, fmt.Errorf("failed to create user: %w", createErr)
		}
	}

	token, err := facades.Auth(ctx).Guard("user").Login(&user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to issue token: %w", err)
	}

	refreshToken, err := helpers.GenerateRefreshToken()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	if err := helpers.StoreRefreshToken(refreshToken, user.ID, "user"); err != nil {
		return nil, nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &AuthResult{AccessToken: token, RefreshToken: refreshToken}, &user, nil
}

// SignInByAccount authenticates a user via account (username, email, or phone) and password.
func SignInByAccount(ctx contractshttp.Context, account, password string) (*AuthResult, *models.User, error) {
	var user models.User

	err := facades.Orm().Query().
		Where("username", account).
		OrWhere("email", account).
		OrWhere("phone", account).
		First(&user)
	if err != nil || user.ID == "" {
		return nil, nil, ErrUserNotFound
	}

	if !helpers.CheckPassword(password, user.Password) {
		return nil, nil, ErrInvalidPassword
	}

	token, err := facades.Auth(ctx).Guard("user").Login(&user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to issue token: %w", err)
	}

	refreshToken, err := helpers.GenerateRefreshToken()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	if err := helpers.StoreRefreshToken(refreshToken, user.ID, "user"); err != nil {
		return nil, nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &AuthResult{AccessToken: token, RefreshToken: refreshToken}, &user, nil
}

// RefreshToken validates an opaque refresh token, issues a new JWT access token,
// and rotates the refresh token.
func RefreshToken(ctx contractshttp.Context, oldRefreshToken string) (*AuthResult, error) {
	data, err := helpers.LookupRefreshToken(oldRefreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	if data.Guard != "user" {
		return nil, ErrInvalidRefreshToken
	}

	var user models.User
	if err := facades.Orm().Query().Where("id", data.UserID).First(&user); err != nil || user.ID == "" {
		return nil, ErrUserNotFound
	}

	accessToken, err := facades.Auth(ctx).Guard("user").Login(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to issue access token: %w", err)
	}

	_ = helpers.DeleteRefreshToken(oldRefreshToken, data.UserID, "user")

	newRefreshToken, err := helpers.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	if err := helpers.StoreRefreshToken(newRefreshToken, data.UserID, "user"); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &AuthResult{AccessToken: accessToken, RefreshToken: newRefreshToken}, nil
}

// Logout deletes the given refresh token from Redis.
func Logout(refreshToken string) error {
	data, err := helpers.LookupRefreshToken(refreshToken)
	if err != nil {
		return nil
	}
	if data.Guard != "user" {
		return nil
	}
	return helpers.DeleteRefreshToken(refreshToken, data.UserID, data.Guard)
}

// GetCurrentUser retrieves the user profile by ID (password excluded via json tag).
func GetCurrentUser(userID string) (*models.User, error) {
	var user models.User
	if err := facades.Orm().Query().Where("id", userID).First(&user); err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user.ID == "" {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

// RecordLogin creates a UserLogin record for audit purposes.
func RecordLogin(userID, ip, userAgent string) {
	agent := userAgent
	login := models.UserLogin{
		ID:     ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
		UserID: userID,
		IP:     ip,
		Agent:  &agent,
	}
	_ = facades.Orm().Query().Create(&login)
}
