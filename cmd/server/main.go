// Package main реалізує CLI інтерфейс для WireGuard Orbit Server
//
// Цей пакет надає командний рядок для управління WireGuard сервером,
// включаючи ініціалізацію інтерфейсів, управління користувачами та генерацію токенів.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/artem/wg-orbit/internal/server"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// rootCmd - коренева команда CLI для wg-orbit-server
var rootCmd = &cobra.Command{
	Use:   "wg-orbit-server",
	Short: "WireGuard Orbit Server - DevOps-friendly WireGuard management",
	Long: `WireGuard Orbit Server provides automated WireGuard peer management with REST API, 
JWT authentication, and centralized configuration.

Available commands:
  init        - Initialize WireGuard interface
  run         - Start the server
  user        - User management
  user add    - Add new user
  user token  - Generate user token
  user enroll-token - Generate enrollment token`,
}

// initCmd - команда для ініціалізації WireGuard інтерфейсу
// Створює новий WireGuard інтерфейс з заданим ім'ям та базовою конфігурацією
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize WireGuard interface",
	Long: `Initializes a new WireGuard interface to work with wg-orbit.

This command creates a new WireGuard interface with the specified name,
configures basic parameters, and prepares the system for server operation.

Example:
  wg-orbit-server init --interface wg0 --config /etc/wg-orbit/server.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		interface_name, _ := cmd.Flags().GetString("interface")
		configPath, _ := cmd.Flags().GetString("config")

		// Створюємо базову конфігурацію сервера
		config := server.DefaultConfig()
		config.Interface = interface_name

		// Завантажуємо конфігурацію з файлу, якщо вказано шлях
		if configPath != "" {
			if err := loadConfigFromFile(config, configPath); err != nil {
				log.Printf("Warning: failed to load config from %s: %v", configPath, err)
			}
		}

		// Ініціалізуємо сервер з конфігурацією
		srv, err := server.NewServer(config)
		if err != nil {
			log.Fatalf("Failed to create server: %v", err)
		}

		// Ініціалізуємо WireGuard інтерфейс
		if err := srv.Initialize(); err != nil {
			log.Fatalf("Failed to initialize interface: %v", err)
		}

		fmt.Printf("WireGuard interface %s initialized successfully\n", interface_name)
	},
}

// runCmd - команда для запуску WireGuard Orbit сервера
// Запускає HTTP сервер з REST API для управління WireGuard peer'ами
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the WireGuard Orbit server",
	Long: `Launches WireGuard Orbit server with REST API.

The server provides an HTTP REST API for:
- User management (add, remove)
- Token generation for clients
- Connection status monitoring
- Automatic IP address allocation

Example:
  wg-orbit-server run --port 8080 --config /etc/wg-orbit/server.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetString("port")
		configPath, _ := cmd.Flags().GetString("config")

		// Створюємо базову конфігурацію сервера
		config := server.DefaultConfig()

		// Завантажуємо конфігурацію з файлу, якщо вказано
		if configPath != "" {
			if err := loadConfigFromFile(config, configPath); err != nil {
				log.Printf("Warning: failed to load config from %s: %v", configPath, err)
			}
		}

		// Перевизначаємо порт з командного рядка, якщо вказано
		if port != "" {
			if _, err := fmt.Sscanf(port, "%d", &config.Port); err != nil {
				log.Fatalf("Invalid port format: %v", err)
			}
		}

		// Ініціалізуємо сервер
		srv, err := server.NewServer(config)
		if err != nil {
			log.Fatalf("Failed to create server: %v", err)
		}

		// Запускаємо HTTP сервер
		if err := srv.Run(); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	},
}

// userCmd - група команд для управління користувачами
// Включає підкоманди для додавання користувачів та генерації токенів
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "User management commands",
	Long: `Commands for managing users in WireGuard Orbit server.

Available subcommands:
  add         - Add a new user
  token       - Generate token for existing user
  enroll-token - Generate enrollment token for new user`,
}

// addUserCmd - команда для додавання нового користувача
// Створює запис користувача в базі даних без генерації токену
var addUserCmd = &cobra.Command{
	Use:   "add [username]",
	Short: "Add a new user",
	Long: `Adds a new user to WireGuard Orbit system.

This command creates a user record in the database, but does not generate any tokens.
To generate tokens, use 'token' or 'enroll-token' subcommands.

Arguments:
  username - user name (required)

Example:
  wg-orbit-server user add alice --config /etc/wg-orbit/server.yaml`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		configPath, _ := cmd.Flags().GetString("config")

		// Створюємо базову конфігурацію сервера
		config := server.DefaultConfig()

		// Завантажуємо конфігурацію з файлу, якщо вказано
		if configPath != "" {
			if err := loadConfigFromFile(config, configPath); err != nil {
				log.Printf("Warning: failed to load config from %s: %v", configPath, err)
			}
		}

		// Ініціалізуємо сервер
		srv, err := server.NewServer(config)
		if err != nil {
			log.Fatalf("Failed to create server: %v", err)
		}

		// Додаємо користувача до системи
		if err := srv.AddUser(username); err != nil {
			log.Fatalf("Failed to add user: %v", err)
		}

		fmt.Printf("User %s added successfully\n", username)
	},
}

