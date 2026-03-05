package cmd

import (
	"fmt"
	"strings"

	"groq-cli/internal/groq"
	"groq-cli/internal/ui"

	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat [message]",
	Short: "💬 Start an interactive chat or send a single message",
	Long: `Start an interactive chat session with Groq AI, or pass a message directly.

Examples:
  groq chat                                    # Interactive session
  groq chat "What is Go?"                      # Single message, bordered UI
  groq chat --plain "What is Go?"              # Raw text only — no borders
  groq chat --copy  "Explain goroutines"       # Response copied to clipboard
  groq chat --plain "list 5 tips" | tee tips.txt`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey, err := ensureAPIKey()
		if err != nil {
			return err
		}

		model, _ := cmd.Flags().GetString("model")
		if model == "" {
			model = groq.DefaultModel()
		}

		maxTokens, _ := cmd.Flags().GetInt("max-tokens")
		temperature, _ := cmd.Flags().GetFloat64("temperature")
		plain, _ := cmd.Flags().GetBool("plain")
		copy_, _ := cmd.Flags().GetBool("copy")

		// Resolve output mode — --plain takes priority over --copy
		mode := ui.ModeDefault
		switch {
		case plain:
			mode = ui.ModePlain
		case copy_:
			mode = ui.ModeCopy
		}

		client := groq.NewClient(apiKey)

		// Interactive mode — flags other than model are ignored
		if len(args) == 0 {
			if plain || copy_ {
				return fmt.Errorf("--plain and --copy require a message argument\n  Usage: groq chat --plain \"your question\"")
			}
			return ui.RunInteractiveChat(client, model)
		}

		// Single-shot mode
		message := strings.Join(args, " ")
		opts := ui.ChatOptions{
			Model:       model,
			MaxTokens:   maxTokens,
			Temperature: temperature,
			Mode:        mode,
		}
		return ui.RunSingleChat(client, opts, message)
	},
}

func init() {
	chatCmd.Flags().StringP("model", "m", "", "Model to use")
	chatCmd.Flags().IntP("max-tokens", "t", 8192, "Max tokens in response")
	chatCmd.Flags().Float64P("temperature", "T", 0.7, "Temperature (0.0-2.0)")
	chatCmd.Flags().Bool("plain", false, "Output raw text — no borders or colors")
	chatCmd.Flags().Bool("copy", false, "Copy response to clipboard (requires xclip/xsel/wl-copy)")

	// --plain and --copy are mutually exclusive
	chatCmd.MarkFlagsMutuallyExclusive("plain", "copy")
}

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "📋 List available Groq models",
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey, err := ensureAPIKey()
		if err != nil {
			return err
		}

		client := groq.NewClient(apiKey)
		models, err := client.ListModels()
		if err != nil {
			return fmt.Errorf("failed to fetch models: %w", err)
		}

		ui.PrintModelsTable(models)
		return nil
	},
}
