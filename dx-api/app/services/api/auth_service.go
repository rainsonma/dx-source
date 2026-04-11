package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/models"
)

// issueSession generates a JWT via Goravel and stores login timestamp in Redis.
func issueSession(ctx contractshttp.Context, userID string) (string, error) {
	token, err := facades.Auth(ctx).Guard("user").LoginUsingID(userID)
	if err != nil {
		return "", fmt.Errorf("failed to issue token: %w", err)
	}

	// Store login timestamp for single-device enforcement
	loginTs := strconv.FormatInt(time.Now().Unix(), 10)
	ttl := time.Duration(facades.Config().GetInt("jwt.refresh_ttl", 20160)) * time.Minute
	if err := helpers.RedisSet(fmt.Sprintf("user_auth:%s:user", userID), loginTs, ttl); err != nil {
		return "", fmt.Errorf("failed to store login timestamp: %w", err)
	}

	return token, nil
}

// SignUp registers a new user with the given email, verification code, username, and password.
func SignUp(ctx contractshttp.Context, email, code, username, password string) (string, *models.User, error) {
	// Verify code
	key := fmt.Sprintf("signup_code:%s", email)
	storedCode, err := helpers.RedisGet(key)
	if err != nil || storedCode != code {
		return "", nil, ErrInvalidCode
	}
	_ = helpers.RedisDel(key)

	// Check duplicate email
	var existing models.User
	err = facades.Orm().Query().Where("email", email).First(&existing)
	if err == nil && existing.ID != "" {
		return "", nil, ErrDuplicateEmail
	}

	// Derive username from email prefix if empty
	if username == "" {
		username = strings.Split(email, "@")[0]
	}

	// Check duplicate username
	err = facades.Orm().Query().Where("username", username).First(&existing)
	if err == nil && existing.ID != "" {
		return "", nil, ErrDuplicateUsername
	}

	// Auto-generate password if empty
	if password == "" {
		password = helpers.GenerateInviteCode(16)
	}

	hashedPassword, err := helpers.HashPassword(password)
	if err != nil {
		return "", nil, fmt.Errorf("failed to hash password: %w", err)
	}

	emailStr := email
	user := models.User{
		ID:         uuid.Must(uuid.NewV7()).String(),
		Grade:      consts.UserGradeFree,
		Username:   username,
		Email:      &emailStr,
		Password:   hashedPassword,
		IsActive:   true,
		InviteCode: helpers.GenerateInviteCode(8),
	}

	if err := facades.Orm().Query().Create(&user); err != nil {
		return "", nil, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := issueSession(ctx, user.ID)
	if err != nil {
		return "", nil, err
	}
	return token, &user, nil
}

// SignInByEmail authenticates a user via email and verification code.
// If the user does not exist, a new account is created automatically.
func SignInByEmail(ctx contractshttp.Context, email, code string) (string, *models.User, error) {
	// Verify code
	key := fmt.Sprintf("signin_code:%s", email)
	storedCode, err := helpers.RedisGet(key)
	if err != nil || storedCode != code {
		return "", nil, ErrInvalidCode
	}
	_ = helpers.RedisDel(key)

	// Find user by email
	var user models.User
	err = facades.Orm().Query().Where("email", email).First(&user)
	if err != nil || user.ID == "" {
		// Auto-register
		username := strings.Split(email, "@")[0]

		var existingUser models.User
		if checkErr := facades.Orm().Query().Where("username", username).First(&existingUser); checkErr == nil && existingUser.ID != "" {
			username = fmt.Sprintf("%s_%s", username, helpers.GenerateCode(4))
		}

		pw := helpers.GenerateInviteCode(16)
		hashedPw, hashErr := helpers.HashPassword(pw)
		if hashErr != nil {
			return "", nil, fmt.Errorf("failed to hash password: %w", hashErr)
		}

		emailStr := email
		user = models.User{
			ID:         uuid.Must(uuid.NewV7()).String(),
			Grade:      consts.UserGradeFree,
			Username:   username,
			Email:      &emailStr,
			Password:   hashedPw,
			IsActive:   true,
			InviteCode: helpers.GenerateInviteCode(8),
		}

		if createErr := facades.Orm().Query().Create(&user); createErr != nil {
			return "", nil, fmt.Errorf("failed to create user: %w", createErr)
		}
	}

	token, err := issueSession(ctx, user.ID)
	if err != nil {
		return "", nil, err
	}
	return token, &user, nil
}

// SignInByAccount authenticates a user via account (username, email, or phone) and password.
func SignInByAccount(ctx contractshttp.Context, account, password string) (string, *models.User, error) {
	var user models.User

	err := facades.Orm().Query().
		Where("username", account).
		OrWhere("email", account).
		OrWhere("phone", account).
		First(&user)
	if err != nil || user.ID == "" {
		return "", nil, ErrUserNotFound
	}

	if !helpers.CheckPassword(password, user.Password) {
		return "", nil, ErrInvalidPassword
	}

	token, err := issueSession(ctx, user.ID)
	if err != nil {
		return "", nil, err
	}
	return token, &user, nil
}

// Logout blacklists the current JWT and deletes the login timestamp from Redis.
func Logout(ctx contractshttp.Context) {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err == nil && userID != "" {
		_ = facades.Auth(ctx).Guard("user").Logout()
		_ = helpers.RedisDel(fmt.Sprintf("user_auth:%s:user", userID))
	}
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
		ID:     uuid.Must(uuid.NewV7()).String(),
		UserID: userID,
		IP:     ip,
		Agent:  &agent,
	}
	_ = facades.Orm().Query().Create(&login)
}
