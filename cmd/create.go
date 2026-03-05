package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"groq-cli/internal/executor"
	"groq-cli/internal/groq"
	"groq-cli/internal/ui"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [description]",
	Short: "🚀 Create a new project from a description",
	Long: `Generate a complete project structure from a natural language description.
Groq AI will create the files, directory structure, and code for you.

Examples:
  groq create "REST API in Go with SQLite"
  groq create "Python web scraper with BeautifulSoup"
  groq create "React todo app with TypeScript" --dir ./my-todo
  groq create "CLI tool in Rust for file compression"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey, err := ensureAPIKey()
		if err != nil {
			return err
		}

		model, _ := cmd.Flags().GetString("model")
		if model == "" {
			model = groq.DefaultModel()
		}

		outDir, _ := cmd.Flags().GetString("dir")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		execute, _ := cmd.Flags().GetBool("execute")

		var description string
		if len(args) > 0 {
			description = strings.Join(args, " ")
		} else {
			description, err = ui.PromptInput("📝 Describe the project you want to create: ")
			if err != nil {
				return err
			}
		}

		if description == "" {
			return fmt.Errorf("project description cannot be empty")
		}

		client := groq.NewClient(apiKey)

		// If no output dir specified, derive from description
		if outDir == "" {
			outDir = deriveProjectName(description)
		}

		absDir, err := filepath.Abs(outDir)
		if err != nil {
			return err
		}

		ui.PrintSection("🚀 Creating Project")
		ui.PrintInfo(fmt.Sprintf("Description: %s", description))
		ui.PrintInfo(fmt.Sprintf("Output dir:  %s", absDir))
		ui.PrintInfo(fmt.Sprintf("Model:       %s", model))
		fmt.Println()

		// Generate project
		project, err := ui.RunProjectGenerator(client, model, description, absDir, dryRun)
		if err != nil {
			return err
		}

		if !dryRun && execute && project != nil {
			fmt.Println()
			if confirm, _ := ui.PromptConfirm("⚡ Run the project now?"); confirm {
				return executor.RunProject(absDir, project.RunCommand)
			}
		}

		return nil
	},
}

func init() {
	createCmd.Flags().StringP("dir", "d", "", "Output directory (default: derived from description)")
	createCmd.Flags().BoolP("dry-run", "n", false, "Show what would be created without writing files")
	createCmd.Flags().BoolP("execute", "e", false, "Execute the project after creation")
	createCmd.Flags().StringP("model", "m", "", "Model to use")
}

func deriveProjectName(description string) string {
	words := strings.Fields(strings.ToLower(description))
	if len(words) == 0 {
		return "groq-project"
	}
	// Take first 3 meaningful words
	stopWords := map[string]bool{
		"a": true, "an": true, "the": true, "in": true,
		"with": true, "for": true, "and": true, "or": true,
		"using": true, "that": true, "to": true, "of": true,
	}
	var parts []string
	for _, w := range words {
		if !stopWords[w] {
			// Clean non-alphanumeric
			clean := strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
					return r
				}
				return '-'
			}, w)
			if clean != "" && clean != "-" {
				parts = append(parts, clean)
			}
		}
		if len(parts) == 3 {
			break
		}
	}
	if len(parts) == 0 {
		return "groq-project"
	}
	name := strings.Join(parts, "-")
	// Avoid overwriting existing dirs
	if _, err := os.Stat(name); err == nil {
		name = name + "-1"
	}
	return name
}
