package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/EwanGreer/chatatui/internal/config"
	"github.com/EwanGreer/chatatui/internal/inspect"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Browse Redis keys for debugging",
	Run: func(cmd *cobra.Command, args []string) {
		redisURL, _ := cmd.Flags().GetString("redis-url")
		if redisURL == "" {
			cfg, err := config.LoadServerConfig()
			if err != nil {
				fmt.Fprintln(os.Stderr, "error: could not load config:", err)
				os.Exit(1)
			}
			redisURL = cfg.RedisURL
		}

		if redisURL == "" {
			fmt.Fprintln(os.Stderr, "error: no redis URL configured — set server.redis_url in config or use --redis-url")
			os.Exit(1)
		}

		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error: invalid redis URL:", err)
			os.Exit(1)
		}

		rdb := redis.NewClient(opt)
		defer func() { _ = rdb.Close() }()

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := rdb.Ping(ctx).Err(); err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot reach Redis at %s: %v\n", redisURL, err)
			os.Exit(1)
		}

		if len(os.Getenv("DEBUG")) > 0 {
			f, err := tea.LogToFile("debug.log", "debug")
			if err != nil {
				fmt.Fprintln(os.Stderr, "fatal:", err)
				os.Exit(1)
			}
			defer func() { _ = f.Close() }()
		}

		model := inspect.NewModel(rdb)
		if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	inspectCmd.Flags().String("redis-url", "", "Redis URL (overrides config)")
	rootCmd.AddCommand(inspectCmd)
}
