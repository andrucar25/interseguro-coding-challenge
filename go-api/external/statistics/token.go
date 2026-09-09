package statistics

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const serviceTokenLifetime = time.Minute

// TokenSigner creates short-lived credentials for the Node statistics API.
type TokenSigner struct {
	secret   []byte
	issuer   string
	audience string
	now      func() time.Time
}

// NewTokenSigner constructs an HS256 signer with the required JWT claims.
func NewTokenSigner(secret, issuer, audience string) (*TokenSigner, error) {
	return newTokenSigner(secret, issuer, audience, time.Now)
}

func newTokenSigner(secret, issuer, audience string, now func() time.Time) (*TokenSigner, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("JWT secret must not be empty")
	}
	if strings.TrimSpace(issuer) == "" {
		return nil, errors.New("JWT issuer must not be empty")
	}
	if strings.TrimSpace(audience) == "" {
		return nil, errors.New("JWT audience must not be empty")
	}
	if now == nil {
		return nil, errors.New("JWT clock must not be nil")
	}

	return &TokenSigner{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
		now:      now,
	}, nil
}

// Sign creates a one-minute HS256 JWT for a Node statistics request.
func (s *TokenSigner) Sign() (string, error) {
	issuedAt := s.now().UTC()
	claims := jwt.RegisteredClaims{
		Issuer:    s.issuer,
		Audience:  jwt.ClaimStrings{s.audience},
		IssuedAt:  jwt.NewNumericDate(issuedAt),
		ExpiresAt: jwt.NewNumericDate(issuedAt.Add(serviceTokenLifetime)),
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}
