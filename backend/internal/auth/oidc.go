package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCProvider handles OpenID Connect authentication
type OIDCProvider struct {
	Provider     *oidc.Provider
	Verifier     *oidc.IDTokenVerifier
	Config       oauth2.Config
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Issuer       string
}

// NewOIDCProvider creates a new OIDC provider
func NewOIDCProvider() (*OIDCProvider, error) {
	issuer := os.Getenv("OIDC_ISSUER")
	clientID := os.Getenv("OIDC_CLIENT_ID")
	clientSecret := os.Getenv("OIDC_CLIENT_SECRET")
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

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &OIDCProvider{
		Provider:     provider,
		Verifier:     verifier,
		Config:       config,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Issuer:       issuer,
	}, nil
}

// IsEnabled returns true if OIDC is configured
func (p *OIDCProvider) IsEnabled() bool {
	return p.Provider != nil
}

// GetAuthURL returns the URL for initiating OIDC authentication
func (p *OIDCProvider) GetAuthURL(state string) string {
	return p.Config.AuthCodeURL(state)
}

// ExchangeCode exchanges an authorization code for tokens
func (p *OIDCProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.Config.Exchange(ctx, code)
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
