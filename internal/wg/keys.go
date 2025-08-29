package wg

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

// GenerateKeyPair генерує пару приватний/публічний ключ для WireGuard
func GenerateKeyPair() (privateKey, publicKey string, err error) {
	// Генеруємо приватний ключ (32 байти)
	privateKeyBytes := make([]byte, 32)
	if _, err := rand.Read(privateKeyBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}

	// Обробляємо приватний ключ згідно з RFC 7748
	privateKeyBytes[0] &= 248
	privateKeyBytes[31] &= 127
	privateKeyBytes[31] |= 64

	// Генеруємо публічний ключ з приватного
	publicKeyBytes, err := curve25519.X25519(privateKeyBytes, curve25519.Basepoint)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate public key: %w", err)
	}

	// Кодуємо ключі в base64
	privateKey = base64.StdEncoding.EncodeToString(privateKeyBytes)
	publicKey = base64.StdEncoding.EncodeToString(publicKeyBytes)

	return privateKey, publicKey, nil
}

// ValidatePrivateKey перевіряє валідність приватного ключа
func ValidatePrivateKey(key string) error {
	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return fmt.Errorf("invalid base64 encoding: %w", err)
	}

	if len(keyBytes) != 32 {
		return fmt.Errorf("private key must be 32 bytes, got %d", len(keyBytes))
	}

	return nil
}

// ValidatePublicKey перевіряє валідність публічного ключа
func ValidatePublicKey(key string) error {
	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return fmt.Errorf("invalid base64 encoding: %w", err)
	}

	if len(keyBytes) != 32 {
		return fmt.Errorf("public key must be 32 bytes, got %d", len(keyBytes))
	}

	return nil
}

// GeneratePresharedKey генерує preshared ключ для додаткової безпеки
func GeneratePresharedKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("failed to generate preshared key: %w", err)
	}

	return base64.StdEncoding.EncodeToString(key), nil
}
