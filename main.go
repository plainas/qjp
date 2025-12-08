// Copyright (c) 2025 Pedro (http://github.com/plainas)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/term"
)

const (
	// ANSI escape codes
	clearScreen  = "\033[2J"
	cursorHome   = "\033[H"
	hideCursor   = "\033[?25l"
	showCursor   = "\033[?25h"
	clearLine    = "\033[2K"
	colorReset   = "\033[0m"
	colorReverse = "\033[7m"
	colorCyan    = "\033[36m"
	colorGreen   = "\033[32m"
	altScreenOn  = "\033[?1049h"
	altScreenOff = "\033[?1049l"
)

func setRawMode(fd uintptr) (*term.State, error) {
	oldState, err := term.MakeRaw(int(fd))
	if err != nil {
		return nil, err
	}
	return oldState, nil
}

func restoreTerminal(fd uintptr, oldState *term.State) {
	if oldState == nil {
		return
	}
	_ = term.Restore(int(fd), oldState)
}

func getTerminalSize(tty *os.File) (width, height int, err error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = tty
	out, err := cmd.Output()
	if err != nil {
		return 80, 24, nil // default values
	}
	fmt.Sscanf(string(out), "%d %d", &height, &width)
	return width, height, nil
}

type App struct {
	objects     []map[string]interface{}
	displayAttr string
	outputAttr  string
	cursor      int
	filtered    []int
	filter      string
	width       int
	height      int
	tty         *os.File
	truncate    bool
}

func newApp(objects []map[string]interface{}, displayAttr, outputAttr string, tty *os.File, truncate bool) *App {
	width, height, _ := getTerminalSize(tty)
	filtered := make([]int, len(objects))
	for i := range objects {
		filtered[i] = i
	}

	return &App{
		objects:     objects,
		displayAttr: displayAttr,
		outputAttr:  outputAttr,
		cursor:      0,
		filtered:    filtered,
		filter:      "",
		width:       width,
		height:      height,
		tty:         tty,
		truncate:    truncate,
	}
}

func (a *App) updateFilter() {
	filterText := strings.ToLower(a.filter)
	if filterText == "" {
		a.filtered = make([]int, len(a.objects))
		for i := range a.objects {
			a.filtered[i] = i
		}
		return
	}

	a.filtered = []int{}
	for i, obj := range a.objects {
		displayVal := a.getDisplayValue(obj)
		if strings.Contains(strings.ToLower(displayVal), filterText) {
			a.filtered = append(a.filtered, i)
		}
	}

	// Adjust cursor if needed
	if a.cursor >= len(a.filtered) {
		a.cursor = max(0, len(a.filtered)-1)
	}
}

func (a *App) getDisplayValue(obj map[string]interface{}) string {
	if a.displayAttr == "" {
		// Display entire object as JSON on one line
		jsonBytes, err := json.Marshal(obj)
		if err == nil {
			return string(jsonBytes)
		}
		return ""
	}
	if val, ok := obj[a.displayAttr]; ok {
		return fmt.Sprintf("%v", val)
	}
	return ""
}

func (a *App) calculateLines(displayVal string) int {
	if displayVal == "" {
		return 1
	}
	if a.truncate {
		return 1
	}
	// Account for "> " or "  " prefix (2 chars)
	effectiveWidth := a.width - 2
	if effectiveWidth <= 0 {
		return 1
	}
	lines := (len(displayVal) + effectiveWidth - 1) / effectiveWidth
	if lines == 0 {
		return 1
	}
	return lines
}

func (a *App) render() {
	fmt.Fprint(a.tty, clearScreen+cursorHome)

	// Display filter
	fmt.Fprintf(a.tty, "%sFilter:%s %s\r\n", colorCyan, colorReset, a.filter)

	// Calculate visible window based on actual line usage
	availableLines := a.height - 4
	if availableLines <= 0 {
		availableLines = 1
	}

	// Find the range of items to display
	start := 0
	end := len(a.filtered)

	if len(a.filtered) > 0 {
		// First, try to center cursor in viewport
		usedLines := 0
		start = a.cursor

		// Expand upward from cursor
		for start > 0 {
			idx := a.filtered[start-1]
			obj := a.objects[idx]
			displayVal := a.getDisplayValue(obj)
			itemLines := a.calculateLines(displayVal)
			if usedLines+itemLines > availableLines/2 {
				break
			}
			start--
			usedLines += itemLines
		}

		// Add cursor item
		idx := a.filtered[a.cursor]
		obj := a.objects[idx]
		displayVal := a.getDisplayValue(obj)
		usedLines += a.calculateLines(displayVal)

		// Expand downward from cursor
		end = a.cursor + 1
		for end < len(a.filtered) {
			idx := a.filtered[end]
			obj := a.objects[idx]
			displayVal := a.getDisplayValue(obj)
			itemLines := a.calculateLines(displayVal)
			if usedLines+itemLines > availableLines {
				break
			}
			usedLines += itemLines
			end++
		}
	}

	// Display items
	for i := start; i < end; i++ {
		idx := a.filtered[i]
		obj := a.objects[idx]
		displayVal := a.getDisplayValue(obj)

		// Truncate if needed
		if a.truncate {
			maxWidth := a.width - 2 // Account for "> " or "  " prefix
			if len(displayVal) > maxWidth && maxWidth > 3 {
				displayVal = displayVal[:maxWidth-3] + "..."
			}
		}

		if i == a.cursor {
			fmt.Fprintf(a.tty, "%s> %s%s\r\n", colorReverse, displayVal, colorReset)
		} else {
			fmt.Fprintf(a.tty, "  %s\r\n", displayVal)
		}
	}

	if len(a.filtered) == 0 {
		fmt.Fprint(a.tty, "  (no matches)\r\n")
	}
}

