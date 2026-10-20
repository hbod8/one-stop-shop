package server

import (
	"errors"
	"fmt"
)

type Auth struct {
	config *Config
}

func NewAuth(config *Config) *Auth {
	return &Auth{config: config}
}

var ErrInvalidCredentials = errors.New("invalid credentials")

func (a *Auth) Authenticate(username, password string) (token, refreshToken string, err error) {
	if username == "harry" && password == "secret" {
		token = fmt.Sprintf("fake-jwt-token-exp:%d", a.config.Auth.TokenTtlSeconds)
		refreshToken = fmt.Sprintf("fake-jwt-refresh-token-exp:%d", a.config.Auth.RefreshTokenTtlSeconds)
		return token, refreshToken, nil
	}
	return "", "", ErrInvalidCredentials
}
