package storage

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/artem/wg-orbit/internal/wg"
)

// Storage інтерфейс для роботи з базою даних
type Storage interface {
	// Interface operations
	SaveInterface(iface *wg.Interface) error
	GetInterface(name string) (*wg.Interface, error)

	// Peer operations
	SavePeer(peer *wg.Peer) error
	GetPeer(id uuid.UUID) (*wg.Peer, error)
	GetPeerByName(name string) (*wg.Peer, error)
	ListPeers() ([]*wg.Peer, error)
	DeletePeer(id uuid.UUID) error
	UpdatePeerLastSeen(id uuid.UUID, lastSeen time.Time) error

	// Connection management
	Close() error
}

// Config представляє конфігурацію для storage
type Config struct {
	Type     string `yaml:"type" json:"type"`         // sqlite, postgres
	Host     string `yaml:"host" json:"host"`         // для postgres
	Port     int    `yaml:"port" json:"port"`         // для postgres
	Database string `yaml:"database" json:"database"` // назва БД або шлях до файлу
	Username string `yaml:"username" json:"username"` // для postgres
	Password string `yaml:"password" json:"password"` // для postgres
	SSLMode  string `yaml:"ssl_mode" json:"ssl_mode"` // для postgres
}

// NewStorage створює новий storage на основі конфігурації
func NewStorage(config *Config) (Storage, error) {
	switch config.Type {
	case "sqlite", "":
		return NewSQLiteStorage(config.Database)
	case "postgres":
		// TODO: Implement PostgreSQL storage
		return nil, fmt.Errorf("PostgreSQL storage not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.Type)
	}
}

// DefaultConfig повертає конфігурацію за замовчуванням
func DefaultConfig() *Config {
	return &Config{
		Type:     "sqlite",
		Database: "wg-orbit.db",
	}
}
