package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

// ── Palette: dark bg, mint-green accent, dim gray, off-white text ──────────
var (
	isTTY = isatty.IsTerminal(os.Stdout.Fd())

	// Core
	cMint    = color.New(color.FgHiGreen, color.Bold) // ◆ active / accent
	cMintDim = color.New(color.FgGreen)               // softer mint
	cWhite   = color.New(color.FgHiWhite)             // primary text
	cGray    = color.New(color.FgHiBlack)             // dim / muted
	cGrayMid = color.New(color.FgWhite)               // secondary text

	// Semantic
	cSuccess = color.New(color.FgHiGreen, color.Bold)
	cError   = color.New(color.FgHiRed, color.Bold)
	cWarn    = color.New(color.FgYellow)
	cInfo    = color.New(color.FgHiBlack)

	// Chat
	cUser      = color.New(color.FgHiWhite, color.Bold)
	cAssistant = color.New(color.FgHiGreen)
	cCode      = color.New(color.FgHiGreen)

	// Aliases used by other files
	colorPrimary   = cMint
	colorSecondary = cGrayMid
	colorAccent    = cMint
	colorSuccess   = cSuccess
	colorError     = cError
	colorWarning   = cWarn
	colorInfo      = cInfo
	colorDim       = cGray
	colorWhite     = cWhite
	colorCode      = cCode
	colorUser      = cUser
	colorAssistant = cAssistant
)

const version = "v1.0.0"
const devName = "Wesley Alves"
const devHandle = "Devalvez"

// logo — compact wordmark in monospace block style
const logo = `
   ______  ______  ______  ______     ______  __      __
  /\  ___\/\  == \/\  __ \/\  __ \   /\  ___\/\ \    /\ \
  \ \ \__ \ \  __<\ \ \/\ \ \ \/\_\  \ \ \___\ \ \___\ \ \
   \ \_____\ \_\ \_\ \_____\ \___\_\  \ \_____\ \_____\ \_\
    \/_____/\/_/ /_/\/_____/\/___/_/   \/_____/\/_____/\/_/`

// ── Welcome ────────────────────────────────────────────────────────────────

func RunWelcome() error {
	clearScreen()
	printWelcomeScreen()
	printCommands()
	printQuickStart()
	printFooter()
	return nil
}

// printLogo renders the ASCII logo — shared by welcome screen and chat header.
func printLogo() {
	fmt.Println()
	lines := strings.Split(logo, "\n")
	for i, line := range lines {
		if i%2 == 0 {
			cMint.Println(line)
		} else {
			cMintDim.Println(line)
		}
	}
}

func printWelcomeScreen() {
	printLogo()
	fmt.Println()
	printHRule()

	// Tagline row
	cGray.Printf("  ◆ ")
	cWhite.Printf("Ultra-fast LLM inference")
	cGray.Printf("  ·  ")
	cGrayMid.Printf("Powered by Groq AI")
	cGray.Printf("  ·  ")
	cMint.Printf("%s", version)
	fmt.Println()

	// Developer credit
	cGray.Printf("  ◆ built by ")
	cWhite.Printf("%s ", devName)
	cMint.Printf("(%s)", devHandle)
	fmt.Println()

	printHRule()
}

func printCommands() {
	fmt.Println()
	printSectionLabel("COMMANDS")
	fmt.Println()

	type entry struct{ cmd, desc string }
	entries := []entry{
		{"groq chat", "start interactive AI chat session"},
		{"groq chat \"question\"", "send a single message and exit"},
		{"groq create \"project\"", "generate a full project from description"},
		{"groq run \"task\"", "execute a task with AI-generated code"},
		{"groq models", "list all available Groq models"},
		{"groq config", "manage configuration & API key"},
		{"groq uninstall", "remove Groq CLI from the system"},
	}

	for _, e := range entries {
		cGray.Printf("  ▸ ")
		cMint.Printf("%-38s", e.cmd)
		cGray.Printf("%s\n", e.desc)
	}
}

