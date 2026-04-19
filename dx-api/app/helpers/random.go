package helpers

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const (
	digits       = "0123456789"
	alphanumeric = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

// GenerateCode returns a random N-digit numeric string (e.g. "482916")
func GenerateCode(length int) string {
	return randomString(length, digits)
}

// GenerateInviteCode returns a random alphanumeric invite code
func GenerateInviteCode(length int) string {
	return randomString(length, alphanumeric)
}

// GenerateDefaultNickname returns a default user nickname of the form
// "斗友_XXXXXX" where XXXXXX is 6 crypto-random digits. Uniqueness is not
// guaranteed by design: the users.nickname column has no unique constraint,
// and downstream display code accepts duplicate nicknames.
func GenerateDefaultNickname() string {
	return "斗友_" + GenerateCode(6)
}

// randomString generates a cryptographically random string from the given charset
func randomString(length int, charset string) string {
	var sb strings.Builder
	sb.Grow(length)
	for i := 0; i < length; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		sb.WriteByte(charset[n.Int64()])
	}
	return sb.String()
}
