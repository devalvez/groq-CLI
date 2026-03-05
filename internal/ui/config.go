package ui

import (
	"fmt"
	"strings"

	"groq-cli/internal/config"
	"groq-cli/internal/groq"
)

func PrintConfigFull(cfg *config.Config) {
	fmt.Println()
	printHRule()
	cGray.Printf("  ◆ ")
	cMint.Printf("CONFIGURATION\n")
	printHRule()
	fmt.Println()

	// API Key
	cGray.Printf("  ▸ %-20s", "api_key")
	if cfg.APIKey != "" {
		masked := cfg.APIKey[:4] + strings.Repeat("·", len(cfg.APIKey)-8) + cfg.APIKey[len(cfg.APIKey)-4:]
		cSuccess.Printf("%s\n", masked)
	} else {
		cError.Printf("not set\n")
	}

	cGray.Printf("  ▸ %-20s", "default_model")
	cWhite.Printf("%s\n", cfg.DefaultModel)

	cGray.Printf("  ▸ %-20s", "safe_mode")
	if cfg.SafeMode {
		cMint.Printf("enabled\n")
	} else {
		cWarn.Printf("disabled\n")
	}

	cGray.Printf("  ▸ %-20s", "stream_output")
	if cfg.StreamOutput {
		cMint.Printf("enabled\n")
	} else {
		cGray.Printf("disabled\n")
	}

	cGray.Printf("  ▸ %-20s", "theme")
	cWhite.Printf("%s\n", cfg.Theme)

	if path, err := config.ConfigPath(); err == nil {
		fmt.Println()
		cGray.Printf("  · config file  %s\n", path)
	}

	fmt.Println()
	printHRule()
	cGray.Printf("  ◆ COMMANDS\n")
	printHRule()
	fmt.Println()
	cGray.Printf("  ▸ ")
	cCode.Printf("groq config set-key YOUR_KEY")
	cGray.Printf("    set API key\n")
	cGray.Printf("  ▸ ")
	cCode.Printf("groq config set-model MODEL_ID")
	cGray.Printf("  set default model\n")
	cGray.Printf("  ▸ ")
	cCode.Printf("groq config reset")
	cGray.Printf("             restore defaults\n")
	fmt.Println()
}

func PrintModelsTable(models []groq.Model) {
	fmt.Println()
	printHRule()
	cGray.Printf("  ◆ ")
	cMint.Printf("AVAILABLE MODELS\n")
	printHRule()
	fmt.Println()

	cGray.Printf("  %-44s %s\n", "MODEL ID", "OWNER")
	cGray.Printf("  %s\n", strings.Repeat("─", 60))

	for _, m := range models {
		cGray.Printf("  ▸ ")
		cMint.Printf("%-42s", m.ID)
		cGray.Printf("%s\n", m.OwnedBy)
	}

	fmt.Println()
	cGray.Printf("  · ")
	cWhite.Printf("%d", len(models))
	cGray.Printf(" models available\n")
	fmt.Println()
	cGray.Printf("  usage  ")
	cCode.Printf("groq chat --model MODEL_ID\n")
	fmt.Println()
}
