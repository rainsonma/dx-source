package api

import (
	"fmt"
	"time"

	"dx-api/app/helpers"
	"dx-api/app/models"
	"dx-api/app/services/com"

	"github.com/goravel/framework/facades"
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

// SendChangeEmailCode sends a verification code for changing email.
func SendChangeEmailCode(userID, email string) error {
	key := fmt.Sprintf("change_email_code:%s", userID)

	allowed, err := helpers.CheckRateLimit(fmt.Sprintf("rate:change_email_code:%s", userID), 1, 60)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	if !allowed {
		return ErrRateLimited
	}

	// Check email not already taken by another user
	var existing models.User
	err = facades.Orm().Query().Where("email", email).Where("id != ?", userID).First(&existing)
	if err == nil && existing.ID != "" {
		return ErrDuplicateEmail
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
