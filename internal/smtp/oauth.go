package smtp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ThalysSilva/unicast-backend/internal/auth"
	configenv "github.com/ThalysSilva/unicast-backend/internal/config/env"
	"github.com/ThalysSilva/unicast-backend/internal/encryption"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
)

const (
	AuthModePassword   = "password"
	AuthModeOAuth      = "oauth"
	ProviderCustomSMTP = "custom_smtp"
	ProviderGoogle     = "google"
)

type oauthState struct {
	UserID    string `json:"userId"`
	Provider  string `json:"provider"`
	ExpiresAt int64  `json:"expiresAt"`
}

type oauthTokenPayload struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type oauthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	IDToken      string `json:"id_token"`
	Scope        string `json:"scope"`
}

type oauthIDTokenClaims struct {
	Email             string `json:"email"`
	PreferredUsername string `json:"preferred_username"`
}

type oauthProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AuthURL      string
	TokenURL     string
	Scopes       []string
	Host         string
	Port         int
}

func buildOAuthProviderConfig(cfg configenv.OAuth, provider string) (*oauthProviderConfig, error) {
	switch provider {
	case ProviderGoogle:
		if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" || cfg.GoogleRedirectURL == "" {
			return nil, customerror.Make("OAuth Google não configurado", http.StatusBadRequest, fmt.Errorf("missing Google OAuth env vars"))
		}
		return &oauthProviderConfig{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:     "https://oauth2.googleapis.com/token",
			Scopes: []string{
				"openid",
				"email",
				"https://www.googleapis.com/auth/gmail.send",
			},
			Host: "smtp.gmail.com",
			Port: 587,
		}, nil
	default:
		return nil, customerror.Make("provedor OAuth inválido", http.StatusBadRequest, fmt.Errorf("invalid provider"))
	}
}

