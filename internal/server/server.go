package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/artem/wg-orbit/api/rest"
	"github.com/artem/wg-orbit/internal/auth"
	"github.com/artem/wg-orbit/internal/storage"
	"github.com/artem/wg-orbit/internal/wg"
)

// Server представляє WireGuard Orbit сервер
type Server struct {
	storage     storage.Storage
	tokenMgr    *auth.TokenManager
	restServer  *rest.Server
	interfaceMgr *InterfaceManager
	config      *Config
}

// Config представляє конфігурацію сервера
type Config struct {
	Host         string `yaml:"host" json:"host"`
	Port         int    `yaml:"port" json:"port"`
	Interface    string `yaml:"interface" json:"interface"`
	StorageType  string `yaml:"storage_type" json:"storage_type"`
	StorageConfig storage.Config `yaml:"storage" json:"storage"`
	JWTSecret    string `yaml:"jwt_secret" json:"jwt_secret"`
	TokenTTL     time.Duration `yaml:"token_ttl" json:"token_ttl"`
	IPAMNetwork  string `yaml:"ipam_network" json:"ipam_network"`
}

// DefaultConfig повертає конфігурацію за замовчуванням
func DefaultConfig() *Config {
	return &Config{
		Host:        "0.0.0.0",
		Port:        8080,
		Interface:   "wg0",
		StorageType: "sqlite",
		StorageConfig: storage.Config{
			Type:     "sqlite",
			Database: "wg-orbit.db",
		},
		JWTSecret:   "change-me-in-production",
		TokenTTL:    24 * time.Hour,
		IPAMNetwork: "10.0.0.0/24",
	}
}

// NewServer створює новий сервер
func NewServer(config *Config) (*Server, error) {
	// Ініціалізація storage
	store, err := storage.NewStorage(&config.StorageConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Ініціалізація token manager
	tokenMgr := auth.NewTokenManager([]byte(config.JWTSecret), "wg-orbit")

	// Ініціалізація interface manager
	interfaceMgr, err := NewInterfaceManager(config.Interface, config.IPAMNetwork)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize interface manager: %w", err)
	}

	// Ініціалізація REST API
	restConfig := &rest.Config{
		Host: config.Host,
		Port: config.Port,
	}
	restServer := rest.NewServer(store, tokenMgr, restConfig)

	return &Server{
		storage:      store,
		tokenMgr:     tokenMgr,
		restServer:   restServer,
		interfaceMgr: interfaceMgr,
		config:       config,
	}, nil
}

// Initialize ініціалізує WireGuard інтерфейс
func (s *Server) Initialize() error {
	log.Printf("Initializing WireGuard interface: %s", s.config.Interface)
	
	// Створюємо інтерфейс якщо він не існує
	if err := s.interfaceMgr.CreateInterface(); err != nil {
		return fmt.Errorf("failed to create interface: %w", err)
	}

	// Зберігаємо конфігурацію інтерфейсу в БД
	interfaceConfig := &wg.Interface{
		Name:       s.config.Interface,
		PublicKey:  s.interfaceMgr.PublicKey(),
		PrivateKey: s.interfaceMgr.PrivateKey(),
		ListenPort: s.interfaceMgr.ListenPort(),
		Address:    s.config.IPAMNetwork,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.storage.SaveInterface(interfaceConfig); err != nil {
		return fmt.Errorf("failed to save interface config: %w", err)
	}

	log.Printf("WireGuard interface %s initialized successfully", s.config.Interface)
	return nil
}

// Run запускає сервер
func (s *Server) Run() error {
	log.Printf("Starting WireGuard Orbit server on %s:%d", s.config.Host, s.config.Port)

	// Запускаємо REST API сервер
	go func() {
		if err := s.restServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start REST server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// REST server не має методу Shutdown, тому просто логуємо
	log.Println("REST server shutdown not implemented")

	if err := s.storage.Close(); err != nil {
		log.Printf("Error closing storage: %v", err)
	}

	log.Println("Server exited")
	return nil
}

// AddUser додає нового користувача
func (s *Server) AddUser(username string) error {
	log.Printf("Adding user: %s", username)

	// Створюємо нового peer'а
	peer, err := wg.NewPeer(username)
	if err != nil {
		return fmt.Errorf("failed to create peer: %w", err)
	}

	// Виділяємо IP адресу
	ip, err := s.interfaceMgr.AllocateIP()
	if err != nil {
		return fmt.Errorf("failed to allocate IP: %w", err)
	}

	peer.AllowedIPs = []string{ip.String() + "/32"}

	// Зберігаємо peer'а в БД
	if err := s.storage.SavePeer(peer); err != nil {
		return fmt.Errorf("failed to save peer: %w", err)
	}

	log.Printf("User %s added successfully with IP %s", username, ip.String())
	return nil
}

// GenerateToken генерує токен для користувача
func (s *Server) GenerateToken(username string) (string, error) {
	log.Printf("Generating token for user: %s", username)

	// Перевіряємо, чи існує користувач
	peer, err := s.storage.GetPeerByName(username)
	if err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	// Генеруємо токен
	token, err := s.tokenMgr.GenerateToken(peer.ID, username, "user", &peer.ID, 24*time.Hour)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	log.Printf("Token generated for user %s", username)
	return token, nil
}