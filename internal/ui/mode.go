package ui

// OutputMode controls how a response is rendered and delivered.
type OutputMode int

const (
	// ModeDefault — bordered, colored terminal UI
	ModeDefault OutputMode = iota

	// ModePlain — raw text only, no borders, no colors.
	// Ideal for piping: groq chat --plain "q" | grep foo
	ModePlain

	// ModeCopy — renders plain text and copies it to the clipboard.
	// Prints a confirmation line to stderr so stdout stays clean.
	ModeCopy
)
