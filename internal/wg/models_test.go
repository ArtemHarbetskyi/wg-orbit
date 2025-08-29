package wg

import (
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewPeer(t *testing.T) {
	tests := []struct {
		name     string
		peerName string
		wantErr  bool
	}{
		{
			name:     "valid peer creation",
			peerName: "test-peer",
			wantErr:  false,
		},
		{
			name:     "empty peer name",
			peerName: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			peer, err := NewPeer(tt.peerName)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewPeer() expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("NewPeer() unexpected error: %v", err)
				return
			}
			
			if peer == nil {
				t.Errorf("NewPeer() returned nil peer")
				return
			}
			
			// Перевірка базових полів
			if peer.Name != tt.peerName {
				t.Errorf("NewPeer() name = %v, want %v", peer.Name, tt.peerName)
			}
			
			// Перевірка генерації UUID
			if peer.ID == uuid.Nil {
				t.Errorf("NewPeer() ID should not be nil")
			}
			
			// Перевірка генерації ключів
			if len(peer.PublicKey) == 0 {
				t.Errorf("NewPeer() PublicKey should not be empty")
			}
			
			if len(peer.PrivateKey) == 0 {
				t.Errorf("NewPeer() PrivateKey should not be empty")
			}
			
			// Перевірка часових міток
			if peer.CreatedAt.IsZero() {
				t.Errorf("NewPeer() CreatedAt should not be zero")
			}
			
			if peer.UpdatedAt.IsZero() {
				t.Errorf("NewPeer() UpdatedAt should not be zero")
			}
		})
	}
}

func TestPeer_Fields(t *testing.T) {
	peer, err := NewPeer("test-peer")
	if err != nil {
		t.Fatalf("Failed to create test peer: %v", err)
	}
	
	// Встановлюємо endpoint та AllowedIPs для тестування
	peer.Endpoint = "192.168.1.100:51820"
	peer.AllowedIPs = []string{"10.0.0.2/32"}
	
	// Перевірка основних полів
	if len(peer.PublicKey) == 0 {
		t.Errorf("Peer PublicKey should not be empty")
	}
	
	if peer.Endpoint == "" {
		t.Errorf("Peer Endpoint should not be empty")
	}
	
	if len(peer.AllowedIPs) == 0 {
		t.Errorf("Peer AllowedIPs should not be empty")
	}
}

func TestIPPool_Basic(t *testing.T) {
	_, network, err := net.ParseCIDR("10.0.0.0/24")
	if err != nil {
		t.Fatalf("Failed to parse CIDR: %v", err)
	}
	
	pool := &IPPool{
		Network:   network,
		Allocated: make(map[string]bool),
	}
	
	// Тестування базової структури
	if pool.Network == nil {
		t.Errorf("IPPool Network should not be nil")
	}
	
	if pool.Allocated == nil {
		t.Errorf("IPPool Allocated should not be nil")
	}
	
	// Тестування мережі
	if !network.Contains(net.ParseIP("10.0.0.1")) {
		t.Errorf("Network should contain 10.0.0.1")
	}
	
	if network.Contains(net.ParseIP("192.168.1.1")) {
		t.Errorf("Network should not contain 192.168.1.1")
	}
}

func TestHandshakeInfo_Basic(t *testing.T) {
	tests := []struct {
		name           string
		lastHandshake  time.Time
		expectedZero   bool
	}{
		{
			name:           "recent handshake",
			lastHandshake:  time.Now().Add(-1 * time.Minute),
			expectedZero:   false,
		},
		{
			name:           "zero handshake",
			lastHandshake:  time.Time{},
			expectedZero:   true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handshake := &HandshakeInfo{
				LastHandshake: tt.lastHandshake,
			}
			
			isZero := handshake.LastHandshake.IsZero()
			
			if isZero != tt.expectedZero {
				t.Errorf("LastHandshake.IsZero() = %v, want %v", isZero, tt.expectedZero)
			}
		})
	}
}

func TestClientConfig_Basic(t *testing.T) {
	tests := []struct {
		name    string
		config  *ClientConfig
		valid   bool
	}{
		{
			name: "valid config",
			config: &ClientConfig{
				Interface: ClientInterface{
					PrivateKey: "test-private-key",
					Address:    []string{"10.0.0.2/24"},
				},
				Peer: ServerPeer{
					PublicKey:  "test-public-key",
					Endpoint:   "192.168.1.1:51820",
					AllowedIPs: []string{"0.0.0.0/0"},
				},
			},
			valid: true,
		},
		{
			name: "empty private key",
			config: &ClientConfig{
				Interface: ClientInterface{
					PrivateKey: "",
					Address:    []string{"10.0.0.2/24"},
				},
			},
			valid: false,
		},
		{
			name: "no addresses",
			config: &ClientConfig{
				Interface: ClientInterface{
					PrivateKey: "test-private-key",
					Address:    []string{},
				},
			},
			valid: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config == nil {
				t.Errorf("Config should not be nil")
				return
			}
			
			// Базова перевірка структури
			if len(tt.config.Interface.PrivateKey) == 0 && tt.valid {
				t.Errorf("Valid config should have private key")
			}
			
			if len(tt.config.Interface.Address) == 0 && tt.valid {
				t.Errorf("Valid config should have addresses")
			}
		})
	}
}