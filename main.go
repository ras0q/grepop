package main

import (
	"errors"
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
	noBorder      = flag.Bool("no-border", false, "Disable popup border")
	heightPercent = flag.Uint("height", 90, "Percentage of terminal height")
	color         = flag.Uint("color", 212, "Foreground color")
	borderColor   = flag.Uint("border-color", 63, "Border foreground color")
	sleep         = flag.Uint("sleep", 500, "Milliseconds to wait for output")
)

func main() {
	flag.Usage = func() {
		fmt.Println(
			"Usage: grepop [option]... PATTERN\n" +
				"\n" +
				"Examples:\n" +
				"  cat access.log | grepop ERROR\n" +
				"\n" +
				"Options:",
		)
		flag.PrintDefaults()
	}

	flag.Parse()

	if *isDebug {
		log.SetLevel(log.DebugLevel)
	}

	var popupVPad, popupHPad int
	popupStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.ANSIColor(*color))

	if !*noBorder {
		popupVPad = 1
		popupHPad = 2
		popupStyle = popupStyle.
			PaddingLeft(popupHPad - 1).
			PaddingRight(popupHPad - 1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.ANSIColor(*borderColor))
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
		log.Debug("display popup", "i", i, "loc", loc, "match", wrapped[loc[0]:loc[1]])

		popupLines := strings.Split(
			popupStyle.Render(fmt.Sprintf("%s", wrapped[loc[0]:loc[1]])),
			"\n",
		)

		var row, col, sum int
		for _row, line := range originalBgLines {
			if sum+len(line+"\n") > loc[0] {
				row = _row - popupVPad
				col = loc[0] - sum - popupHPad
				break
			}

			sum += len(line + "\n")
		}

		log.Debug("position", "row", row, "col", col)
		if *isDebug {
			actual := originalBgLines[row+popupVPad][col+popupHPad : col+popupHPad+loc[1]-loc[0]]
			expected := wrapped[loc[0]:loc[1]]
			if expected != actual {
				log.Error("incorrect position", "expected", expected, "actual", actual)
			}
		}

		for j := range len(popupLines) {
			if row+j < 0 {
				continue
			}

			if l := row + j + len(popupLines) - len(bgLines); l > 0 {
				bgLines = append(
					bgLines,
					slices.Repeat([]string{""}, l)...,
				)
			}

			popupLine := popupLines[j]
			bgLine := bgLines[row+j]
			if sw := ansi.StringWidth(bgLine); sw < col {
				bgLine += strings.Repeat(" ", col-sw)
			}

			bgLeft := ansi.Truncate(bgLine, col, "")

			if col < 0 {
				_popupLine, err := cutLeft(popupLine, -col)
				if err != nil {
					log.Error("cut left of popupLine", "err", err)
				}

				log.Debug("popupLine", "popupLine", _popupLine)

				popupLine = _popupLine
			}

			bgRight, err := cutLeft(bgLine, max(0, col)+ansi.StringWidth(popupLine))
			if err != nil {
				log.Error("cut left of bgLine", "err", err)
			}

			bgLines[row+j] = ansi.Truncate(bgLeft+popupLine+bgRight, w, "")
		}

		start, end := 0, len(bgLines)
		if end-start > h {
			padding := h / 2
			start = max(0, row-padding)
			end = min(len(bgLines), row-padding+h)
		}

		var prefix string
		if !*isDebug {
			prefix = ansi.EraseEntireLine
		}
		fmt.Println(prefix + strings.Join(bgLines[start:end], "\n"+prefix))
		time.Sleep(time.Millisecond * time.Duration(*sleep))

		if i != len(matchLocs)-1 && !*isDebug {
			fmt.Print(ansi.CursorPreviousLine(end - start))
		}
	}
}

func cutLeft(line string, padding int) (string, error) {
	if strings.Contains(line, "\n") {
		return "", errors.New("line must not contain newline")
	}

	// NOTE: line has no newline, so [strings.Join] after [strings.Split] is safe.
	wrapped := strings.Split(ansi.Hardwrap(line, padding, true), "\n")
	if len(wrapped) == 1 {
		return wrapped[0], nil
	}

	var ansiStyle string
	ansiStyles := ansiStyleRegexp.FindAllString(wrapped[0], -1)
	if l := len(ansiStyles); l > 0 {
		ansiStyle = ansiStyles[l-1]
	}

	return ansiStyle + strings.Join(wrapped[1:], ""), nil
}
