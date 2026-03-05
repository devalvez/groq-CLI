package executor

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"groq-cli/internal/groq"
	"groq-cli/internal/ui"
)

// TaskResult is what the AI returns for a task
type TaskResult struct {
	Language string `json:"language"`
	Code     string `json:"code"`
	Filename string `json:"filename"`
	Command  string `json:"command"`
	Explain  string `json:"explain"`
}

// RunTask generates and executes code for a natural language task
func RunTask(client *groq.Client, model, task, preferredLang string, safeMode, showCode bool) error {
	systemPrompt := buildTaskSystemPrompt(preferredLang)

	userPrompt := fmt.Sprintf("Task: %s", task)

	ui.PrintSection("🧠 Generating Solution")
	fmt.Println()

	// Show spinner while generating
	done := make(chan bool)
	go showSpinner("Generating code", done)

	var rawResponse strings.Builder
	err := client.ChatStream(groq.ChatRequest{
		Model: model,
		Messages: []groq.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens:   4096,
		Temperature: 0.2,
		Stream:      true,
	}, func(chunk string) {
		rawResponse.WriteString(chunk)
	})

	done <- true
	fmt.Printf("\r                              \r")

	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	// Parse response
	raw := strings.TrimSpace(rawResponse.String())
	raw = stripMarkdown(raw)

	var result TaskResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		// Fallback: try to extract JSON
		start := strings.Index(raw, "{")
		end := strings.LastIndex(raw, "}")
		if start >= 0 && end > start {
			raw = raw[start : end+1]
			if err2 := json.Unmarshal([]byte(raw), &result); err2 != nil {
				return fmt.Errorf("failed to parse AI response: %v", err)
			}
		} else {
			return fmt.Errorf("failed to parse AI response: %v", err)
		}
	}

	// Display what was generated
	if result.Explain != "" {
		ui.PrintSection("💡 Solution")
		fmt.Println()
		fmt.Printf("  %s\n", result.Explain)
	}

	// Show code
	if showCode || safeMode {
		fmt.Println()
		ui.PrintSection(fmt.Sprintf("📄 Generated Code (%s)", result.Language))
		fmt.Println()
		printCodeBlock(result.Code, result.Language)
		fmt.Println()
		colorRun := fmt.Sprintf("  Run command: %s", result.Command)
		ui.PrintInfo(colorRun)
	}

	// Safe mode: confirm before running
	if safeMode {
		fmt.Println()
		confirm, err := ui.PromptConfirm("⚡ Execute this code?")
		if err != nil || !confirm {
			ui.PrintInfo("Execution cancelled.")
			return nil
		}
	}

	// Write and execute
	return executeTask(result)
}

func executeTask(result TaskResult) error {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "groq-task-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write the code file
	filename := result.Filename
	if filename == "" {
		filename = defaultFilename(result.Language)
	}
	codePath := filepath.Join(tmpDir, filename)

	if err := os.WriteFile(codePath, []byte(result.Code), 0755); err != nil {
		return fmt.Errorf("failed to write code: %w", err)
	}

	// Build command
	cmdStr := result.Command
	if cmdStr == "" {
		cmdStr = defaultRunCommand(result.Language, filename)
	}

	// Replace placeholder
	cmdStr = strings.ReplaceAll(cmdStr, "{file}", filename)
	cmdStr = strings.ReplaceAll(cmdStr, "{dir}", tmpDir)

	ui.PrintSection("⚡ Executing")
	fmt.Println()

	// Parse and run command
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return fmt.Errorf("empty run command")
	}

	start := time.Now()
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = tmpDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Run()
	elapsed := time.Since(start)

	fmt.Println()
	ui.PrintDivider()
	if err != nil {
		ui.PrintError(fmt.Sprintf("Execution failed: %v", err))
	} else {
		ui.PrintSuccess(fmt.Sprintf("✅ Completed in %s", elapsed.Round(time.Millisecond)))
	}
	ui.PrintDivider()

	return err
}

// RunProject runs a generated project
func RunProject(dir, runCmd string) error {
	if runCmd == "" {
		return fmt.Errorf("no run command specified")
	}

	ui.PrintSection("⚡ Running Project")
	fmt.Println()

	parts := strings.Fields(runCmd)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func buildTaskSystemPrompt(lang string) string {
	langHint := ""
	if lang != "" {
		langHint = fmt.Sprintf(" Prefer using %s.", lang)
	}

	return fmt.Sprintf(`You are an expert programmer that generates working code to accomplish tasks.%s

CRITICAL: Respond ONLY with valid JSON matching exactly this structure:
{
  "language": "bash|python|go|node",
  "filename": "script.sh",
  "code": "complete working code here",
  "command": "bash script.sh",
  "explain": "One sentence explaining what this does"
}

Rules:
- Write complete, working, self-contained code
- Prefer bash for simple file/system tasks
- Prefer python for data processing, math, web requests
- The code must work without external dependencies unless standard for the language
- command should be the exact command to run the file
- Do NOT add markdown or explanation outside the JSON`, langHint)
}

func printCodeBlock(code, lang string) {
	fmt.Printf("  ╭─ %s\n", lang)
	lines := strings.Split(code, "\n")
	// Show max 30 lines
	maxLines := 30
	if len(lines) <= maxLines {
		for _, line := range lines {
			fmt.Printf("  │ %s\n", line)
		}
	} else {
		for _, line := range lines[:maxLines] {
			fmt.Printf("  │ %s\n", line)
		}
		fmt.Printf("  │ ... (%d more lines)\n", len(lines)-maxLines)
	}
	fmt.Printf("  ╰─\n")
}

func stripMarkdown(s string) string {
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimSuffix(s, "```")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}

func defaultFilename(lang string) string {
	switch lang {
	case "python":
		return "task.py"
	case "go":
		return "main.go"
	case "node", "javascript":
		return "task.js"
	case "ruby":
		return "task.rb"
	default:
		return "task.sh"
	}
}

func defaultRunCommand(lang, filename string) string {
	switch lang {
	case "python":
		return "python3 " + filename
	case "go":
		return "go run " + filename
	case "node", "javascript":
		return "node " + filename
	case "ruby":
		return "ruby " + filename
	default:
		return "bash " + filename
	}
}

func showSpinner(label string, done chan bool) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	for {
		select {
		case <-done:
			return
		default:
			fmt.Printf("\r  %s %s...", frames[i%len(frames)], label)
			time.Sleep(80 * time.Millisecond)
			i++
		}
	}
}
