package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"rca.agent/test/internal/config"
	"rca.agent/test/internal/httputil"
)

// OAuthTokenResponse represents the OAuth2 token response
type OAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// OAuthTokenManager manages OAuth2 tokens with caching and refresh
type OAuthTokenManager struct {
	mu         sync.RWMutex
	cfg        *config.Config
	token      string
	expiresAt  time.Time
	httpClient *http.Client
}

// NewOAuthTokenManager creates a new token manager
func NewOAuthTokenManager(cfg *config.Config) *OAuthTokenManager {
	return &OAuthTokenManager{
		cfg:        cfg,
		httpClient: httputil.NewHTTPClient(30*time.Second, cfg.TLSInsecureSkipVerify),
	}
}

// GetToken returns a valid token, fetching a new one if necessary
func (m *OAuthTokenManager) GetToken(ctx context.Context) (string, error) {
	m.mu.RLock()
	if m.token != "" && time.Now().Before(m.expiresAt.Add(-30*time.Second)) {
		token := m.token
		m.mu.RUnlock()
		return token, nil
	}
	m.mu.RUnlock()

	return m.refreshToken(ctx)
}

func (m *OAuthTokenManager) refreshToken(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring lock
	if m.token != "" && time.Now().Before(m.expiresAt.Add(-30*time.Second)) {
		return m.token, nil
	}

	slog.Debug("Fetching OAuth token", "url", m.cfg.OAuthTokenURL)

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", m.cfg.OAuthClientID)
	data.Set("client_secret", m.cfg.OAuthClientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", m.cfg.OAuthTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp OAuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	m.token = tokenResp.AccessToken
	m.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	slog.Debug("OAuth token acquired", "expires_in", tokenResp.ExpiresIn)

	return m.token, nil
}

// FetchOAuthToken is a simple helper to fetch a token once
func FetchOAuthToken(ctx context.Context, cfg *config.Config) (string, error) {
	if !cfg.IsOAuthConfigured() {
		return "", fmt.Errorf("oauth not configured")
	}

	manager := NewOAuthTokenManager(cfg)
	return manager.GetToken(ctx)
}
