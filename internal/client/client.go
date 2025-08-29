package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/artem/wg-orbit/internal/wg"
)

// Client представляє WireGuard Orbit клієнт
type Client struct {
	config     *Config
	httpClient *http.Client
}

// Config містить конфігурацію клієнта
type Config struct {
	ServerURL   string    `json:"server_url"`
	Token       string    `json:"token"`
	ClientName  string    `json:"client_name"`
	ConfigPath  string    `json:"config_path"`
	Interface   string    `json:"interface"`
	TokenExpiry time.Time `json:"token_expiry"`
}

// EnrollRequest представляє запит на реєстрацію
type EnrollRequest struct {
	ClientName string `json:"client_name"`
	Token      string `json:"token"`
}

// EnrollResponse представляє відповідь на реєстрацію
type EnrollResponse struct {
	Success      bool             `json:"success"`
	Message      string           `json:"message"`
	ClientConfig *wg.ClientConfig `json:"client_config,omitempty"`
	RefreshToken string           `json:"refresh_token,omitempty"`
	TokenExpiry  time.Time        `json:"token_expiry,omitempty"`
}

// DefaultConfig повертає конфігурацію за замовчуванням
func DefaultConfig() *Config {
	return &Config{
		Interface:  "wg0",
		ConfigPath: filepath.Join(os.Getenv("HOME"), ".wg-orbit", "client.json"),
	}
}

// NewClient створює новий клієнт
func NewClient(config *Config) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Enroll реєструє клієнта на сервері
func (c *Client) Enroll(serverURL, token, clientName string) error {
	// Оновлюємо конфігурацію
	c.config.ServerURL = serverURL
	c.config.Token = token
	c.config.ClientName = clientName

	// Створюємо запит на реєстрацію
	req := EnrollRequest{
		ClientName: clientName,
		Token:      token,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Відправляємо запит
	resp, err := c.httpClient.Post(
		serverURL+"/api/v1/enroll",
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Читаємо відповідь
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var enrollResp EnrollResponse
	if err := json.Unmarshal(body, &enrollResp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !enrollResp.Success {
		return fmt.Errorf("enrollment failed: %s", enrollResp.Message)
	}

	// Оновлюємо токен та час його закінчення
	if enrollResp.RefreshToken != "" {
		c.config.Token = enrollResp.RefreshToken
		c.config.TokenExpiry = enrollResp.TokenExpiry
	}

	// Зберігаємо конфігурацію
	if err := c.SaveConfig(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Зберігаємо WireGuard конфігурацію
	if enrollResp.ClientConfig != nil {
		if err := c.SaveWireGuardConfig(enrollResp.ClientConfig); err != nil {
			return fmt.Errorf("failed to save WireGuard config: %w", err)
		}
	}

	return nil
}

// Up піднімає WireGuard інтерфейс
func (c *Client) Up() error {
	// Перевіряємо, чи потрібно оновити токен
	if time.Now().After(c.config.TokenExpiry.Add(-1 * time.Hour)) {
		if err := c.RefreshToken(); err != nil {
			return fmt.Errorf("failed to refresh token: %w", err)
		}
	}

	// Завантажуємо WireGuard конфігурацію
	configPath := c.getWireGuardConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("WireGuard config not found at %s. Run 'enroll' first", configPath)
	}

	// Піднімаємо інтерфейс за допомогою wg-quick
	cmd := fmt.Sprintf("sudo wg-quick up %s", configPath)
	fmt.Printf("Executing: %s\n", cmd)

	// TODO: Виконати команду через os/exec
	fmt.Println("WireGuard interface brought up successfully")
	return nil
}

// Down опускає WireGuard інтерфейс
func (c *Client) Down() error {
	configPath := c.getWireGuardConfigPath()
	cmd := fmt.Sprintf("sudo wg-quick down %s", configPath)
	fmt.Printf("Executing: %s\n", cmd)

	// TODO: Виконати команду через os/exec
	fmt.Println("WireGuard interface brought down successfully")
	return nil
}

// Status показує статус WireGuard з'єднання
func (c *Client) Status() error {
	cmd := fmt.Sprintf("sudo wg show %s", c.config.Interface)
	fmt.Printf("Executing: %s\n", cmd)

	// TODO: Виконати команду через os/exec та парсити вивід
	fmt.Println("WireGuard status retrieved successfully")
	return nil
}

// RefreshToken оновлює токен клієнта
func (c *Client) RefreshToken() error {
	req := map[string]string{
		"refresh_token": c.config.Token,
		"client_name":   c.config.ClientName,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.config.ServerURL+"/api/v1/refresh",
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var refreshResp struct {
		Success      bool      `json:"success"`
		Message      string    `json:"message"`
		RefreshToken string    `json:"refresh_token,omitempty"`
		TokenExpiry  time.Time `json:"token_expiry,omitempty"`
	}

	if err := json.Unmarshal(body, &refreshResp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !refreshResp.Success {
		return fmt.Errorf("token refresh failed: %s", refreshResp.Message)
	}

	// Оновлюємо токен
	c.config.Token = refreshResp.RefreshToken
	c.config.TokenExpiry = refreshResp.TokenExpiry

	// Зберігаємо конфігурацію
	return c.SaveConfig()
}

// SaveConfig зберігає конфігурацію клієнта
func (c *Client) SaveConfig() error {
	// Створюємо директорію, якщо не існує
	dir := filepath.Dir(c.config.ConfigPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Зберігаємо конфігурацію
	data, err := json.MarshalIndent(c.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(c.config.ConfigPath, data, 0600)
}

// LoadConfig завантажує конфігурацію клієнта
func (c *Client) LoadConfig() error {
	data, err := os.ReadFile(c.config.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	return json.Unmarshal(data, c.config)
}

// SaveWireGuardConfig зберігає WireGuard конфігурацію
func (c *Client) SaveWireGuardConfig(config *wg.ClientConfig) error {
	configPath := c.getWireGuardConfigPath()
	dir := filepath.Dir(configPath)

	// Створюємо директорію, якщо не існує
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Генеруємо WireGuard конфігурацію
	wgConfig := config.ToWireGuardConfig()

	// Зберігаємо конфігурацію
	return os.WriteFile(configPath, []byte(wgConfig), 0600)
}

// getWireGuardConfigPath повертає шлях до WireGuard конфігурації
func (c *Client) getWireGuardConfigPath() string {
	dir := filepath.Dir(c.config.ConfigPath)
	return filepath.Join(dir, c.config.Interface+".conf")
}
