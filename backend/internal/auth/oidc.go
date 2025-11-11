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

	// Get OIDC scopes from environment or use defaults
	oidcScopes := os.Getenv("OIDC_SCOPES")
	if oidcScopes == "" {
		oidcScopes = "openid,email,groups"
	}
	
	// Parse scopes from comma-separated string and convert to slice
	scopes := []string{}
	for _, scope := range splitAndTrim(oidcScopes, ",") {
		if scope != "" {
			scopes = append(scopes, scope)
		}
	}
	
	// Ensure openid is always included
	hasOpenID := false
	for _, s := range scopes {
		if s == "openid" {
			hasOpenID = true
			break
		}
	}
	if !hasOpenID {
		scopes = append([]string{"openid"}, scopes...)
	}

	// Configure for public client (no client secret required)
	config := oauth2.Config{
		ClientID:    clientID,
		RedirectURL: redirectURL,
		Endpoint:    provider.Endpoint(),
		Scopes:      scopes,
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
	Sub           string   `json:"sub"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Name          string   `json:"name"`
	Picture       string   `json:"picture"`
	Groups        []string `json:"groups"` // Groups from OIDC token
}

// ExtractClaims extracts claims from an ID token
func ExtractClaims(idToken *oidc.IDToken) (*Claims, error) {
	// Get groups claim name from environment or use default
	groupsClaimName := os.Getenv("OIDC_GROUPS_CLAIM")
	if groupsClaimName == "" {
		groupsClaimName = "groups"
	}

	// Extract all claims first
	var allClaims map[string]interface{}
	if err := idToken.Claims(&allClaims); err != nil {
		return nil, fmt.Errorf("failed to extract claims: %w", err)
	}

	// Extract standard claims
	var claims Claims
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to extract standard claims: %w", err)
	}

	// Extract groups from the configured claim
	if groupsInterface, ok := allClaims[groupsClaimName]; ok {
		switch groups := groupsInterface.(type) {
		case []interface{}:
			for _, g := range groups {
				if groupStr, ok := g.(string); ok {
					claims.Groups = append(claims.Groups, groupStr)
				}
			}
		case []string:
			claims.Groups = groups
		case string:
			// Single group as string
			claims.Groups = []string{groups}
		}
	}

	return &claims, nil
}

// splitAndTrim splits a string by delimiter and trims spaces
func splitAndTrim(s string, delimiter string) []string {
	parts := []string{}
	for _, part := range splitString(s, delimiter) {
		trimmed := trimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

// splitString splits a string by delimiter
func splitString(s string, delimiter string) []string {
	if s == "" {
		return []string{}
	}
	result := []string{}
	current := ""
	for i := 0; i < len(s); i++ {
		if i+len(delimiter) <= len(s) && s[i:i+len(delimiter)] == delimiter {
			result = append(result, current)
			current = ""
			i += len(delimiter) - 1
		} else {
			current += string(s[i])
		}
	}
	result = append(result, current)
	return result
}

// trimSpace removes leading and trailing whitespace
func trimSpace(s string) string {
	start := 0
	end := len(s)
	
	for start < end && isSpace(s[start]) {
		start++
	}
	
	for end > start && isSpace(s[end-1]) {
		end--
	}
	
	return s[start:end]
}

// isSpace checks if a byte is a whitespace character
func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}
