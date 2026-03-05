package ui

import (
	"strings"
	"unicode/utf8"
)

// ── Layout ────────────────────────────────────────────────────────────────────
//
// Terminal width = 78 columns. Chat box anatomy:
//
//   col: 1234567890...                                              ...7778
//        "  ┌─ label ───────────────────────────────────────────────────┐"
//        "  │  content (up to 70 runes)                                 │"
//        "  │  content (up to 70 runes) <padded with spaces to col 75>  │"
//        "  └──────────────────────────────────────────────────────────┘"
//
// Breakdown per row:
//   left  = "  │  "     → 5 chars  (2 indent + │ + 2 spaces)
//   body  = ≤ 70 runes  (content padded to exactly 70 with spaces)
//   right = "  │"       → 3 chars  (2 spaces + │)
//   total = 5 + 70 + 3  = 78 ✓
//
// Top border:  "  ┌─ " (5) + label + " " + "─"*(71-len(label)) + "┐" (1) = 78
// Bot border:  "  └"   (3) + "─"*74 + "┘" (1)                             = 78

const (
	boxLeft         = "  │  "    // left wall + inner padding
	boxRight        = "  │"      // right wall (2 spaces + │)
	boxContentWidth = 70         // max runes of text per line
)

// ── wrapWriter ────────────────────────────────────────────────────────────────
// Buffers streaming AI chunks and renders them word-wrapped inside the box,
// with a right-side border │ on every line.
//
// .Plain accumulates raw text with NO box characters — safe to copy/paste.

type wrapWriter struct {
	buf   strings.Builder // partial current line
	Plain strings.Builder // clean text, zero box chars
	col   int             // rune count in buf
}

func (w *wrapWriter) Write(chunk string) {
	w.Plain.WriteString(chunk)

	for _, ch := range chunk {
		switch ch {
		case '\n':
			w.flushLine()
			w.printEmptyLine()
		case '\r':
			// discard
		default:
			w.buf.WriteRune(ch)
			w.col++

			if w.col >= boxContentWidth {
				// Work in rune space — strings.LastIndex returns a byte offset
				// which is wrong for multibyte chars (accents, em-dash, etc.)
				// and causes "slice bounds out of range" panics.
				runes := []rune(w.buf.String())

				// Scan runes right-to-left for last space
				breakAt := -1
				for i := len(runes) - 1; i >= 0; i-- {
					if runes[i] == ' ' {
						breakAt = i
						break
					}
				}

				if breakAt <= 0 {
					// Hard break — no space found
					w.flushLine()
				} else {
					// Soft break at last space (rune index — always safe)
					w.flushLineStr(string(runes[:breakAt]))
					rest := strings.TrimLeft(string(runes[breakAt:]), " ")
					w.buf.Reset()
					w.buf.WriteString(rest)
					w.col = utf8.RuneCountInString(rest)
				}
			}
		}
	}
}

func (w *wrapWriter) Flush() {
	if w.col > 0 {
		w.flushLine()
	}
}

func (w *wrapWriter) flushLine() {
	w.flushLineStr(w.buf.String())
	w.buf.Reset()
	w.col = 0
}

// flushLineStr prints one content row padded to boxContentWidth + right border.
func (w *wrapWriter) flushLineStr(line string) {
	runeLen := utf8.RuneCountInString(line)
	pad := boxContentWidth - runeLen
	if pad < 0 {
		pad = 0
	}
	cGray.Printf("%s", boxLeft)
	cWhite.Printf("%s", line)
	cGray.Printf("%s%s\n", strings.Repeat(" ", pad), boxRight)
}

// printEmptyLine prints a blank interior row (preserves the right border).
func (w *wrapWriter) printEmptyLine() {
	cGray.Printf("%s%s%s\n", boxLeft, strings.Repeat(" ", boxContentWidth), boxRight)
}

// ── wrapLines ─────────────────────────────────────────────────────────────────
// Splits text into word-wrapped lines of at most boxContentWidth runes each.
// Lines do NOT include any box characters — callers apply borders.

func wrapLines(text string) []string {
	var result []string

	for _, para := range strings.Split(text, "\n") {
		if strings.TrimSpace(para) == "" {
			result = append(result, "")
			continue
		}
		words := strings.Fields(para)
		var line strings.Builder
		lineRunes := 0

		for _, word := range words {
			wRunes := utf8.RuneCountInString(word)
			if lineRunes == 0 {
				line.WriteString(word)
				lineRunes = wRunes
			} else if lineRunes+1+wRunes <= boxContentWidth {
				line.WriteByte(' ')
				line.WriteString(word)
				lineRunes += 1 + wRunes
			} else {
				result = append(result, line.String())
				line.Reset()
				line.WriteString(word)
				lineRunes = wRunes
			}
		}
		if lineRunes > 0 {
			result = append(result, line.String())
		}
	}
	return result
}

// ── printWrappedBox ───────────────────────────────────────────────────────────
// Renders a full bordered box with right-side │ for pre-known content.
// Only this function and wrapWriter emit box characters.

type boxOpts struct {
	label      string
	labelMint  bool // true = mint, false = bold-white (user)
	content    string
	contentDim bool // true = cGrayMid, false = cWhite
}

func printWrappedBox(opts boxOpts) {
	printBoxTop(opts.label, opts.labelMint)

	for _, l := range wrapLines(opts.content) {
		printBoxContentLine(l, opts.contentDim)
	}

	printBoxBottom()
}

// printBoxTop renders the ┌─ label ──...──┐ line.
func printBoxTop(label string, mint bool) {
	// "  ┌─ " = 5, label, " " = 1, dashes, "┐" = 1  → total = 78
	// dashes = 78 - 5 - len(label) - 1 - 1 = 71 - len(label)
	labelRunes := utf8.RuneCountInString(label)
	dashes := 71 - labelRunes
	if dashes < 1 {
		dashes = 1
	}
	cGray.Printf("  ┌─ ")
	if mint {
		cMint.Printf("%s", label)
	} else {
		cUser.Printf("%s", label)
	}
	cGray.Printf(" %s┐\n", strings.Repeat("─", dashes))
}

// printBoxContentLine renders one "  │  content<pad>  │" row.
func printBoxContentLine(line string, dim bool) {
	runeLen := utf8.RuneCountInString(line)
	pad := boxContentWidth - runeLen
	if pad < 0 {
		pad = 0
	}
	if line == "" {
		// empty interior row
		cGray.Printf("%s%s%s\n", boxLeft, strings.Repeat(" ", boxContentWidth), boxRight)
		return
	}
	cGray.Printf("%s", boxLeft)
	if dim {
		cGrayMid.Printf("%s", line)
	} else {
		cWhite.Printf("%s", line)
	}
	cGray.Printf("%s%s\n", strings.Repeat(" ", pad), boxRight)
}

// printBoxBottom renders the └──...──┘ line.
func printBoxBottom() {
	// "  └" (3) + "─"*74 + "┘" (1) = 78
	cGray.Printf("  └%s┘\n", strings.Repeat("─", 74))
}
