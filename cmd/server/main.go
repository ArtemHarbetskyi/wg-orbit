package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/artem/wg-orbit/internal/server"
)

var rootCmd = &cobra.Command{
	Use:   "wg-orbit-server",
	Short: "WireGuard Orbit Server - DevOps-friendly WireGuard management",
	Long:  `WireGuard Orbit Server provides automated WireGuard peer management with REST API, JWT authentication, and centralized configuration.`,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize WireGuard interface",
	Run: func(cmd *cobra.Command, args []string) {
		interface_name, _ := cmd.Flags().GetString("interface")
		
		// Створюємо конфігурацію сервера
		config := server.DefaultConfig()
		config.Interface = interface_name
		
		// Створюємо сервер
		srv, err := server.NewServer(config)
		if err != nil {
			log.Fatalf("Failed to create server: %v", err)
		}
		
		// Ініціалізуємо інтерфейс
		if err := srv.Initialize(); err != nil {
			log.Fatalf("Failed to initialize interface: %v", err)
		}
		
		fmt.Printf("WireGuard interface %s initialized successfully\n", interface_name)
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the WireGuard Orbit server",
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetString("port")
		
		// Створюємо конфігурацію сервера
		config := server.DefaultConfig()
		if port != "" {
			fmt.Sscanf(port, "%d", &config.Port)
		}
		
		// Створюємо сервер
		srv, err := server.NewServer(config)
		if err != nil {
			log.Fatalf("Failed to create server: %v", err)
		}
		
		// Запускаємо сервер
		if err := srv.Run(); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	},
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "User management commands",
}

var addUserCmd = &cobra.Command{
	Use:   "add [username]",
	Short: "Add a new user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		
		// Створюємо конфігурацію сервера
		config := server.DefaultConfig()
		
		// Створюємо сервер
		srv, err := server.NewServer(config)
		if err != nil {
			log.Fatalf("Failed to create server: %v", err)
		}
		
		// Додаємо користувача
		if err := srv.AddUser(username); err != nil {
			log.Fatalf("Failed to add user: %v", err)
		}
		
		fmt.Printf("User %s added successfully\n", username)
	},
}

var tokenCmd = &cobra.Command{
	Use:   "token [username]",
	Short: "Generate a token for a user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		
		// Створюємо конфігурацію сервера
		config := server.DefaultConfig()
		
		// Створюємо сервер
		srv, err := server.NewServer(config)
		if err != nil {
			log.Fatalf("Failed to create server: %v", err)
		}
		
		// Генеруємо токен
		token, err := srv.GenerateToken(username)
		if err != nil {
			log.Fatalf("Failed to generate token: %v", err)
		}
		
		fmt.Printf("Token for user %s: %s\n", username, token)
	},
}

func init() {
	// Init command flags
	initCmd.Flags().StringP("interface", "i", "wg0", "WireGuard interface name")
	
	// Run command flags
	runCmd.Flags().StringP("port", "p", "8080", "Server port")
	runCmd.Flags().StringP("config", "c", "/etc/wg-orbit/server.yaml", "Configuration file path")
	
	// Add subcommands
	userCmd.AddCommand(addUserCmd, tokenCmd)
	rootCmd.AddCommand(initCmd, runCmd, userCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}