package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims представляє JWT claims для wg-orbit
type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	PeerID   uuid.UUID `json:"peer_id,omitempty"`
	jwt.RegisteredClaims
}

// TokenManager керує JWT токенами
type TokenManager struct {
	secretKey []byte
	issuer    string
}

// NewTokenManager створює новий менеджер токенів
func NewTokenManager(secretKey []byte, issuer string) *TokenManager {
	return &TokenManager{
		secretKey: secretKey,
		issuer:    issuer,
	}
}

// GenerateToken генерує JWT токен для користувача
func (tm *TokenManager) GenerateToken(userID uuid.UUID, username, role string, peerID *uuid.UUID, duration time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tm.issuer,
			Subject:   userID.String(),
			Audience:  []string{"wg-orbit"},
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	if peerID != nil {
		claims.PeerID = *peerID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.secretKey)
}

// ValidateToken валідує JWT токен і повертає claims
func (tm *TokenManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// GenerateEnrollmentToken генерує одноразовий токен для реєстрації
func (tm *TokenManager) GenerateEnrollmentToken(username string, duration time.Duration) (string, error) {
	userID := uuid.New() // Тимчасовий ID для enrollment
	return tm.GenerateToken(userID, username, "enrollment", nil, duration)
}

// RefreshToken оновлює токен з новим терміном дії
func (tm *TokenManager) RefreshToken(tokenString string, duration time.Duration) (string, error) {
	claims, err := tm.ValidateToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("invalid token for refresh: %w", err)
	}

	// Генеруємо новий токен з тими ж даними, але новим терміном дії
	var peerID *uuid.UUID
	if claims.PeerID != uuid.Nil {
		peerID = &claims.PeerID
	}

	return tm.GenerateToken(claims.UserID, claims.Username, claims.Role, peerID, duration)
}
