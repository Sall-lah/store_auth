package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// Service handles RS256 token issuance, cryptographic verification, and JWKS exporting.
type Service struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	expiry     time.Duration
	keyID      string
}

// NewService initializes RSA keys from PEM files and configures token expiration.
// Why: Sets up cryptographic signing and verification infrastructure with explicit token lifetime boundaries.
func NewService(privateKeyPath, publicKeyPath string, expiryMinutes int) (*Service, error) {
	privKey, err := loadPrivateKey(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	pubKey, err := loadPublicKey(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	return &Service{
		privateKey: privKey,
		publicKey:  pubKey,
		expiry:     time.Duration(expiryMinutes) * time.Minute,
		keyID:      "store-auth-key-1",
	}, nil
}

// GenerateToken signs a new RS256 JWT access token containing subject ID, email, role, and expiration.
func (s *Service) GenerateToken(userID, email, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		Email: email,
		Role:  role,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(s.expiry)),
			Issuer:    "store_auth",
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodRS256, claims)
	token.Header["kid"] = s.keyID

	signedToken, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateToken parses and cryptographically verifies an incoming JWT string using the service's RSA public key.
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, &Claims{}, func(t *jwtlib.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwtlib.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// GetJWKS exports the RSA public key formatted as an RFC 7517 JSON Web Key Set for cross-service consumption.
func (s *Service) GetJWKS() JWKS {
	nStr := base64.RawURLEncoding.EncodeToString(s.publicKey.N.Bytes())
	eBytes := big.NewInt(int64(s.publicKey.E)).Bytes()
	eStr := base64.RawURLEncoding.EncodeToString(eBytes)

	jwk := JWK{
		Kty: "RSA",
		Use: "sig",
		Alg: "RS256",
		Kid: s.keyID,
		N:   nStr,
		E:   eStr,
	}

	return JWKS{
		Keys: []JWK{jwk},
	}
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, errors.New("invalid PEM block in private key")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not RSA")
	}

	return rsaKey, nil
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, errors.New("invalid PEM block in public key")
	}

	if key, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return key, nil
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not RSA")
	}

	return rsaKey, nil
}
