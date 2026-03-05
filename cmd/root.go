package cmd

import (
	"fmt"
	"os"

	"groq-cli/internal/config"
	"groq-cli/internal/ui"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "groq",
	Short: "⚡ Groq CLI — AI-powered terminal assistant",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		return ui.RunWelcome()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(chatCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(modelsCmd)

	rootCmd.PersistentFlags().String("model", "", "Groq model to use (overrides config)")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable colored output")
}

func initConfig() {
	if err := config.Load(); err != nil {
		// Config may not exist yet — that's fine
		_ = err
	}

	// Check for API key
	if os.Getenv("GROQ_API_KEY") == "" {
		cfg := config.Get()
		if cfg.APIKey == "" {
			// Will be handled by individual commands
		}
	}
}

// ensureAPIKey checks for an API key and prompts if missing
func ensureAPIKey() (string, error) {
	if key := os.Getenv("GROQ_API_KEY"); key != "" {
		return key, nil
	}
	cfg := config.Get()
	if cfg.APIKey != "" {
		return cfg.APIKey, nil
	}
	fmt.Println()
	ui.PrintWarning("No API key found. Set it with:")
	ui.PrintCode("  groq config set-key YOUR_API_KEY")
	ui.PrintInfo("Or export it: export GROQ_API_KEY=your_key")
	fmt.Println()
	return "", fmt.Errorf("GROQ_API_KEY not set")
}
