package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"groq-cli/internal/groq"
)

type ProjectSpec struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Language    string     `json:"language"`
	RunCommand  string     `json:"run_command"`
	Files       []FileSpec `json:"files"`
}

type FileSpec struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func RunProjectGenerator(client *groq.Client, model, description, outDir string, dryRun bool) (*ProjectSpec, error) {
	systemPrompt := `You are an expert software engineer. Generate complete, working project files based on the description provided.

CRITICAL: Respond ONLY with valid JSON matching this exact structure:
{
  "name": "project-name",
  "description": "Brief description",
  "language": "primary language",
  "run_command": "command to run the project",
  "files": [
    { "path": "relative/path/to/file.ext", "content": "complete file content here" }
  ]
}

Rules:
- Create ALL necessary files for a working project
- Include README.md with setup instructions
- Include proper dependency files (go.mod, package.json, requirements.txt, etc.)
- Write complete, production-quality code
- Always include a .gitignore
- The run_command should work after dependencies are installed
- Do NOT include any explanation outside the JSON`

	userPrompt := fmt.Sprintf("Create a complete, working project: %s\n\nOutput directory will be: %s", description, outDir)

	fmt.Println()
	printHRule()
	cGray.Printf("  ◆ ")
	cMint.Printf("GENERATING PROJECT\n")
	printHRule()
	fmt.Println()
	cGray.Printf("  · description  ")
	cWhite.Printf("%s\n", description)
	cGray.Printf("  · output       ")
	cWhite.Printf("%s\n", outDir)
	cGray.Printf("  · model        ")
	cWhite.Printf("%s\n", model)
	fmt.Println()

	done := make(chan bool)
	go showSpinner("generating structure", done)

	var rawResponse strings.Builder
	err := client.ChatStream(groq.ChatRequest{
		Model:       model,
		Messages:    []groq.Message{{Role: "system", Content: systemPrompt}, {Role: "user", Content: userPrompt}},
		MaxTokens:   8192,
		Temperature: 0.3,
		Stream:      true,
	}, func(chunk string) {
		rawResponse.WriteString(chunk)
	})

	done <- true
	fmt.Printf("\r                                      \r")

	if err != nil {
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	raw := strings.TrimSpace(rawResponse.String())
	raw = stripMarkdownFences(raw)

	var spec ProjectSpec
	if err := json.Unmarshal([]byte(raw), &spec); err != nil {
		start := strings.Index(raw, "{")
		end := strings.LastIndex(raw, "}")
		if start >= 0 && end > start {
			raw = raw[start : end+1]
			if err2 := json.Unmarshal([]byte(raw), &spec); err2 != nil {
				return nil, fmt.Errorf("failed to parse AI response: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to parse AI response: %w", err)
		}
	}

	// Show plan
	printHRule()
	cGray.Printf("  ◆ ")
	cMint.Printf("PROJECT PLAN\n")
	printHRule()
	fmt.Println()
	cGray.Printf("  ▸ %-14s", "name")
	cWhite.Printf("%s\n", spec.Name)
	cGray.Printf("  ▸ %-14s", "language")
	cWhite.Printf("%s\n", spec.Language)
	cGray.Printf("  ▸ %-14s", "run")
	cCode.Printf("%s\n", spec.RunCommand)
	fmt.Println()

	cGray.Printf("  ◆ FILES  ")
	cGray.Printf("(%d)\n", len(spec.Files))
	cGray.Printf("  %s\n", strings.Repeat("─", 60))
	for _, f := range spec.Files {
		cGray.Printf("  ▸ ")
		cMintDim.Printf("%-42s", f.Path)
		cGray.Printf("%d B\n", len(f.Content))
	}

	if dryRun {
		fmt.Println()
		cWarn.Println("  ▲ dry-run — no files written")
		fmt.Println()
		return &spec, nil
	}

	fmt.Println()
	confirm, err := PromptConfirm(fmt.Sprintf("create %d files in %s?", len(spec.Files), outDir))
	if err != nil || !confirm {
		cGray.Println("  · cancelled")
		return nil, nil
	}

	// Write files
	fmt.Println()
	printHRule()
	cGray.Printf("  ◆ ")
	cMint.Printf("WRITING FILES\n")
	printHRule()
	fmt.Println()

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output dir: %w", err)
	}

	for _, f := range spec.Files {
		fullPath := filepath.Join(outDir, f.Path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			PrintError(fmt.Sprintf("mkdir failed for %s: %v", f.Path, err))
			continue
		}
		if err := os.WriteFile(fullPath, []byte(f.Content), 0644); err != nil {
			PrintError(fmt.Sprintf("write failed %s: %v", f.Path, err))
			continue
		}
		cGray.Printf("  ▸ [✓] ")
		cWhite.Printf("%s\n", f.Path)
		time.Sleep(25 * time.Millisecond)
	}

	fmt.Println()
	printHRule()
	cMint.Printf("  ◆ done — project created\n")
	printHRule()
	fmt.Println()
	cGray.Printf("  · location  ")
	cWhite.Printf("%s\n", outDir)
	fmt.Println()
	cGray.Printf("  next steps\n")
	cGray.Printf("  ▸ ")
	cCode.Printf("cd %s\n", outDir)
	cGray.Printf("  ▸ ")
	cCode.Printf("%s\n", spec.RunCommand)
	fmt.Println()

	return &spec, nil
}

func stripMarkdownFences(s string) string {
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func showSpinner(label string, done chan bool) {
	frames := []string{"◐", "◓", "◑", "◒"}
	i := 0
	for {
		select {
		case <-done:
			return
		default:
			cMint.Printf("\r  %s %s", frames[i%len(frames)], label)
			time.Sleep(100 * time.Millisecond)
			i++
		}
	}
}
