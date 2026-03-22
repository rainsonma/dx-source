package helpers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const deepseekAPIURL = "https://api.deepseek.com/chat/completions"

// DeepSeekMessage represents a single chat message.
type DeepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepSeekRequest holds parameters for a chat completion call.
type DeepSeekRequest struct {
	Messages    []DeepSeekMessage
	Temperature float64
}

// deepseekPayload is the HTTP request body sent to DeepSeek.
type deepseekPayload struct {
	Model       string            `json:"model"`
	Messages    []DeepSeekMessage `json:"messages"`
	Temperature float64           `json:"temperature"`
	Stream      bool              `json:"stream"`
}

// deepseekResponse is the HTTP response from DeepSeek.
type deepseekResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

var (
	ErrDeepSeekEmpty         = errors.New("deepseek: empty response")
	ErrDeepSeekAuth          = errors.New("deepseek: invalid API key")
	ErrDeepSeekQuota         = errors.New("deepseek: quota exceeded")
	ErrDeepSeekRateLimit     = errors.New("deepseek: rate limited")
	ErrDeepSeekUnavail       = errors.New("deepseek: service unavailable")
	ErrDeepSeekNotConfigured = errors.New("deepseek: API key not configured")
)

var deepseekClient = &http.Client{Timeout: 120 * time.Second}

// CallDeepSeek calls the DeepSeek chat completion API and returns the response text.
func CallDeepSeek(req DeepSeekRequest) (string, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return "", ErrDeepSeekNotConfigured
	}

	payload := deepseekPayload{
		Model:       "deepseek-chat",
		Messages:    req.Messages,
		Temperature: req.Temperature,
		Stream:      false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("deepseek: failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, deepseekAPIURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("deepseek: failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := deepseekClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("deepseek: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", mapDeepSeekHTTPError(resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("deepseek: failed to read response: %w", err)
	}

	var dsResp deepseekResponse
	if err := json.Unmarshal(respBody, &dsResp); err != nil {
		return "", fmt.Errorf("deepseek: failed to parse response: %w", err)
	}

	if len(dsResp.Choices) == 0 || dsResp.Choices[0].Message.Content == "" {
		return "", ErrDeepSeekEmpty
	}

	return dsResp.Choices[0].Message.Content, nil
}

func mapDeepSeekHTTPError(status int) error {
	switch status {
	case 401:
		return ErrDeepSeekAuth
	case 402:
		return ErrDeepSeekQuota
	case 429:
		return ErrDeepSeekRateLimit
	default:
		return fmt.Errorf("deepseek: HTTP %d", status)
	}
}

// MapDeepSeekError maps a DeepSeek error to a user-facing Chinese message and HTTP status.
func MapDeepSeekError(err error, serviceLabel string) (string, int) {
	switch {
	case errors.Is(err, ErrDeepSeekEmpty):
		return serviceLabel + "返回为空", http.StatusBadGateway
	case errors.Is(err, ErrDeepSeekAuth):
		return serviceLabel + "密钥无效", http.StatusInternalServerError
	case errors.Is(err, ErrDeepSeekQuota):
		return serviceLabel + "额度已用完", http.StatusBadGateway
	case errors.Is(err, ErrDeepSeekRateLimit):
		return "请求过于频繁，请稍后再试", http.StatusTooManyRequests
	case errors.Is(err, ErrDeepSeekNotConfigured):
		return serviceLabel + "未配置", http.StatusInternalServerError
	default:
		return serviceLabel + "暂时不可用", http.StatusBadGateway
	}
}

// ParseAIJSONArray strips markdown code fences and parses a JSON array from AI output.
func ParseAIJSONArray(raw string) ([]json.RawMessage, error) {
	text := []byte(raw)

	// Strip markdown code fences if present
	trimmed := bytes.TrimSpace(text)
	if bytes.HasPrefix(trimmed, []byte("```")) {
		// Find end fence
		if idx := bytes.LastIndex(trimmed, []byte("```")); idx > 3 {
			inner := trimmed[3:idx]
			// Skip optional language tag on first line
			if nl := bytes.IndexByte(inner, '\n'); nl >= 0 {
				inner = inner[nl+1:]
			}
			text = bytes.TrimSpace(inner)
		}
	}

	var arr []json.RawMessage
	if err := json.Unmarshal(text, &arr); err != nil {
		return nil, fmt.Errorf("deepseek: invalid JSON array: %w", err)
	}
	return arr, nil
}

// CountWords counts English words by splitting on whitespace.
func CountWords(text string) int {
	count := 0
	inWord := false
	for _, r := range text {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			inWord = false
		} else if !inWord {
			inWord = true
			count++
		}
	}
	return count
}
