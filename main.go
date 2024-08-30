package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
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
	boxStyle = lipgloss.NewStyle().
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.ANSIColor(63))
	ansiStyleRegexp = regexp.MustCompile(`\x1b[[\d;]*m`)

	isDebug = flag.Bool("debug", false, "Debug Mode")
)

func main() {
	flag.Parse()

	if isDebug != nil && *isDebug {
		log.SetLevel(log.DebugLevel)
	}

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

	matches := re.FindAll(b, -1)
	log.Debug("find", "matches", len(matches))

	w, _, err := term.GetSize(os.Stderr.Fd())
	if err != nil {
		log.Fatal("get terminal size", "err", err)
	}

	bg := string(b)
	var maxRow int
	for i, match := range matches {
		boxLines := strings.Split(
			boxStyle.Render(fmt.Sprintf("%d: %s", i, match)),
			"\n",
		)
		bgLines := strings.Split(ansi.Hardwrap(bg, w, true), "\n")
		if i == 0 {
			maxRow = len(bgLines)
		}

		row := rand.IntN(maxRow)
		col := rand.IntN(w - ansi.StringWidth(boxLines[0]))
		log.Debug("point", "i", i, "row", row, "col", col)

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

			newBgLine := ansi.Truncate(bgLine, col, "") + boxLine

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

		bg = strings.Join(bgLines, "\n")

		fmt.Printf(
			"%s%s%s",
			bg,
			ansi.CursorPreviousLine(len(bgLines)-1),
			ansi.CursorLeft(len(bgLines[0])),
		)
		time.Sleep(time.Millisecond * 500)
	}

	fmt.Println(bg)
}
