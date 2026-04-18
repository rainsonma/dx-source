package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchWechatSession_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"openid":      "test_openid_abc123",
			"session_key": "test_session_key",
			"unionid":     "test_unionid_xyz",
			"errcode":     0,
			"errmsg":      "ok",
		})
	}))
	defer srv.Close()

	resp, err := fetchWechatSession("appid", "secret", "code", srv.URL+"?%s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.OpenID != "test_openid_abc123" {
		t.Errorf("OpenID = %q, want %q", resp.OpenID, "test_openid_abc123")
	}
	if resp.UnionID != "test_unionid_xyz" {
		t.Errorf("UnionID = %q, want %q", resp.UnionID, "test_unionid_xyz")
	}
	if resp.ErrCode != 0 {
		t.Errorf("ErrCode = %d, want 0", resp.ErrCode)
	}
}

func TestFetchWechatSession_WechatError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 40029,
			"errmsg":  "invalid code",
		})
	}))
	defer srv.Close()

	resp, err := fetchWechatSession("appid", "secret", "badcode", srv.URL+"?%s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ErrCode != 40029 {
		t.Errorf("ErrCode = %d, want 40029", resp.ErrCode)
	}
}

func TestGenerateWxUsername(t *testing.T) {
	tests := []struct {
		openID string
		want   string
	}{
		{"abcdefghijklmno", "wx_abcdefgh"},
		{"12345678xyz", "wx_12345678"},
		{"abc", "wx_abc"},
	}
	for _, tt := range tests {
		got := generateWxUsername(tt.openID)
		if got != tt.want {
			t.Errorf("generateWxUsername(%q) = %q, want %q", tt.openID, got, tt.want)
		}
	}
}
