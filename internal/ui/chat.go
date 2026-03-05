package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"groq-cli/internal/clipboard"
	"groq-cli/internal/groq"
)

// ── Interactive chat ──────────────────────────────────────────────────────────

func RunInteractiveChat(client *groq.Client, model string) error {
	clearScreen()
	printChatHeader(model)

	var messages []groq.Message
	messages = append(messages, groq.Message{
		Role:    "system",
		Content: "You are a helpful, knowledgeable assistant running inside a terminal CLI powered by Groq. Be concise, accurate, and helpful. Format code with proper markdown when showing code snippets.",
	})

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println()
		cGray.Printf("  → ")

		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// ── Built-in commands ─────────────────────────────────────────────
		switch strings.ToLower(input) {
		case "/exit", "/quit", "/q":
			fmt.Println()
			printHRule()
			cGray.Printf("  ◆ ")
			cMintDim.Printf("session ended  ·  goodbye 👋\n")
			printHRule()
			fmt.Println()
			return nil
		case "/clear":
			clearScreen()
			printChatHeader(model)
			messages = messages[:1]
			continue
		case "/help":
			printChatHelp()
			continue
		}

		if strings.HasPrefix(input, "/model ") {
			model = strings.TrimSpace(strings.TrimPrefix(input, "/model "))
			cSuccess.Printf("  ◆ model → %s\n", model)
			continue
		}

		messages = append(messages, groq.Message{Role: "user", Content: input})

		// User bubble
		fmt.Println()
		printWrappedBox(boxOpts{
			label:      "you",
			labelMint:  false,
			content:    input,
			contentDim: true,
		})

		// Assistant bubble — always ModeDefault in interactive mode
		fmt.Println()
		plain, err := streamResponse(client, groq.ChatRequest{
			Model:       model,
			Messages:    messages,
			MaxTokens:   8192,
			Temperature: 0.7,
			Stream:      true,
		}, ModeDefault)

		if err != nil {
			PrintError(fmt.Sprintf("request failed: %v", err))
			messages = messages[:len(messages)-1]
			continue
		}

		messages = append(messages, groq.Message{Role: "assistant", Content: plain})
		cGray.Printf("  · %d messages in context\n", len(messages)-1)
	}

	return nil
}

// ── Single-shot chat ──────────────────────────────────────────────────────────

// ChatOptions carries flags passed from the cobra command.
type ChatOptions struct {
	Model       string
	MaxTokens   int
	Temperature float64
	Mode        OutputMode
}

func RunSingleChat(client *groq.Client, opts ChatOptions, message string) error {
	req := groq.ChatRequest{
		Model: opts.Model,
		Messages: []groq.Message{
			{Role: "system", Content: "You are a helpful assistant. Be concise and accurate."},
			{Role: "user", Content: message},
		},
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
		Stream:      true,
	}

	switch opts.Mode {
	case ModePlain:
		// No borders — print raw text directly to stdout
		return streamPlainToStdout(client, req)

	case ModeCopy:
		// Collect full response, copy to clipboard, print confirmation
		return streamAndCopy(client, req)

	default:
		// Default: bordered UI
		fmt.Println()
		printWrappedBox(boxOpts{
			label:      "you",
			labelMint:  false,
			content:    message,
			contentDim: true,
		})
		fmt.Println()
		_, err := streamResponse(client, req, ModeDefault)
		fmt.Println()
		return err
	}
}

// ── Stream helpers ────────────────────────────────────────────────────────────

// streamResponse streams a request and renders it according to mode.
// Returns the plain text of the full response.
func streamResponse(client *groq.Client, req groq.ChatRequest, mode OutputMode) (string, error) {
	switch mode {
	case ModePlain:
		err := streamPlainToStdout(client, req)
		return "", err // plain mode doesn't track the text
	default:
		// Bordered box
		printAssistantBoxTop(req.Model)
		w := &wrapWriter{}
		err := client.ChatStream(req, func(chunk string) {
			w.Write(chunk)
		})
		w.Flush()
		printAssistantBoxBottom()
		return w.Plain.String(), err
	}
}

