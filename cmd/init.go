package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a default config file",
	Long:  `Creates ~/.chatatui.toml with default values if it does not already exist.`,
	Run: func(cmd *cobra.Command, args []string) {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: could not determine home directory: %v\n", err)
			os.Exit(1)
		}

		path := filepath.Join(home, ".chatatui.toml")

		if _, err := os.Stat(path); err == nil {
			fmt.Fprintf(os.Stderr, "config file already exists: %s\n", path)
			os.Exit(1)
		}

		const defaultConfig = `# chatatui configuration

# Client settings
host    = "localhost:8080"
api_key = ""

# Server settings
[server]
addr                  = ":8080"
database_dsn          = "postgres://root:password@localhost:5432/chatatui?sslmode=disable"
redis_url             = "redis://localhost:6379"
message_history_limit = 50
room_list_limit       = 100
rate_limit_requests   = 100
rate_limit_window_secs = 60
`

		if err := os.WriteFile(path, []byte(defaultConfig), 0o600); err != nil {
			fmt.Fprintf(os.Stderr, "error: could not write config file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("config file created: %s\n", path)
		fmt.Println("Next steps:")
		fmt.Println("  1. Start the server:  chatatui serve")
		fmt.Println("  2. Register a user:   curl -s -X POST http://localhost:8080/register -H 'Content-Type: application/json' -d '{\"name\":\"yourname\"}' | jq .")
		fmt.Println("  3. Set api_key in the config file")
		fmt.Println("  4. Start the client:  chatatui")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
