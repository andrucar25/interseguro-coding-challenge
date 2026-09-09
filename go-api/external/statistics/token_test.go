package statistics

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestTokenSignerSignsRequiredHS256Claims(t *testing.T) {
	before := time.Now().UTC()
	signer, err := NewTokenSigner(testJWTSecret, testJWTIssuer, testJWTAudience)
	if err != nil {
		t.Fatalf("NewTokenSigner() error = %v", err)
	}

	rawToken, err := signer.Sign()
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	after := time.Now().UTC()

	claims := parseServiceToken(t, rawToken)
	if claims.Issuer != testJWTIssuer {
		t.Errorf("issuer = %q, want %q", claims.Issuer, testJWTIssuer)
	}
	if !containsAudience(claims.Audience, testJWTAudience) {
		t.Errorf("audience = %v, want %q", claims.Audience, testJWTAudience)
	}
	if claims.IssuedAt == nil {
		t.Fatal("issued-at claim is missing")
	}
	if claims.ExpiresAt == nil {
		t.Fatal("expiration claim is missing")
	}
	if claims.IssuedAt.Time.Before(before.Add(-time.Second)) || claims.IssuedAt.Time.After(after.Add(time.Second)) {
		t.Errorf("issued-at = %s, want between %s and %s", claims.IssuedAt.Time, before, after)
	}
	if lifetime := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time); lifetime != serviceTokenLifetime {
		t.Errorf("lifetime = %s, want %s", lifetime, serviceTokenLifetime)
	}
}

func containsAudience(audiences jwt.ClaimStrings, target string) bool {
	for _, audience := range audiences {
		if audience == target {
			return true
		}
	}
	return false
}

func TestNewTokenSignerRejectsInvalidConfiguration(t *testing.T) {
	testCases := []struct {
		name     string
		secret   string
		issuer   string
		audience string
	}{
		{name: "empty secret", secret: "", issuer: testJWTIssuer, audience: testJWTAudience},
		{name: "blank secret", secret: "  ", issuer: testJWTIssuer, audience: testJWTAudience},
		{name: "empty issuer", secret: testJWTSecret, issuer: "", audience: testJWTAudience},
		{name: "empty audience", secret: testJWTSecret, issuer: testJWTIssuer, audience: ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := NewTokenSigner(testCase.secret, testCase.issuer, testCase.audience); err == nil {
				t.Fatal("NewTokenSigner() error = nil, want error")
			}
		})
	}

	if _, err := newTokenSigner(testJWTSecret, testJWTIssuer, testJWTAudience, nil); err == nil {
		t.Fatal("newTokenSigner() error = nil, want error for nil clock")
	}
}

func parseServiceToken(t *testing.T, rawToken string) *jwt.RegisteredClaims {
	t.Helper()
	claims := &jwt.RegisteredClaims{}
	parsedToken, err := jwt.ParseWithClaims(
		rawToken,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errors.New("unexpected signing algorithm")
			}
			return []byte(testJWTSecret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(testJWTIssuer),
		jwt.WithAudience(testJWTAudience),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		t.Fatalf("ParseWithClaims() error = %v", err)
	}
	if !parsedToken.Valid {
		t.Fatal("token is invalid")
	}
	return claims
}
