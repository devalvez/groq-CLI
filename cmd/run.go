package cmd

import (
	"fmt"
	"strings"

	"groq-cli/internal/executor"
	"groq-cli/internal/groq"
	"groq-cli/internal/ui"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [task]",
	Short: "⚡ Execute a task or program via natural language",
	Long: `Describe what you want to do and Groq AI will generate and run the code.
The AI can write scripts, run shell commands, process files, and more.

Examples:
  groq run "list all .go files in current dir sorted by size"
  groq run "find duplicate files in /tmp"
  groq run "create a bash script that monitors CPU usage"
  groq run "write and run a Python script to generate a Fibonacci sequence"
  groq run --lang python "sort a list of numbers and plot a histogram"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey, err := ensureAPIKey()
		if err != nil {
			return err
		}

		model, _ := cmd.Flags().GetString("model")
		if model == "" {
			model = groq.DefaultModel()
		}

		lang, _ := cmd.Flags().GetString("lang")
		safeMode, _ := cmd.Flags().GetBool("safe")
		showCode, _ := cmd.Flags().GetBool("show-code")

		var task string
		if len(args) > 0 {
			task = strings.Join(args, " ")
		} else {
			task, err = ui.PromptInput("⚡ What task should I perform? ")
			if err != nil {
				return err
			}
		}

		if task == "" {
			return fmt.Errorf("task description cannot be empty")
		}

		client := groq.NewClient(apiKey)

		ui.PrintSection("⚡ Task Runner")
		ui.PrintInfo(fmt.Sprintf("Task:  %s", task))
		if lang != "" {
			ui.PrintInfo(fmt.Sprintf("Lang:  %s", lang))
		}
		ui.PrintInfo(fmt.Sprintf("Model: %s", model))
		fmt.Println()

		return executor.RunTask(client, model, task, lang, safeMode, showCode)
	},
}

func init() {
	runCmd.Flags().StringP("lang", "l", "", "Preferred language (bash, python, go, node)")
	runCmd.Flags().Bool("safe", true, "Preview code before executing (recommended)")
	runCmd.Flags().Bool("show-code", false, "Always show generated code")
	runCmd.Flags().StringP("model", "m", "", "Model to use")
}
