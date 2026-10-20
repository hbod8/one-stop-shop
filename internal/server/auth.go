package server

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Auth struct {
	config     *Config
	privateKey crypto.PrivateKey
	publicKey  crypto.PublicKey
	method     jwt.SigningMethod
}

var MissingConfigError = errors.New("missing config for Auth")

func NewAuth(config *Config) (*Auth, error) {
	if config == nil {
		return nil, MissingConfigError
	}

	privateKeyBytes, pvkerr := os.ReadFile(config.Auth.SecretKeyPath)
	publicKeyBytes, pbkerr := os.ReadFile(config.Auth.PublicKeyPath)

	bothMissing := errors.Is(pvkerr, fs.ErrNotExist) && errors.Is(pbkerr, fs.ErrNotExist)

	switch {
	case bothMissing && config.Build.GenerateSecrets:
		var err error
		privateKeyBytes, publicKeyBytes, err = generateSecrets(config.Auth.PublicKeyPath, config.Auth.SecretKeyPath, config.Auth.Crypto.Algorithm, config.Auth.Crypto.KeySize)
		if err != nil {
			return nil, fmt.Errorf("failed to generate secrets: %v", err)
		}
	case pvkerr != nil:
		return nil, fmt.Errorf("failed to read secret key: %v", pvkerr)
	case pbkerr != nil:
		return nil, fmt.Errorf("failed to read public key: %v", pbkerr)
	}

	privBlock, _ := pem.Decode(privateKeyBytes)
	if privBlock == nil {
		return nil, errors.New("private key is not valid PEM")
	}
	priv, err := x509.ParsePKCS8PrivateKey(privBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	pubBlock, _ := pem.Decode(publicKeyBytes)
	if pubBlock == nil {
		return nil, errors.New("public key is not valid PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	signer, ok := priv.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("private key type %T cannot sign", priv)
	}
	if eq, ok := pub.(interface{ Equal(crypto.PublicKey) bool }); !ok || !eq.Equal(signer.Public()) {
		return nil, errors.New("public key does not match private key")
	}

	var method jwt.SigningMethod
	switch priv.(type) {
	case ed25519.PrivateKey:
		method = jwt.SigningMethodEdDSA
	case *rsa.PrivateKey:
		method = jwt.SigningMethodRS256
	default:
		return nil, fmt.Errorf("unsupported private key type %T", priv)
	}

	return &Auth{config: config, privateKey: priv, publicKey: pub, method: method}, nil
}

var ErrInvalidCredentials = errors.New("invalid credentials")

type claims struct {
	jwt.RegisteredClaims
	Use string `json:"use"`
}

func (a *Auth) authenticate(username, password string) (token, refreshToken string, err error) {
	if username != "harry" || password != "secret" {
		return "", "", ErrInvalidCredentials
	}

	now := time.Now()
	token, err = a.sign(username, "access", now, time.Duration(a.config.Auth.TokenTtlSeconds)*time.Second)
	if err != nil {
		return "", "", fmt.Errorf("sign access token: %w", err)
	}
	refreshToken, err = a.sign(username, "refresh", now, time.Duration(a.config.Auth.RefreshTokenTtlSeconds)*time.Second)
	if err != nil {
		return "", "", fmt.Errorf("sign refresh token: %w", err)
	}
	return token, refreshToken, nil
}

func (a *Auth) sign(subject, use string, now time.Time, ttl time.Duration) (string, error) {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return "", err
	}
	c := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        hex.EncodeToString(id),
		},
		Use: use,
	}
	return jwt.NewWithClaims(a.method, c).SignedString(a.privateKey)
}

func generateSecrets(publicKeyPath, secretKeyPath, alg string, size int) (secretKey, publicKey []byte, err error) {
	var priv crypto.PrivateKey
	var pub crypto.PublicKey

	switch strings.ToLower(alg) {
	case "rsa":
		k, err := rsa.GenerateKey(rand.Reader, size)
		if err != nil {
			return nil, nil, fmt.Errorf("generate rsa key: %w", err)
		}
		priv, pub = k, &k.PublicKey
	case "ed25519":
		pk, sk, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, fmt.Errorf("generate ed25519 key: %w", err)
		}
		priv, pub = sk, pk
	default:
		return nil, nil, fmt.Errorf("unsupported key algorithm %q", alg)
	}

	privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal private key: %w", err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal public key: %w", err)
	}

	secretKey = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER})
	publicKey = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})

	if err := writeNewFile(secretKeyPath, secretKey, 0o600); err != nil {
		return nil, nil, fmt.Errorf("failed to write secret key: %v", err)
	}
	if err := writeNewFile(publicKeyPath, publicKey, 0o644); err != nil {
		os.Remove(secretKeyPath) // don't leave a private key without its public pair
		return nil, nil, fmt.Errorf("failed to write public key: %v", err)
	}
	return secretKey, publicKey, nil
}

func writeNewFile(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create directory for file %s: %v", path, err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", path, err)
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return fmt.Errorf("failed to write to file %s: %v", path, err)
	}
	return f.Close()
}
