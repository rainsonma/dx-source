package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/models"
)

// wechatSessionURL is the production jscode2session endpoint.
// The single %s is replaced with the encoded query string.
const wechatSessionURL = "https://api.weixin.qq.com/sns/jscode2session?%s"

var wechatHTTPClient = &http.Client{Timeout: 5 * time.Second}

type wechatSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// fetchWechatSession calls the WeChat jscode2session endpoint.
// urlFmt must contain a single %s placeholder for the query string.
// Use wechatSessionURL for production; inject a test server URL in tests.
func fetchWechatSession(appID, secret, code, urlFmt string) (*wechatSessionResponse, error) {
	q := url.Values{}
	q.Set("appid", appID)
	q.Set("secret", secret)
	q.Set("js_code", code)
	q.Set("grant_type", "authorization_code")
	endpoint := fmt.Sprintf(urlFmt, q.Encode())
	resp, err := wechatHTTPClient.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("wechat session request failed: %w", err)
	}
	defer resp.Body.Close()

	var result wechatSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode wechat session response: %w", err)
	}
	return &result, nil
}

// generateWxUsername returns "wx_" + first 8 chars of openID (or full openID if shorter).
func generateWxUsername(openID string) string {
	suffix := openID
	if len(openID) > 8 {
		suffix = openID[:8]
	}
	return "wx_" + suffix
}

// WechatMiniSignIn exchanges a wx.login code for a JWT, registering new users automatically.
func WechatMiniSignIn(ctx contractshttp.Context, code string) (string, *models.User, error) {
	appID := facades.Config().GetString("wechat.mini_app_id")
	secret := facades.Config().GetString("wechat.mini_app_secret")

	session, err := fetchWechatSession(appID, secret, code, wechatSessionURL)
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch wechat session: %w", err)
	}
	if session.ErrCode != 0 {
		return "", nil, fmt.Errorf("wechat error %d: %s", session.ErrCode, session.ErrMsg)
	}
	if session.OpenID == "" {
		return "", nil, fmt.Errorf("wechat returned empty openid")
	}

	// Find existing user by openid
	var user models.User
	err = facades.Orm().Query().Where("openid", session.OpenID).First(&user)
	if err != nil || user.ID == "" {
		// Auto-register new user
		username := generateWxUsername(session.OpenID)

		// Ensure username uniqueness
		var existing models.User
		if checkErr := facades.Orm().Query().Where("username", username).First(&existing); checkErr == nil && existing.ID != "" {
			username = fmt.Sprintf("%s_%s", username, helpers.GenerateCode(4))
		}

		pw := helpers.GenerateInviteCode(16)
		hashedPw, hashErr := helpers.HashPassword(pw)
		if hashErr != nil {
			return "", nil, fmt.Errorf("failed to hash password: %w", hashErr)
		}

		openID := session.OpenID
		nickname := helpers.GenerateDefaultNickname()
		user = models.User{
			ID:         uuid.Must(uuid.NewV7()).String(),
			Grade:      consts.UserGradeFree,
			Username:   username,
			Nickname:   &nickname,
			Password:   hashedPw,
			IsActive:   true,
			InviteCode: helpers.GenerateInviteCode(8),
			OpenID:     &openID,
		}
		if session.UnionID != "" {
			unionID := session.UnionID
			user.UnionID = &unionID
		}

		if createErr := facades.Orm().Query().Create(&user); createErr != nil {
			return "", nil, fmt.Errorf("failed to create user: %w", createErr)
		}

		if refErr := RecordReferralIfPresent(ctx, user.ID); refErr != nil {
			facades.Log().Warningf("record referral failed: %v", refErr)
		}
	}

	token, err := issueSession(ctx, user.ID)
	if err != nil {
		return "", nil, err
	}
	return token, &user, nil
}
