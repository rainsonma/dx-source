package shared

import (
	"fmt"

	"github.com/goravel/framework/facades"

	"github.com/goravel/framework/contracts/mail"
)

// SendVerificationEmail sends a verification code email to the given address.
func SendVerificationEmail(to string, code string) error {
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background-color: #f5f5f5; margin: 0; padding: 20px;">
  <div style="max-width: 480px; margin: 0 auto; background: #ffffff; border-radius: 12px; padding: 40px; box-shadow: 0 2px 8px rgba(0,0,0,0.08);">
    <h2 style="color: #333; margin: 0 0 24px; font-size: 20px; text-align: center;">Douxue 验证码</h2>
    <p style="color: #666; font-size: 15px; line-height: 1.6; margin: 0 0 24px; text-align: center;">您的验证码为：</p>
    <div style="background: #f8f9fa; border-radius: 8px; padding: 20px; text-align: center; margin: 0 0 24px;">
      <span style="font-size: 32px; font-weight: bold; letter-spacing: 8px; color: #333;">%s</span>
    </div>
    <p style="color: #999; font-size: 13px; line-height: 1.5; margin: 0; text-align: center;">验证码 5 分钟内有效，请勿泄露给他人。</p>
  </div>
</body>
</html>`, code)

	err := facades.Mail().
		To([]string{to}).
		Content(mail.Content{Html: html}).
		Subject("Douxue 验证码").
		Send()
	if err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}
