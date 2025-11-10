package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCProvider handles OpenID Connect authentication
type OIDCProvider struct {
	Provider    *oidc.Provider
	Verifier    *oidc.IDTokenVerifier
	Config      oauth2.Config
	ClientID    string
	RedirectURL string
	Issuer      string
}

// PKCEData contains PKCE verifier and challenge
type PKCEData struct {
	Verifier  string
	Challenge string
}

// NewOIDCProvider creates a new OIDC provider
func NewOIDCProvider() (*OIDCProvider, error) {
	issuer := os.Getenv("OIDC_ISSUER")
	clientID := os.Getenv("OIDC_CLIENT_ID")
	redirectURL := os.Getenv("OIDC_REDIRECT_URL")

	// OIDC is optional, can be disabled
	if issuer == "" || clientID == "" {
		return &OIDCProvider{}, nil
	}

	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	// Configure for public client (no client secret required)
	config := oauth2.Config{
		ClientID:    clientID,
		RedirectURL: redirectURL,
		Endpoint:    provider.Endpoint(),
		Scopes:      []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &OIDCProvider{
		Provider:    provider,
		Verifier:    verifier,
		Config:      config,
		ClientID:    clientID,
		RedirectURL: redirectURL,
		Issuer:      issuer,
	}, nil
}

// GeneratePKCE generates PKCE verifier and challenge for public clients
func GeneratePKCE() (*PKCEData, error) {
	// Generate verifier (43-128 characters)
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		return nil, err
	}
	verifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	// Generate challenge (SHA256 of verifier)
	h := sha256.New()
	h.Write([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return &PKCEData{
		Verifier:  verifier,
		Challenge: challenge,
	}, nil
}

// IsEnabled returns true if OIDC is configured
func (p *OIDCProvider) IsEnabled() bool {
	return p.Provider != nil
}

// GetAuthURL returns the URL for initiating OIDC authentication with PKCE
func (p *OIDCProvider) GetAuthURL(state string, pkceChallenge string) string {
	return p.Config.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", pkceChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
}

// ExchangeCode exchanges an authorization code for tokens with PKCE
func (p *OIDCProvider) ExchangeCode(ctx context.Context, code string, pkceVerifier string) (*oauth2.Token, error) {
	return p.Config.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", pkceVerifier),
	)
}

// VerifyIDToken verifies an ID token
func (p *OIDCProvider) VerifyIDToken(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	return p.Verifier.Verify(ctx, rawIDToken)
}

// Claims represents the claims extracted from an ID token
type Claims struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// ExtractClaims extracts claims from an ID token
func ExtractClaims(idToken *oidc.IDToken) (*Claims, error) {
	var claims Claims
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to extract claims: %w", err)
	}
	return &claims, nil
}
