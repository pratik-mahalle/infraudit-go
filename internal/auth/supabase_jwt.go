package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"

	"github.com/golang-jwt/jwt/v5"
)

// SupabaseClaims represents the JWT claims issued by Supabase Auth
type SupabaseClaims struct {
	Sub   string `json:"sub"`   // User UUID from auth.users
	Email string `json:"email"` // User email
	Role  string `json:"role"`  // "authenticated" for logged-in users, "anon" for anonymous
	jwt.RegisteredClaims
}

// JWKSKeyFunc provides a jwt.Keyfunc that supports both HS256 and ES256.
// For ES256, it fetches and caches the JWKS from the Supabase Auth endpoint.
type JWKSKeyFunc struct {
	supabaseURL string
	hmacSecret  string

	mu      sync.RWMutex
	ecKeys  map[string]*ecdsa.PublicKey // kid -> public key
	fetched bool
}

// NewJWKSKeyFunc creates a key function that verifies Supabase JWTs.
// It supports HS256 (using hmacSecret) and ES256 (fetching JWKS from supabaseURL).
func NewJWKSKeyFunc(supabaseURL, hmacSecret string) *JWKSKeyFunc {
	return &JWKSKeyFunc{
		supabaseURL: supabaseURL,
		hmacSecret:  hmacSecret,
		ecKeys:      make(map[string]*ecdsa.PublicKey),
	}
}

// Keyfunc returns the jwt.Keyfunc to use with jwt.ParseWithClaims.
func (j *JWKSKeyFunc) Keyfunc(t *jwt.Token) (interface{}, error) {
	switch t.Method.Alg() {
	case "HS256":
		return []byte(j.hmacSecret), nil

	case "ES256":
		kid, _ := t.Header["kid"].(string)
		key, err := j.getECKey(kid)
		if err != nil {
			return nil, fmt.Errorf("failed to get EC key: %w", err)
		}
		return key, nil

	default:
		return nil, fmt.Errorf("unexpected signing method: %s", t.Method.Alg())
	}
}

// getECKey returns the ECDSA public key for the given kid, fetching JWKS if needed.
func (j *JWKSKeyFunc) getECKey(kid string) (*ecdsa.PublicKey, error) {
	// Try cache first
	j.mu.RLock()
	if key, ok := j.ecKeys[kid]; ok {
		j.mu.RUnlock()
		return key, nil
	}
	fetched := j.fetched
	j.mu.RUnlock()

	// If already fetched and kid not found, try re-fetching once (key rotation)
	if fetched {
		if err := j.fetchJWKS(); err != nil {
			return nil, err
		}
		j.mu.RLock()
		key, ok := j.ecKeys[kid]
		j.mu.RUnlock()
		if !ok {
			return nil, fmt.Errorf("key %q not found in JWKS", kid)
		}
		return key, nil
	}

	// First fetch
	if err := j.fetchJWKS(); err != nil {
		return nil, err
	}

	j.mu.RLock()
	defer j.mu.RUnlock()

	if kid == "" {
		// No kid in token; return the first key
		for _, key := range j.ecKeys {
			return key, nil
		}
		return nil, fmt.Errorf("no EC keys in JWKS")
	}

	key, ok := j.ecKeys[kid]
	if !ok {
		return nil, fmt.Errorf("key %q not found in JWKS", kid)
	}
	return key, nil
}

// jwksResponse is the JSON structure of the JWKS endpoint
type jwksResponse struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
}

// fetchJWKS fetches the JWKS from the Supabase Auth endpoint
func (j *JWKSKeyFunc) fetchJWKS() error {
	url := j.supabaseURL + "/auth/v1/.well-known/jwks.json"
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS response: %w", err)
	}

	j.mu.Lock()
	defer j.mu.Unlock()

	for _, key := range jwks.Keys {
		if key.Kty != "EC" || key.Crv != "P-256" {
			continue
		}

		xBytes, err := base64.RawURLEncoding.DecodeString(key.X)
		if err != nil {
			return fmt.Errorf("failed to decode JWK x coordinate: %w", err)
		}
		yBytes, err := base64.RawURLEncoding.DecodeString(key.Y)
		if err != nil {
			return fmt.Errorf("failed to decode JWK y coordinate: %w", err)
		}

		pubKey := &ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     new(big.Int).SetBytes(xBytes),
			Y:     new(big.Int).SetBytes(yBytes),
		}

		j.ecKeys[key.Kid] = pubKey
	}

	j.fetched = true
	return nil
}

// ParseSupabaseClaims parses and validates a Supabase JWT using the given JWKSKeyFunc.
func ParseSupabaseClaims(tokenStr string, kf *JWKSKeyFunc) (*SupabaseClaims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &SupabaseClaims{}, kf.Keyfunc)
	if err != nil {
		return nil, err
	}
	if c, ok := t.Claims.(*SupabaseClaims); ok && t.Valid {
		return c, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}
