package session

import (
	"errors"
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/o1egl/paseto"
	"golang.org/x/crypto/chacha20poly1305"
)

type UserSession struct {
	Token        string
	RefreshToken string
}

type IdentityGenerator interface {
	GenerateTokens(userID string) (UserSession, error)
	ValidateToken(t string) error
}

// todo: get this from env
type Config struct {
	TokenSecret []byte
}

type jwtTokenManager struct {
	config Config
}

func NewJsonWebToken(config Config) *jwtTokenManager {
	return &jwtTokenManager{config: config}
}

type pasetoTokenManager struct {
	config Config
}

func NewPasetoTokenManager(config Config) *pasetoTokenManager {
	return &pasetoTokenManager{config: config}
}

func (p pasetoTokenManager) GenerateTokens(userID string) (UserSession, error) {
	if len(p.config.TokenSecret) != chacha20poly1305.KeySize {
		return UserSession{}, nil
	}
	now := time.Now()
	exp := now.Add(24 * time.Hour)
	nbt := now

	jsonToken := paseto.JSONToken{
		Audience:   "test",
		Issuer:     "test_service",
		Jti:        "123",
		Subject:    "test_subject",
		IssuedAt:   now,
		Expiration: exp,
		NotBefore:  nbt,
	}

	// Encrypt data
	v2 := paseto.NewV2()
	token, err := v2.Encrypt(p.config.TokenSecret, jsonToken, nil)
	if err != nil {
		return UserSession{}, fmt.Errorf("encrypt token: %w", err)
	}

	jsonRefreshToken := paseto.JSONToken{
		Audience:   "test",
		Issuer:     "test_service",
		Jti:        "123",
		Subject:    "test_subject",
		IssuedAt:   now,
		Expiration: time.Now().Add(24 * time.Hour),
		NotBefore:  time.Now().Add(24 * time.Hour),
	}

	refreshToken, err := v2.Encrypt(p.config.TokenSecret, jsonRefreshToken, nil)
	if err != nil {
		return UserSession{}, fmt.Errorf("encrypt token: %w", err)
	}

	return UserSession{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil

}

func (p pasetoTokenManager) ValidateToken(token string) error {
	v2 := paseto.NewV2()
	var newJsonToken paseto.JSONToken

	err := v2.Decrypt(token, p.config.TokenSecret, &newJsonToken, nil)
	if err != nil {
		return fmt.Errorf("decrypt token: %w", err)
	}
	err = newJsonToken.Validate()
	if err != nil {
		return fmt.Errorf("invalid token due to: %w", err)
	}

	return nil
}

func (j jwtTokenManager) GenerateTokens(userID string) (UserSession, error) {
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  userID,
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})

	token, err := tokenClaims.SignedString(j.config.TokenSecret)
	if err != nil {
		return UserSession{}, fmt.Errorf("problem to sign token: %w", err)
	}

	refreshTokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})

	refreshToken, err := refreshTokenClaims.SignedString(j.config.TokenSecret)
	if err != nil {
		return UserSession{}, fmt.Errorf("problem to sign token: %w", err)
	}

	return UserSession{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func (j jwtTokenManager) ValidateToken(t string) error {
	token, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return j.config.TokenSecret, nil
	})
	if err != nil {
		return fmt.Errorf("validate token err: %w", err)
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return nil
	}

	return errors.New("token is not valid - expired")
}
