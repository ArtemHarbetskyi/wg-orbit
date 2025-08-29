package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"

	"github.com/artem/wg-orbit/internal/wg"
)

// SQLiteStorage реалізує Storage інтерфейс для SQLite
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage створює новий SQLite storage
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &SQLiteStorage{db: db}
	if err := storage.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return storage, nil
}

// createTables створює необхідні таблиці
func (s *SQLiteStorage) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS interfaces (
			name TEXT PRIMARY KEY,
			public_key TEXT NOT NULL,
			private_key TEXT NOT NULL,
			listen_port INTEGER NOT NULL,
			address TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS peers (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			public_key TEXT NOT NULL UNIQUE,
			private_key TEXT NOT NULL,
			allowed_ips TEXT NOT NULL,
			endpoint TEXT,
			preshared_key TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			last_seen DATETIME,
			is_active BOOLEAN NOT NULL DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS tokens (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			username TEXT NOT NULL,
			token_hash TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL,
			is_used BOOLEAN NOT NULL DEFAULT 0
		)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

// SaveInterface зберігає інтерфейс
func (s *SQLiteStorage) SaveInterface(iface *wg.Interface) error {
	query := `INSERT OR REPLACE INTO interfaces 
			   (name, public_key, private_key, listen_port, address, created_at, updated_at)
			   VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, iface.Name, iface.PublicKey, iface.PrivateKey,
		iface.ListenPort, iface.Address, iface.CreatedAt, iface.UpdatedAt)

	return err
}

// GetInterface отримує інтерфейс за назвою
func (s *SQLiteStorage) GetInterface(name string) (*wg.Interface, error) {
	query := `SELECT name, public_key, private_key, listen_port, address, created_at, updated_at
			   FROM interfaces WHERE name = ?`

	row := s.db.QueryRow(query, name)

	var iface wg.Interface
	err := row.Scan(&iface.Name, &iface.PublicKey, &iface.PrivateKey,
		&iface.ListenPort, &iface.Address, &iface.CreatedAt, &iface.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &iface, err
}

// SavePeer зберігає peer
func (s *SQLiteStorage) SavePeer(peer *wg.Peer) error {
	allowedIPsStr := strings.Join(peer.AllowedIPs, ",")

	query := `INSERT OR REPLACE INTO peers 
			   (id, name, public_key, private_key, allowed_ips, endpoint, preshared_key, 
			    created_at, updated_at, last_seen, is_active)
			   VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, peer.ID.String(), peer.Name, peer.PublicKey, peer.PrivateKey,
		allowedIPsStr, peer.Endpoint, peer.PresharedKey, peer.CreatedAt, peer.UpdatedAt,
		peer.LastSeen, peer.IsActive)

	return err
}

// GetPeer отримує peer за ID
func (s *SQLiteStorage) GetPeer(id uuid.UUID) (*wg.Peer, error) {
	query := `SELECT id, name, public_key, private_key, allowed_ips, endpoint, preshared_key,
			         created_at, updated_at, last_seen, is_active
			   FROM peers WHERE id = ?`

	row := s.db.QueryRow(query, id.String())

	var peer wg.Peer
	var allowedIPsStr string
	var idStr string

	err := row.Scan(&idStr, &peer.Name, &peer.PublicKey, &peer.PrivateKey,
		&allowedIPsStr, &peer.Endpoint, &peer.PresharedKey, &peer.CreatedAt,
		&peer.UpdatedAt, &peer.LastSeen, &peer.IsActive)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	peer.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse peer ID: %w", err)
	}

	if allowedIPsStr != "" {
		peer.AllowedIPs = strings.Split(allowedIPsStr, ",")
	}

	return &peer, nil
}

// GetPeerByName отримує peer за ім'ям
func (s *SQLiteStorage) GetPeerByName(name string) (*wg.Peer, error) {
	query := `SELECT id, name, public_key, private_key, allowed_ips, endpoint, preshared_key,
			         created_at, updated_at, last_seen, is_active
			   FROM peers WHERE name = ?`

	row := s.db.QueryRow(query, name)

	var peer wg.Peer
	var allowedIPsStr string
	var idStr string

	err := row.Scan(&idStr, &peer.Name, &peer.PublicKey, &peer.PrivateKey,
		&allowedIPsStr, &peer.Endpoint, &peer.PresharedKey, &peer.CreatedAt,
		&peer.UpdatedAt, &peer.LastSeen, &peer.IsActive)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	peer.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse peer ID: %w", err)
	}

	if allowedIPsStr != "" {
		peer.AllowedIPs = strings.Split(allowedIPsStr, ",")
	}

	return &peer, nil
}

// ListPeers повертає список всіх peer'ів
func (s *SQLiteStorage) ListPeers() ([]*wg.Peer, error) {
	query := `SELECT id, name, public_key, private_key, allowed_ips, endpoint, preshared_key,
			         created_at, updated_at, last_seen, is_active
			   FROM peers ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var peers []*wg.Peer
	for rows.Next() {
		var peer wg.Peer
		var allowedIPsStr string
		var idStr string

		err := rows.Scan(&idStr, &peer.Name, &peer.PublicKey, &peer.PrivateKey,
			&allowedIPsStr, &peer.Endpoint, &peer.PresharedKey, &peer.CreatedAt,
			&peer.UpdatedAt, &peer.LastSeen, &peer.IsActive)
		if err != nil {
			return nil, err
		}

		peer.ID, err = uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse peer ID: %w", err)
		}

		if allowedIPsStr != "" {
			peer.AllowedIPs = strings.Split(allowedIPsStr, ",")
		}

		peers = append(peers, &peer)
	}

	return peers, rows.Err()
}

// DeletePeer видаляє peer
func (s *SQLiteStorage) DeletePeer(id uuid.UUID) error {
	query := `DELETE FROM peers WHERE id = ?`
	_, err := s.db.Exec(query, id.String())
	return err
}

// UpdatePeerLastSeen оновлює час останнього підключення peer'а
func (s *SQLiteStorage) UpdatePeerLastSeen(id uuid.UUID, lastSeen time.Time) error {
	query := `UPDATE peers SET last_seen = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, lastSeen, time.Now(), id.String())
	return err
}

// Close закриває з'єднання з базою даних
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}