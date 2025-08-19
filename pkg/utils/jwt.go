package utils

import (
	"errors"
	"time"

	"eiam-platform/config"

	"github.com/golang-jwt/jwt/v5"
)

// AccessTokenClaims access token claims
type AccessTokenClaims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	DisplayName string   `json:"display_name"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	SessionID   string   `json:"session_id"` // 添加session_id关联
	TradeID     string   `json:"trade_id"`
	TokenType   string   `json:"token_type"` // "access"
	jwt.RegisteredClaims
}

// RefreshTokenClaims refresh token claims
type RefreshTokenClaims struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	TradeID   string `json:"trade_id"`
	TokenType string `json:"token_type"` // "refresh"
	jwt.RegisteredClaims
}

// JWTManager JWT manager
type JWTManager struct {
	secretKey            []byte
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
	issuer               string
}

// NewJWTManager create JWT manager
func NewJWTManager(cfg *config.JWTConfig) *JWTManager {
	return &JWTManager{
		secretKey:            []byte(cfg.Secret),
		accessTokenDuration:  time.Duration(cfg.AccessTokenExpire) * time.Second,
		refreshTokenDuration: time.Duration(cfg.RefreshTokenExpire) * time.Second,
		issuer:               cfg.Issuer,
	}
}

// TokenInfo token information structure
type TokenInfo struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	DisplayName string   `json:"display_name"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	SessionID   string   `json:"session_id"`
	TradeID     string   `json:"trade_id"`
}

// GenerateAccessToken generate access token
func (j *JWTManager) GenerateAccessToken(tokenInfo *TokenInfo) (string, error) {
	now := time.Now()
	claims := AccessTokenClaims{
		UserID:      tokenInfo.UserID,
		Username:    tokenInfo.Username,
		Email:       tokenInfo.Email,
		DisplayName: tokenInfo.DisplayName,
		Roles:       tokenInfo.Roles,
		Permissions: tokenInfo.Permissions,
		SessionID:   tokenInfo.SessionID, // 添加session_id
		TradeID:     tokenInfo.TradeID,
		TokenType:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    j.issuer,
			Subject:   tokenInfo.UserID,
			ID:        GenerateTradeIDString("access"),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// GenerateRefreshToken generate refresh token
func (j *JWTManager) GenerateRefreshToken(userID, sessionID, tradeID string) (string, error) {
	now := time.Now()
	claims := RefreshTokenClaims{
		UserID:    userID,
		SessionID: sessionID,
		TradeID:   tradeID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    j.issuer,
			Subject:   userID,
			ID:        GenerateTradeIDString("refresh"),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateAccessToken validate access token
func (j *JWTManager) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AccessTokenClaims); ok && token.Valid {
		if claims.TokenType != "access" {
			return nil, errors.New("token type mismatch")
		}
		return claims, nil
	}

	return nil, errors.New("invalid access token")
}

// ValidateRefreshToken validate refresh token
func (j *JWTManager) ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*RefreshTokenClaims); ok && token.Valid {
		if claims.TokenType != "refresh" {
			return nil, errors.New("token type mismatch")
		}
		return claims, nil
	}

	return nil, errors.New("invalid refresh token")
}

// ExtractTokenFromHeader extract token from request header
func ExtractTokenFromHeader(authHeader string) string {
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}

// GetTokenExpiration get token expiration time
func (j *JWTManager) GetTokenExpiration() time.Duration {
	return j.accessTokenDuration
}

// GetRefreshTokenExpiration get refresh token expiration time
func (j *JWTManager) GetRefreshTokenExpiration() time.Duration {
	return j.refreshTokenDuration
}

// GenerateAccessToken generate access token (simple function - legacy)
func GenerateAccessToken(userID, username, secret string, expireSeconds int) (string, error) {
	return GenerateAccessTokenWithUserInfo(userID, username, "", "", []string{}, []string{}, secret, expireSeconds)
}

// GenerateAccessTokenWithUserInfo generate access token with complete user info
func GenerateAccessTokenWithUserInfo(userID, username, email, displayName string, roles, permissions []string, secret string, expireSeconds int) (string, error) {
	now := time.Now()
	claims := AccessTokenClaims{
		UserID:      userID,
		Username:    username,
		Email:       email,
		DisplayName: displayName,
		Roles:       roles,
		Permissions: permissions,
		TokenType:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expireSeconds) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "eiam-platform",
			Subject:   userID,
			ID:        GenerateTradeIDString("access"),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken generate refresh token (simple function)
func GenerateRefreshToken(userID, username, secret string, expireSeconds int) (string, error) {
	now := time.Now()
	claims := RefreshTokenClaims{
		UserID:    userID,
		SessionID: GenerateTradeIDString("session"),
		TradeID:   GenerateTradeIDString("refresh"),
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expireSeconds) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "eiam-platform",
			Subject:   userID,
			ID:        GenerateTradeIDString("refresh"),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateRefreshToken validate refresh token (simple function)
func ValidateRefreshToken(tokenString, secret string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*RefreshTokenClaims); ok && token.Valid {
		if claims.TokenType != "refresh" {
			return nil, errors.New("token type mismatch")
		}
		return claims, nil
	}

	return nil, errors.New("invalid refresh token")
}
