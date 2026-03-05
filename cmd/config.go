package cmd

import (
	"fmt"

	"groq-cli/internal/config"
	"groq-cli/internal/ui"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "⚙️  Manage CLI configuration",
	Long:  `View and manage your Groq CLI configuration including API key and default model.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		ui.PrintConfigFull(cfg)
		return nil
	},
}

var configSetKeyCmd = &cobra.Command{
	Use:   "set-key [api-key]",
	Short: "Set your Groq API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		cfg.APIKey = args[0]
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		ui.PrintSuccess("✅ API key saved successfully!")
		ui.PrintInfo("You can now use all Groq CLI commands.")
		return nil
	},
}

var configSetModelCmd = &cobra.Command{
	Use:   "set-model [model]",
	Short: "Set the default model",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		cfg.DefaultModel = args[0]
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		ui.PrintSuccess(fmt.Sprintf("✅ Default model set to: %s", args[0]))
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		ui.PrintConfigFull(cfg)
		return nil
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	RunE: func(cmd *cobra.Command, args []string) error {
		confirm, err := ui.PromptConfirm("⚠️  Reset all configuration?")
		if err != nil || !confirm {
			ui.PrintInfo("Reset cancelled.")
			return nil
		}
		cfg := config.Default()
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		ui.PrintSuccess("✅ Configuration reset to defaults.")
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetKeyCmd)
	configCmd.AddCommand(configSetModelCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configResetCmd)
}