func printQuickStart() {
	fmt.Println()
	printSectionLabel("QUICK START")
	fmt.Println()

	steps := []struct{ n, label, code string }{
		{"01", "get a free API key", "https://console.groq.com"},
		{"02", "save your key", "groq config set-key YOUR_GROQ_API_KEY"},
		{"03", "start chatting", "groq chat"},
		{"04", "generate a project", `groq create "REST API in Go with PostgreSQL"`},
	}

	for _, s := range steps {
		cGray.Printf("  ")
		cGray.Printf("[")
		cMint.Printf("%s", s.n)
		cGray.Printf("] ")
		cGrayMid.Printf("%-26s", s.label)
		cCode.Printf("%s\n", s.code)
	}
}

func printFooter() {
	fmt.Println()
	printHRule()
	cGray.Printf("  docs   ")
	cMintDim.Printf("https://console.groq.com/docs/overview\n")
	cGray.Printf("  github ")
	cMintDim.Printf("https://github.com/%s/groq-cli\n", strings.ToLower(devHandle))
	printHRule()
	fmt.Println()
}

// ── Section / Rule helpers ──────────────────────────────────────────────────

func printHRule() {
	cGray.Printf("  %s\n", strings.Repeat("─", terminalWidth()-4))
}

func printSectionLabel(label string) {
	cGray.Printf("  ◆ ")
	cMint.Printf("%s\n", label)
}

// ── Public helpers used across the codebase ─────────────────────────────────

func PrintDivider() {
	printHRule()
}

func PrintSection(title string) {
	fmt.Println()
	cGray.Printf("  ◆ ")
	cMint.Printf("%s\n", title)
	cGray.Printf("  %s\n", strings.Repeat("─", len([]rune(title))+4))
}

func PrintSuccess(msg string) {
	cSuccess.Printf("  ◆ %s\n", msg)
}

func PrintError(msg string) {
	cError.Printf("  ✗ %s\n", msg)
}

func PrintWarning(msg string) {
	cWarn.Printf("  ▲ %s\n", msg)
}

func PrintInfo(msg string) {
	cGray.Printf("  · %s\n", msg)
}

func PrintCode(msg string) {
	cCode.Printf("    %s\n", msg)
}

// ── Chat UI ──────────────────────────────────────────────────────────────────

func PrintUserMessage(msg string) {
	fmt.Println()
	cGray.Printf("  ┌─ ")
	cUser.Printf("you")
	cGray.Printf(" %s\n", strings.Repeat("─", terminalWidth()-10))
	cGrayMid.Printf("  │  %s\n", msg)
	cGray.Printf("  └%s\n", strings.Repeat("─", terminalWidth()-4))
}

func PrintAssistantHeader(model string) {
	fmt.Println()
	cGray.Printf("  ┌─ ")
	cMint.Printf("◆ groq")
	cGray.Printf("  [%s]  %s\n", shortenModel(model), strings.Repeat("─", max(0, terminalWidth()-len(shortenModel(model))-16)))
	cGray.Printf("  │\n")
	cGray.Printf("  │  ")
}

func PrintAssistantChunk(chunk string) {
	cWhite.Printf("%s", chunk)
}

func PrintAssistantDone() {
	fmt.Println()
	cGray.Printf("  └%s\n", strings.Repeat("─", terminalWidth()-4))
}

func PrintThinking() {
	frames := []string{"◐", "◓", "◑", "◒"}
	for i := 0; i < 8; i++ {
		cMint.Printf("\r  %s working...", frames[i%len(frames)])
		time.Sleep(120 * time.Millisecond)
	}
	fmt.Printf("\r                    \r")
}

// ── Prompts ───────────────────────────────────────────────────────────────────

func PromptInput(prompt string) (string, error) {
	fmt.Println()
	cGray.Printf("  ┌─ ")
	cMint.Printf("input\n")
	cGray.Printf("  └─ ")
	cWhite.Printf("%s", prompt)
	var input string
	_, err := fmt.Scanln(&input)
	return input, err
}

func PromptConfirm(prompt string) (bool, error) {
	fmt.Println()
	cGray.Printf("  ◆ ")
	cWarn.Printf("%s ", prompt)
	cGray.Printf("[y/N] ")
	var input string
	fmt.Scanln(&input)
	return strings.ToLower(strings.TrimSpace(input)) == "y", nil
}

// ── Internal ──────────────────────────────────────────────────────────────────

func clearScreen() {
	if isTTY {
		fmt.Print("\033[2J\033[H")
	}
}

func terminalWidth() int {
	return 78
}

func shortenModel(model string) string {
	if len(model) > 28 {
		return model[:25] + "..."
	}
	return model
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
