package server

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"

	"github.com/artem/wg-orbit/internal/wg"
)

// InterfaceManager керує WireGuard інтерфейсом
type InterfaceManager struct {
	interfaceName string
	privateKey    string
	publicKey     string
	listenPort    int
	ipPool        *wg.IPPool
}

// NewInterfaceManager створює новий менеджер інтерфейсу
func NewInterfaceManager(interfaceName, network string) (*InterfaceManager, error) {
	// Генеруємо ключі для інтерфейсу
	privateKey, publicKey, err := wg.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate keys: %w", err)
	}

	// Створюємо IP пул
	ipPool, err := wg.NewIPPool(network)
	if err != nil {
		return nil, fmt.Errorf("failed to create IP pool: %w", err)
	}

	return &InterfaceManager{
		interfaceName: interfaceName,
		privateKey:    privateKey,
		publicKey:     publicKey,
		listenPort:    51820, // Стандартний порт WireGuard
		ipPool:        ipPool,
	}, nil
}

// CreateInterface створює WireGuard інтерфейс
func (im *InterfaceManager) CreateInterface() error {
	// Перевіряємо, чи інтерфейс вже існує
	if im.interfaceExists() {
		return fmt.Errorf("interface %s already exists", im.interfaceName)
	}

	// Створюємо інтерфейс
	cmd := exec.Command("ip", "link", "add", "dev", im.interfaceName, "type", "wireguard")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create interface: %w", err)
	}

	// Встановлюємо приватний ключ
	cmd = exec.Command("wg", "set", im.interfaceName, "private-key", "/dev/stdin")
	cmd.Stdin = strings.NewReader(im.privateKey)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set private key: %w", err)
	}

	// Встановлюємо порт
	cmd = exec.Command("wg", "set", im.interfaceName, "listen-port", strconv.Itoa(im.listenPort))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set listen port: %w", err)
	}

	// Встановлюємо IP адресу інтерфейсу
	networkIP := im.ipPool.Network.IP.String()
	mask, _ := im.ipPool.Network.Mask.Size()
	interfaceAddr := fmt.Sprintf("%s/%d", networkIP, mask)
	
	cmd = exec.Command("ip", "addr", "add", interfaceAddr, "dev", im.interfaceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set interface address: %w", err)
	}

	// Піднімаємо інтерфейс
	cmd = exec.Command("ip", "link", "set", "up", "dev", im.interfaceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring up interface: %w", err)
	}

	return nil
}

// interfaceExists перевіряє, чи існує інтерфейс
func (im *InterfaceManager) interfaceExists() bool {
	cmd := exec.Command("ip", "link", "show", im.interfaceName)
	return cmd.Run() == nil
}

// AllocateIP виділяє нову IP адресу
func (im *InterfaceManager) AllocateIP() (net.IP, error) {
	return im.ipPool.AllocateIP()
}

// ReleaseIP звільняє IP адресу
func (im *InterfaceManager) ReleaseIP(ip net.IP) {
	im.ipPool.ReleaseIP(ip)
}

// PublicKey повертає публічний ключ інтерфейсу
func (im *InterfaceManager) PublicKey() string {
	return im.publicKey
}

// PrivateKey повертає приватний ключ інтерфейсу
func (im *InterfaceManager) PrivateKey() string {
	return im.privateKey
}

// ListenPort повертає порт прослуховування
func (im *InterfaceManager) ListenPort() int {
	return im.listenPort
}