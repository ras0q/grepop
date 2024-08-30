package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
)

var (
	ansiStyleRegexp = regexp.MustCompile(`\x1b[[\d;]*m`)

	isDebug       = flag.Bool("debug", false, "Debug mode")
	heightPercent = flag.Uint("height", 90, "Percentage of terminal height")
	color         = flag.Uint("color", 212, "Foreground color")
	borderColor   = flag.Uint("border-color", 63, "Border foreground color")
	sleep         = flag.Uint("sleep", 500, "Milliseconds to wait for output")
)

func main() {
	flag.Parse()

	if *isDebug {
		log.SetLevel(log.DebugLevel)
	}

	boxStyle := lipgloss.NewStyle().
		Bold(true).
		PaddingLeft(1).
		PaddingRight(1).
		Foreground(lipgloss.ANSIColor(*color)).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.ANSIColor(*borderColor))

	args := flag.Args()
	log.Debug("parse flag", "args", args)
	if len(args) != 1 {
		log.Fatal("Usage: grepop PATTERN") // TODO: support multiple patterns
	}

	pattern := args[0]
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatal("compile pattern", "err", err)
	}

	if term.IsTerminal(os.Stdin.Fd()) {
		log.Fatal("Piped stdin is only supported now.")
	}

	r := os.Stdin
	b, err := io.ReadAll(r)
	if err != nil {
		log.Fatal("read from input", "err", err)
	}

	w, h, err := term.GetSize(os.Stderr.Fd())
	if err != nil {
		log.Fatal("get terminal size", "err", err)
	}

	h = (h * int(*heightPercent)) / 100
	log.Debug("term size", "w", w, "h", h)

	wrapped := ansi.Hardwrap(string(b), w, true)

	matchLocs := re.FindAllStringIndex(wrapped, -1)
	log.Debug("find", "matches", len(matchLocs))

	originalBgLines := strings.Split(wrapped, "\n")
	bgLines := make([]string, len(originalBgLines))
	copy(bgLines, originalBgLines)

	for i, loc := range matchLocs {
		boxLines := strings.Split(
			boxStyle.Render(fmt.Sprintf("%s", wrapped[loc[0]:loc[1]])),
			"\n",
		)

		var row, col, sum int
		for _row, line := range originalBgLines {
			if sum+len(line+"\n") > loc[0] {
				row = _row - 1
				col = loc[0] - sum - 2
				break
			}

			sum += len(line + "\n")
		}

		for j := range len(boxLines) {
			if l := row + j + len(boxLines) - len(bgLines); l > 0 {
				bgLines = append(
					bgLines,
					slices.Repeat([]string{""}, l)...,
				)
			}

			boxLine := boxLines[j]
			bgLine := bgLines[row+j]
			if sw := ansi.StringWidth(bgLine); sw < col {
				bgLine += strings.Repeat(" ", col-sw)
			}

			newBgLine := ansi.EraseEntireLine +
				ansi.Truncate(bgLine, col, "") +
				boxLine

			// NOTE: bgLine has no newline, so [strings.Join] after [strings.Split] is safe.
			wrapped := strings.Split(
				ansi.Hardwrap(bgLine, col+ansi.StringWidth(boxLine), true),
				"\n",
			)
			if len(wrapped) >= 2 {
				var ansiStyle string
				ansiStyles := ansiStyleRegexp.FindAllString(wrapped[0], -1)
				if l := len(ansiStyles); l > 0 {
					ansiStyle = ansiStyles[l-1]
				}
				newBgLine += ansiStyle + strings.Join(wrapped[1:], "")
			}

			bgLines[row+j] = newBgLine
		}

		start, end := 0, len(bgLines)
		if end-start > h {
			padding := h / 2
			start = max(0, row-padding)
			end = min(len(bgLines), row-padding+h)
		}
		fmt.Println(strings.Join(bgLines[start:end], "\n"))
		time.Sleep(time.Millisecond * time.Duration(*sleep))

		if i != len(matchLocs)-1 && !*isDebug {
			fmt.Print(ansi.CursorPreviousLine(end - start))
		}
	}
}