func (s *smtpService) StartOAuth(ctx context.Context, userID, provider string) (string, error) {
	providerCfg, err := buildOAuthProviderConfig(s.oauth, provider)
	if err != nil {
		return "", customerror.Trace("StartOAuth", err)
	}

	stateToken, err := auth.GenerateJWE(oauthState{
		UserID:    userID,
		Provider:  provider,
		ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
	}, s.jweSecret)
	if err != nil {
		return "", customerror.Trace("StartOAuth", err)
	}

	params := url.Values{}
	params.Set("client_id", providerCfg.ClientID)
	params.Set("redirect_uri", providerCfg.RedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", strings.Join(providerCfg.Scopes, " "))
	params.Set("state", stateToken)

	if provider == ProviderGoogle {
		params.Set("access_type", "offline")
		params.Set("include_granted_scopes", "true")
		params.Set("prompt", "consent")
	}

	return providerCfg.AuthURL + "?" + params.Encode(), nil
}

func (s *smtpService) HandleOAuthCallback(ctx context.Context, provider, code, stateToken string) (string, error) {
	redirectBase := strings.TrimRight(s.oauth.FrontendBaseURL, "/") + "/integrations"

	state, err := auth.DecryptJWE[oauthState](stateToken, s.jweSecret)
	if err != nil {
		return redirectBase, customerror.Trace("HandleOAuthCallback", err)
	}
	if state.Provider != provider || time.Now().Unix() > state.ExpiresAt {
		return redirectBase, customerror.Trace("HandleOAuthCallback", customerror.Make("estado OAuth inválido ou expirado", http.StatusBadRequest, fmt.Errorf("invalid oauth state")))
	}

	providerCfg, err := buildOAuthProviderConfig(s.oauth, provider)
	if err != nil {
		return redirectBase, customerror.Trace("HandleOAuthCallback", err)
	}

	tokenResp, err := exchangeOAuthCode(ctx, provider, providerCfg, code)
	if err != nil {
		return redirectBase, customerror.Trace("HandleOAuthCallback", err)
	}

	email, err := resolveOAuthEmail(provider, tokenResp.IDToken)
	if err != nil {
		return redirectBase, customerror.Trace("HandleOAuthCallback", err)
	}

	payloadBytes, err := json.Marshal(oauthTokenPayload{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
	})
	if err != nil {
		return redirectBase, customerror.Trace("HandleOAuthCallback", err)
	}

	encryptedPayload, iv, err := encryption.EncryptSmtpPassword(string(payloadBytes), s.jweSecret)
	if err != nil {
		return redirectBase, customerror.Trace("HandleOAuthCallback", err)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	if err := s.smtpRepository.UpsertOAuth(ctx, state.UserID, email, provider, providerCfg.Host, providerCfg.Port, encryptedPayload, iv, &expiresAt); err != nil {
		return redirectBase, customerror.Trace("HandleOAuthCallback", err)
	}

	return redirectBase + "?oauth_status=success&oauth_provider=" + url.QueryEscape(provider), nil
}

func (s *smtpService) RefreshOAuthAccessToken(ctx context.Context, instance *Instance) (string, error) {
	if instance.AuthMode != AuthModeOAuth {
		return "", customerror.Trace("RefreshOAuthAccessToken", customerror.Make("instância não usa OAuth", http.StatusBadRequest, fmt.Errorf("invalid auth mode")))
	}

	payloadJSON, err := encryption.DecryptSmtpPassword(instance.OAuthPayload, s.jweSecret, instance.OAuthIV)
	if err != nil {
		return "", customerror.Trace("RefreshOAuthAccessToken", err)
	}

	var payload oauthTokenPayload
	if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
		return "", customerror.Trace("RefreshOAuthAccessToken", err)
	}

	if payload.AccessToken != "" && instance.TokenExpiresAt != nil && time.Now().Before(instance.TokenExpiresAt.Add(-1*time.Minute)) {
		return payload.AccessToken, nil
	}

	providerCfg, err := buildOAuthProviderConfig(s.oauth, instance.Provider)
	if err != nil {
		return "", customerror.Trace("RefreshOAuthAccessToken", err)
	}

	tokenResp, err := refreshOAuthToken(ctx, instance.Provider, providerCfg, payload.RefreshToken)
	if err != nil {
		return "", customerror.Trace("RefreshOAuthAccessToken", err)
	}
	if tokenResp.RefreshToken == "" {
		tokenResp.RefreshToken = payload.RefreshToken
	}

	nextPayloadBytes, err := json.Marshal(oauthTokenPayload{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
	})
	if err != nil {
		return "", customerror.Trace("RefreshOAuthAccessToken", err)
	}

	encryptedPayload, iv, err := encryption.EncryptSmtpPassword(string(nextPayloadBytes), s.jweSecret)
	if err != nil {
		return "", customerror.Trace("RefreshOAuthAccessToken", err)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	if err := s.smtpRepository.UpdateOAuthTokens(ctx, instance.ID, encryptedPayload, iv, &expiresAt); err != nil {
		return "", customerror.Trace("RefreshOAuthAccessToken", err)
	}

	return tokenResp.AccessToken, nil
}

func exchangeOAuthCode(ctx context.Context, provider string, cfg *oauthProviderConfig, code string) (*oauthTokenResponse, error) {
	values := url.Values{}
	values.Set("client_id", cfg.ClientID)
	values.Set("client_secret", cfg.ClientSecret)
	values.Set("code", code)
	values.Set("redirect_uri", cfg.RedirectURL)
	values.Set("grant_type", "authorization_code")

	return doOAuthTokenRequest(ctx, cfg.TokenURL, values)
}

func refreshOAuthToken(ctx context.Context, provider string, cfg *oauthProviderConfig, refreshToken string) (*oauthTokenResponse, error) {
	values := url.Values{}
	values.Set("client_id", cfg.ClientID)
	values.Set("client_secret", cfg.ClientSecret)
	values.Set("refresh_token", refreshToken)
	values.Set("grant_type", "refresh_token")

	return doOAuthTokenRequest(ctx, cfg.TokenURL, values)
}

func doOAuthTokenRequest(ctx context.Context, tokenURL string, values url.Values) (*oauthTokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, customerror.Make("falha ao trocar token OAuth", http.StatusBadGateway, fmt.Errorf("%s", strings.TrimSpace(string(body))))
	}

	var tokenResp oauthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}
	return &tokenResp, nil
}

func resolveOAuthEmail(provider, idToken string) (string, error) {
	if idToken == "" {
		return "", customerror.Make("provedor OAuth não retornou identidade do usuário", http.StatusBadGateway, fmt.Errorf("missing id_token"))
	}

	parts := strings.Split(idToken, ".")
	if len(parts) < 2 {
		return "", customerror.Make("id_token inválido", http.StatusBadGateway, fmt.Errorf("invalid token format"))
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	var claims oauthIDTokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return "", err
	}

	email := claims.Email
	if email == "" {
		email = claims.PreferredUsername
	}
	if email == "" {
		return "", customerror.Make("provedor OAuth não retornou email da conta", http.StatusBadGateway, fmt.Errorf("missing email claim for provider %s", provider))
	}
	return email, nil
}
