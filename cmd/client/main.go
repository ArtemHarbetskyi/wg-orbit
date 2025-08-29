package main

import (
	"fmt"
	"log"
	"os"

	"github.com/artem/wg-orbit/internal/client"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wg-orbit-client",
	Short: "WireGuard Orbit Client - Automated WireGuard client management",
	Long:  `WireGuard Orbit Client automatically enrolls with the server and manages WireGuard configuration.`,
}

var enrollCmd = &cobra.Command{
	Use:   "enroll",
	Short: "Enroll client with WireGuard Orbit server",
	Run: func(cmd *cobra.Command, args []string) {
		server, _ := cmd.Flags().GetString("server")
		token, _ := cmd.Flags().GetString("token")
		name, _ := cmd.Flags().GetString("name")

		// Створюємо клієнт
		config := client.DefaultConfig()
		cli := client.NewClient(config)

		// Реєструємо клієнта
		if err := cli.Enroll(server, token, name); err != nil {
			log.Fatalf("Failed to enroll client: %v", err)
		}

		fmt.Printf("Client '%s' enrolled successfully with server %s\n", name, server)
	},
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Bring up WireGuard connection",
	Run: func(cmd *cobra.Command, args []string) {
		// Створюємо клієнт
		config := client.DefaultConfig()
		cli := client.NewClient(config)

		// Завантажуємо існуючу конфігурацію
		if err := cli.LoadConfig(); err != nil {
			log.Fatalf("Failed to load config: %v. Run 'enroll' first", err)
		}

		// Піднімаємо з'єднання
		if err := cli.Up(); err != nil {
			log.Fatalf("Failed to bring up connection: %v", err)
		}
	},
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Bring down WireGuard connection",
	Run: func(cmd *cobra.Command, args []string) {
		// Створюємо клієнт
		config := client.DefaultConfig()
		cli := client.NewClient(config)

		// Завантажуємо існуючу конфігурацію
		if err := cli.LoadConfig(); err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		// Опускаємо з'єднання
		if err := cli.Down(); err != nil {
			log.Fatalf("Failed to bring down connection: %v", err)
		}
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show WireGuard connection status",
	Run: func(cmd *cobra.Command, args []string) {
		// Створюємо клієнт
		config := client.DefaultConfig()
		cli := client.NewClient(config)

		// Завантажуємо існуючу конфігурацію
		if err := cli.LoadConfig(); err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		// Показуємо статус
		if err := cli.Status(); err != nil {
			log.Fatalf("Failed to get status: %v", err)
		}
	},
}

func init() {
	// Enroll command flags
	enrollCmd.Flags().StringP("server", "s", "", "WireGuard Orbit server URL (required)")
	enrollCmd.Flags().StringP("token", "t", "", "Enrollment token (required)")
	enrollCmd.Flags().StringP("name", "n", "", "Client name (required)")
	if err := enrollCmd.MarkFlagRequired("server"); err != nil {
		log.Fatalf("Failed to mark server flag as required: %v", err)
	}
	if err := enrollCmd.MarkFlagRequired("token"); err != nil {
		log.Fatalf("Failed to mark token flag as required: %v", err)
	}
	if err := enrollCmd.MarkFlagRequired("name"); err != nil {
		log.Fatalf("Failed to mark name flag as required: %v", err)
	}

	// Up command flags
	upCmd.Flags().StringP("config", "c", "/etc/wg-orbit/client.conf", "WireGuard configuration file")

	// Down command flags
	downCmd.Flags().StringP("interface", "i", "wg0", "WireGuard interface name")

	// Add commands to root
	rootCmd.AddCommand(enrollCmd, upCmd, downCmd, statusCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