// streamPlainToStdout streams response as raw text — no borders, no colors.
// Safe to pipe: groq chat --plain "q" | wc -w
func streamPlainToStdout(client *groq.Client, req groq.ChatRequest) error {
	return client.ChatStream(req, func(chunk string) {
		fmt.Print(chunk)
	})
}

// streamAndCopy collects the full response, copies it to clipboard,
// and prints a brief bordered confirmation to the terminal.
func streamAndCopy(client *groq.Client, req groq.ChatRequest) error {
	// Show a spinner while collecting
	done := make(chan bool)
	go func() {
		frames := []string{"◐", "◓", "◑", "◒"}
		i := 0
		for {
			select {
			case <-done:
				return
			default:
				cMint.Printf("\r  %s collecting response...", frames[i%len(frames)])
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	var full strings.Builder
	err := client.ChatStream(req, func(chunk string) {
		full.WriteString(chunk)
	})
	done <- true
	fmt.Printf("\r                                      \r")

	if err != nil {
		return err
	}

	text := full.String()

	// Print the response in the bordered box so the user can read it
	fmt.Println()
	printAssistantBoxTop(req.Model)
	w := &wrapWriter{}
	w.Write(text)
	w.Flush()
	printAssistantBoxBottom()
	fmt.Println()

	// Copy clean text to clipboard
	if err := clipboard.Copy(text); err != nil {
		// Warn on stderr so stdout stays clean
		fmt.Fprintf(os.Stderr, "\n")
		PrintWarning(fmt.Sprintf("clipboard error: %v", err))
		return nil
	}

	// Confirmation badge
	printHRule()
	cGray.Printf("  ◆ ")
	cSuccess.Printf("copied to clipboard")
	cGray.Printf("  ·  %d chars  ·  %d words\n",
		len([]rune(text)),
		len(strings.Fields(text)),
	)
	printHRule()
	fmt.Println()

	return nil
}

// ── Box helpers ───────────────────────────────────────────────────────────────

func printAssistantBoxTop(model string) {
	label := "◆ groq  [" + shortenModel(model) + "]"
	printBoxTop(label, true)
	// blank gap after header
	cGray.Printf("%s%s%s\n", boxLeft, strings.Repeat(" ", boxContentWidth), boxRight)
}

func printAssistantBoxBottom() {
	fmt.Println()
	// closing blank line with right border
	cGray.Printf("%s%s%s\n", boxLeft, strings.Repeat(" ", boxContentWidth), boxRight)
	printBoxBottom()
}

func printChatHeader(model string) {
	printLogo()
	fmt.Println()
	printHRule()
	cGray.Printf("  ◆ ")
	cMint.Printf("CHAT SESSION\n")
	cGray.Printf("  · model   ")
	cWhite.Printf("%s\n", model)
	cGray.Printf("  · type ")
	cCode.Printf("/help")
	cGray.Printf(" for commands  ·  ")
	cCode.Printf("/exit")
	cGray.Printf(" to quit\n")
	printHRule()
}

func printChatHelp() {
	fmt.Println()
	printSectionLabel("CHAT COMMANDS")
	fmt.Println()
	rows := [][]string{
		{"/exit  /quit  /q", "end session"},
		{"/clear", "clear history and screen"},
		{"/model <id>", "switch model mid-session"},
		{"/help", "show this message"},
	}
	for _, r := range rows {
		cGray.Printf("  ▸ ")
		cMint.Printf("%-26s", r[0])
		cGray.Printf("%s\n", r[1])
	}
	fmt.Println()
}

func renderChunk(chunk string) string { return chunk }

// sleepMs sleeps for n milliseconds — avoids importing time in this file

