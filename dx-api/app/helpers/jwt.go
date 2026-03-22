package helpers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/goravel/framework/facades"
)

// Claims defines the JWT payload. The Key field is read by Goravel's
// Auth().Guard().ID() for backward compatibility.
type Claims struct {
	jwt.RegisteredClaims
	Key    string `json:"key"`     // user ID (Goravel compatible)
	AuthID string `json:"auth_id"` // session identifier for single-device enforcement
}

// IssueAccessToken creates a signed JWT with the user ID and auth_id claims.
func IssueAccessToken(userID, authID string) (string, error) {
	secret := facades.Config().GetString("jwt.secret", "")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET is not configured")
	}

	ttl := facades.Config().GetInt("jwt.ttl", 10)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(ttl) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Key:    userID,
		AuthID: authID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ExtractAuthID reads the auth_id claim from a Bearer token without re-verifying
// the signature. Only call AFTER Goravel's Parse() has validated the token.
func ExtractAuthID(bearerToken string) string {
	raw := strings.TrimPrefix(bearerToken, "Bearer ")
	parts := strings.SplitN(raw, ".", 3)
	if len(parts) != 3 {
		return ""
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims struct {
		AuthID string `json:"auth_id"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return ""
	}
	return claims.AuthID
}