// tokenCmd - команда для генерації регулярного токену доступу
// Генерує JWT токен для існуючого користувача для автентифікації клієнта
var tokenCmd = &cobra.Command{
	Use:   "token [username]",
	Short: "Generate a token for a user",
	Long: `Generates a regular JWT token for an existing user.

This token is used for client authentication when receiving
configuration updates or reconnection.

Differently from enroll-token, this token is not intended for initial registration.

Arguments:
  username - existing user name (required)

Example:
  wg-orbit-server user token alice --config /etc/wg-orbit/server.yaml`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		configPath, _ := cmd.Flags().GetString("config")

		// Створюємо базову конфігурацію сервера
		config := server.DefaultConfig()

		// Завантажуємо конфігурацію з файлу, якщо вказано
		if configPath != "" {
			if err := loadConfigFromFile(config, configPath); err != nil {
				log.Printf("Warning: failed to load config from %s: %v", configPath, err)
			}
		}

		// Ініціалізуємо сервер
		srv, err := server.NewServer(config)
		if err != nil {
			log.Fatalf("Failed to create server: %v", err)
		}

		// Генеруємо токен доступу для користувача
		token, err := srv.GenerateToken(username)
		if err != nil {
			log.Fatalf("Failed to generate token: %v", err)
		}

		fmt.Printf("Token for user %s: %s\n", username, token)
	},
}

// enrollTokenCmd - команда для генерації enrollment токену
// Генерує одноразовий токен для первинної реєстрації нового клієнта
var enrollTokenCmd = &cobra.Command{
	Use:   "enroll-token [username]",
	Short: "Generate an enrollment token for a user",
	Long: `Generates an enrollment token for a new user.

This token is used for initial client registration in the system.
Client uses this token to get initial WireGuard configuration.

Enrollment token has a limited lifetime and can be used only once.

Arguments:
  username - new user name (required)

Example:
  wg-orbit-server user enroll-token alice --config /etc/wg-orbit/server.yaml`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		configPath, _ := cmd.Flags().GetString("config")

		// Створюємо базову конфігурацію сервера
		config := server.DefaultConfig()

		// Завантажуємо конфігурацію з файлу, якщо вказано
		if configPath != "" {
			if err := loadConfigFromFile(config, configPath); err != nil {
				log.Printf("Warning: failed to load config from %s: %v", configPath, err)
			}
		}

		// Ініціалізуємо сервер
		srv, err := server.NewServer(config)
		if err != nil {
			log.Fatalf("Failed to create server: %v", err)
		}

		// Генеруємо enrollment токен для первинної реєстрації
		token, err := srv.GenerateEnrollmentToken(username)
		if err != nil {
			log.Fatalf("Failed to generate enrollment token: %v", err)
		}

		fmt.Printf("Enrollment token for user %s: %s\n", username, token)
	},
}

// loadConfigFromFile завантажує конфігурацію з YAML файлу
func loadConfigFromFile(config *server.Config, configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Створюємо структуру для парсингу YAML
	var yamlConfig struct {
		WireGuard struct {
			Interface string `yaml:"interface"`
		} `yaml:"wireguard"`
		Storage struct {
			Type     string `yaml:"type"`
			Database string `yaml:"database"`
		} `yaml:"storage"`
		Server struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		} `yaml:"server"`
	}

	if err := yaml.Unmarshal(data, &yamlConfig); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Оновлюємо конфігурацію
	if yamlConfig.WireGuard.Interface != "" {
		config.Interface = yamlConfig.WireGuard.Interface
	}
	if yamlConfig.Storage.Type != "" {
		config.StorageConfig.Type = yamlConfig.Storage.Type
	}
	if yamlConfig.Storage.Database != "" {
		config.StorageConfig.Database = yamlConfig.Storage.Database
	}
	if yamlConfig.Server.Host != "" {
		config.Host = yamlConfig.Server.Host
	}
	if yamlConfig.Server.Port != 0 {
		config.Port = yamlConfig.Server.Port
	}

	return nil
}

func init() {
	// Init command flags
	initCmd.Flags().StringP("interface", "i", "wg0", "WireGuard interface name")
	initCmd.Flags().StringP("config", "c", "/etc/wg-orbit/server.yaml", "Configuration file path")

	// Run command flags
	runCmd.Flags().StringP("port", "p", "8080", "Server port")
	runCmd.Flags().StringP("config", "c", "/etc/wg-orbit/server.yaml", "Configuration file path")

	// User command flags
	userCmd.PersistentFlags().StringP("config", "c", "/etc/wg-orbit/server.yaml", "Configuration file path")

	// Add subcommands
	userCmd.AddCommand(addUserCmd, tokenCmd, enrollTokenCmd)
	rootCmd.AddCommand(initCmd, runCmd, userCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
