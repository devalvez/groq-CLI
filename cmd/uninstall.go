package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"groq-cli/internal/config"
	"groq-cli/internal/ui"

	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "🗑  Remove Groq CLI from the system",
	Long:  `Uninstall Groq CLI binary and optionally remove all configuration files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		keepConfig, _ := cmd.Flags().GetBool("keep-config")
		force, _ := cmd.Flags().GetBool("force")

		return runUninstall(keepConfig, force)
	},
}

func init() {
	uninstallCmd.Flags().Bool("keep-config", false, "Keep configuration files and API key")
	uninstallCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall(keepConfig, force bool) error {
	// Find binary path
	binaryPath, err := findBinaryPath()
	if err != nil || binaryPath == "" {
		ui.PrintWarning("Could not locate the groq binary in PATH.")
	}

	// Find config path
	cfgPath, _ := config.ConfigPath()
	cfgDir := filepath.Dir(cfgPath)

	// Show what will be removed
	ui.PrintSection("🗑  Uninstall Groq CLI")
	fmt.Println()

	if binaryPath != "" {
		fmt.Printf("  ")
		ui.PrintInfo(fmt.Sprintf("Binary:  %s", binaryPath))
	}
	if !keepConfig {
		if _, err := os.Stat(cfgDir); err == nil {
			fmt.Printf("  ")
			ui.PrintInfo(fmt.Sprintf("Config:  %s", cfgDir))
		}
	}

	fmt.Println()

	// Confirm
	if !force {
		confirm, err := ui.PromptConfirm("⚠️  Proceed with uninstall?")
		if err != nil || !confirm {
			ui.PrintInfo("Uninstall cancelled.")
			return nil
		}

		if !keepConfig {
			fmt.Println()
			keepCfg, _ := ui.PromptConfirm("💾 Keep your config and saved API key?")
			if keepCfg {
				keepConfig = true
			}
		}
	}

	fmt.Println()
	ui.PrintSection("🗑  Removing")
	fmt.Println()

	// Remove binary
	if binaryPath != "" {
		if err := removeFile(binaryPath); err != nil {
			ui.PrintError(fmt.Sprintf("Failed to remove binary: %v", err))
			ui.PrintInfo("Try running with sudo, or manually: sudo rm " + binaryPath)
		} else {
			ui.PrintSuccess(fmt.Sprintf("✓ Removed binary: %s", binaryPath))
		}
	}

	// Remove config
	if !keepConfig {
		if _, err := os.Stat(cfgDir); err == nil {
			if err := os.RemoveAll(cfgDir); err != nil {
				ui.PrintError(fmt.Sprintf("Failed to remove config: %v", err))
			} else {
				ui.PrintSuccess(fmt.Sprintf("✓ Removed config: %s", cfgDir))
			}
		}
	} else {
		ui.PrintInfo(fmt.Sprintf("Config kept at: %s", cfgDir))
	}

	fmt.Println()
	ui.PrintDivider()

	// Verify removal
	if _, err := exec.LookPath("groq"); err != nil {
		ui.PrintSuccess("✅ Groq CLI successfully uninstalled!")
	} else {
		ui.PrintWarning("'groq' still found in PATH — you may need to remove it manually.")
		remaining, _ := exec.LookPath("groq")
		ui.PrintInfo("Remaining: " + remaining)
	}

	ui.PrintDivider()
	fmt.Println()
	fmt.Println("  Thanks for using Groq CLI! 👋")
	fmt.Println()

	return nil
}

func findBinaryPath() (string, error) {
	// First check common install locations
	candidates := []string{
		"/usr/local/bin/groq",
		"/usr/bin/groq",
		filepath.Join(os.Getenv("HOME"), ".local/bin/groq"),
		filepath.Join(os.Getenv("GOPATH"), "bin/groq"),
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	// Fall back to which
	path, err := exec.LookPath("groq")
	if err != nil {
		return "", err
	}
	return path, nil
}

func removeFile(path string) error {
	err := os.Remove(path)
	if err == nil {
		return nil
	}

	// If permission denied, try with sudo
	if strings.Contains(err.Error(), "permission denied") {
		sudoCmd := exec.Command("sudo", "rm", "-f", path)
		sudoCmd.Stdout = os.Stdout
		sudoCmd.Stderr = os.Stderr
		sudoCmd.Stdin = os.Stdin
		return sudoCmd.Run()
	}

	return err
}