func (a *App) run() (int, error) {
	ttyFd := a.tty.Fd()
	oldState, err := setRawMode(ttyFd)
	if err != nil {
		return -1, err
	}
	defer restoreTerminal(ttyFd, oldState)

	// Use alternate screen buffer
	fmt.Fprint(a.tty, altScreenOn+hideCursor)
	defer fmt.Fprint(a.tty, showCursor+altScreenOff)

	a.render()

	buf := make([]byte, 3)
	for {
		n, err := a.tty.Read(buf)
		if err != nil {
			return -1, err
		}

		if n == 1 {
			switch buf[0] {
			case 3, 27: // Ctrl+C or ESC
				return -1, nil
			case 10, 13: // Enter (newline or carriage return)
				if len(a.filtered) > 0 && a.cursor < len(a.filtered) {
					return a.filtered[a.cursor], nil
				}
			case 127: // Backspace
				if len(a.filter) > 0 {
					a.filter = a.filter[:len(a.filter)-1]
					a.updateFilter()
					a.render()
				}
			default:
				if buf[0] >= 32 && buf[0] < 127 {
					a.filter += string(buf[0])
					a.updateFilter()
					a.render()
				}
			}
		} else if n == 3 && buf[0] == 27 && buf[1] == 91 {
			// Arrow keys
			switch buf[2] {
			case 65: // Up
				if a.cursor > 0 {
					a.cursor--
					a.render()
				}
			case 66: // Down
				if a.cursor < len(a.filtered)-1 {
					a.cursor++
					a.render()
				}
			}
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func output_usage_message_to_stderr() {
	fmt.Fprintln(os.Stderr, "Usage: qjp [display-attribute] [-o output-attribute] [-t] < input.json")
	fmt.Fprintln(os.Stderr, "       qjp [-o output-attribute] [-t] [display-attribute] < input.json")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "If no display-attribute is provided, the whole object is displayed.")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Options:")
	fmt.Fprintln(os.Stderr, "  -o <attr>  Output specific attribute from selected object")
	fmt.Fprintln(os.Stderr, "  -t         Truncate long lines instead of wrapping")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Example:")
	fmt.Fprintln(os.Stderr, "  cat cars.json | qjp")
	fmt.Fprintln(os.Stderr, "  cat cars.json | qjp model")
	fmt.Fprintln(os.Stderr, "  cat cars.json | qjp model -o id")
	fmt.Fprintln(os.Stderr, "  cat cars.json | qjp -t")
}

func main() {
	// Manual parsing to support flags after positional arguments
	var outputAttr string
	var displayAttr string
	var truncate bool

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		if args[i] == "-o" && i+1 < len(args) {
			outputAttr = args[i+1]
			i++ // skip the next arg
		} else if args[i] == "-t" {
			truncate = true
		} else if args[i] == "-h" || args[i] == "--help" {
			output_usage_message_to_stderr()
			os.Exit(0)
		} else if !strings.HasPrefix(args[i], "-") {
			displayAttr = args[i]
		}
	}

	// displayAttr is optional - if not provided, display whole object

	// Read JSON from stdin
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	var objects []map[string]interface{}
	if err := json.Unmarshal(input, &objects); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	if len(objects) == 0 {
		fmt.Fprintln(os.Stderr, "No objects found in input")
		os.Exit(1)
	}

	// Open /dev/tty for interactive input/output
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening /dev/tty: %v\n", err)
		os.Exit(1)
	}
	defer tty.Close()

	app := newApp(objects, displayAttr, outputAttr, tty, truncate)
	selectedIdx, err := app.run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if selectedIdx >= 0 {
		selectedObj := app.objects[selectedIdx]

		if outputAttr != "" {
			// Output specific attribute
			if val, ok := selectedObj[outputAttr]; ok {
				// Format output based on type
				switch v := val.(type) {
				case float64:
					// Check if it's actually an integer
					if v == float64(int64(v)) {
						fmt.Println(int64(v))
					} else {
						fmt.Println(v)
					}
				case string:
					fmt.Println(v)
				default:
					fmt.Println(v)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Attribute '%s' not found in selected object\n", outputAttr)
				os.Exit(1)
			}
		} else {
			// Output entire object on one line
			jsonBytes, err := json.Marshal(selectedObj)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonBytes))
		}
	}
}
