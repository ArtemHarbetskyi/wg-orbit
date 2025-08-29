package wg

import (
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
)

// Peer представляє WireGuard peer'а
type Peer struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	PublicKey   string    `json:"public_key" db:"public_key"`
	PrivateKey  string    `json:"-" db:"private_key"` // Не експортується в JSON
	AllowedIPs  []string  `json:"allowed_ips" db:"allowed_ips"`
	Endpoint    string    `json:"endpoint,omitempty" db:"endpoint"`
	PresharedKey string   `json:"preshared_key,omitempty" db:"preshared_key"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	LastSeen    *time.Time `json:"last_seen,omitempty" db:"last_seen"`
	IsActive    bool      `json:"is_active" db:"is_active"`
}

// Interface представляє WireGuard інтерфейс
type Interface struct {
	Name       string    `json:"name" db:"name"`
	PublicKey  string    `json:"public_key" db:"public_key"`
	PrivateKey string    `json:"-" db:"private_key"`
	ListenPort int       `json:"listen_port" db:"listen_port"`
	Address    string    `json:"address" db:"address"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// Config представляє повну конфігурацію WireGuard
type Config struct {
	Interface Interface `json:"interface"`
	Peers     []Peer    `json:"peers"`
}

// ClientConfig представляє конфігурацію для клієнта
type ClientConfig struct {
	Interface ClientInterface `json:"interface"`
	Peer      ServerPeer      `json:"peer"`
}

// ClientInterface представляє інтерфейс клієнта
type ClientInterface struct {
	PrivateKey string   `json:"private_key"`
	Address    []string `json:"address"`
	DNS        []string `json:"dns,omitempty"`
}

// ServerPeer представляє сервер як peer для клієнта
type ServerPeer struct {
	PublicKey    string   `json:"public_key"`
	Endpoint     string   `json:"endpoint"`
	AllowedIPs   []string `json:"allowed_ips"`
	PresharedKey string   `json:"preshared_key,omitempty"`
}

// HandshakeInfo представляє інформацію про handshake
type HandshakeInfo struct {
	PeerID       uuid.UUID `json:"peer_id"`
	LastHandshake time.Time `json:"last_handshake"`
	RxBytes      int64     `json:"rx_bytes"`
	TxBytes      int64     `json:"tx_bytes"`
}

// IsOnline перевіряє, чи peer онлайн на основі часу останнього handshake
func (h *HandshakeInfo) IsOnline(maxAge time.Duration) bool {
	if h.LastHandshake.IsZero() {
		return false
	}
	return time.Since(h.LastHandshake) <= maxAge
}

// IPPool представляє пул IP адрес
type IPPool struct {
	Network   *net.IPNet `json:"network"`
	Allocated map[string]bool `json:"allocated"`
	NextIP    net.IP     `json:"next_ip"`
}

// NewIPPool створює новий пул IP адрес
func NewIPPool(cidr string) (*IPPool, error) {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %w", err)
	}
	
	return &IPPool{
		Network:   network,
		Allocated: make(map[string]bool),
		NextIP:    network.IP,
	}, nil
}

// AllocateIP виділяє наступну доступну IP адресу
func (pool *IPPool) AllocateIP() (net.IP, error) {
	for ip := pool.NextIP; pool.Network.Contains(ip); ip = nextIP(ip) {
		ipStr := ip.String()
		if !pool.Allocated[ipStr] {
			pool.Allocated[ipStr] = true
			pool.NextIP = nextIP(ip)
			return ip, nil
		}
	}
	return nil, fmt.Errorf("no available IP addresses in pool")
}

// ReleaseIP звільняє IP адресу
func (pool *IPPool) ReleaseIP(ip net.IP) {
	delete(pool.Allocated, ip.String())
}

// nextIP повертає наступну IP адресу
func nextIP(ip net.IP) net.IP {
	next := make(net.IP, len(ip))
	copy(next, ip)
	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] != 0 {
			break
		}
	}
	return next
}

// NewPeer створює новий peer з згенерованими ключами
func NewPeer(name string) (*Peer, error) {
	if name == "" {
		return nil, fmt.Errorf("peer name cannot be empty")
	}
	
	privateKey, publicKey, err := GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	return &Peer{
		ID:         uuid.New(),
		Name:       name,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		IsActive:   true,
	}, nil
}

// ToWireGuardConfig генерує конфігурацію у форматі WireGuard
func (c *ClientConfig) ToWireGuardConfig() string {
	config := "[Interface]\n"
	config += "PrivateKey = " + c.Interface.PrivateKey + "\n"
	
	for _, addr := range c.Interface.Address {
		config += "Address = " + addr + "\n"
	}
	
	for _, dns := range c.Interface.DNS {
		config += "DNS = " + dns + "\n"
	}
	
	config += "\n[Peer]\n"
	config += "PublicKey = " + c.Peer.PublicKey + "\n"
	config += "Endpoint = " + c.Peer.Endpoint + "\n"
	
	for _, allowedIP := range c.Peer.AllowedIPs {
		config += "AllowedIPs = " + allowedIP + "\n"
	}
	
	if c.Peer.PresharedKey != "" {
		config += "PresharedKey = " + c.Peer.PresharedKey + "\n"
	}
	
	return config
}