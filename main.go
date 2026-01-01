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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"

	"golang.org/x/term"
)

const (
	// ANSI escape codes
	clearScreen   = "\033[2J"
	cursorHome    = "\033[H"
	hideCursor    = "\033[?25l"
	showCursor    = "\033[?25h"
	clearLine     = "\033[2K"
	colorReset    = "\033[0m"
	colorReverse  = "\033[7m"
	colorCyan     = "\033[36m"
	colorGreen    = "\033[32m"
	colorSelected = "\033[42m" // Green background for selected
	altScreenOn   = "\033[?1049h"
	altScreenOff  = "\033[?1049l"
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
	objects      []map[string]interface{}
	displayAttrs []string
	outputAttr   string
	cursor       int
	filtered     []int
	filter       string
	width        int
	height       int
	tty          *os.File
	truncate     bool
	tableMode    bool
	selected     map[int]bool
	separator    string
	colWidths    []int
}

func newApp(objects []map[string]interface{}, displayAttrs []string, outputAttr string, tty *os.File, truncate bool, tableMode bool, separator string) *App {
	width, height, _ := getTerminalSize(tty)
	filtered := make([]int, len(objects))
	for i := range objects {
		filtered[i] = i
	}

	app := &App{
		objects:      objects,
		displayAttrs: displayAttrs,
		outputAttr:   outputAttr,
		cursor:       0,
		filtered:     filtered,
		filter:       "",
		width:        width,
		height:       height,
		tty:          tty,
		truncate:     truncate,
		tableMode:    tableMode,
		selected:     make(map[int]bool),
		separator:    separator,
	}

	if tableMode && len(displayAttrs) > 0 {
		app.calculateColumnWidths()
	}

	return app
}

func (a *App) calculateColumnWidths() {
	a.colWidths = make([]int, len(a.displayAttrs))

	// Calculate max width for each column
	for _, obj := range a.objects {
		for i, attr := range a.displayAttrs {
			if val, ok := obj[attr]; ok {
				valStr := fmt.Sprintf("%v", val)
				if len(valStr) > a.colWidths[i] {
					a.colWidths[i] = len(valStr)
				}
			}
		}
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
	if len(a.displayAttrs) == 0 {
		// Display entire object as JSON on one line
		jsonBytes, err := json.Marshal(obj)
		if err == nil {
			return string(jsonBytes)
		}
		return ""
	}

	// Get values for each display attribute
	values := []string{}
	for i, attr := range a.displayAttrs {
		var valStr string
		if val, ok := obj[attr]; ok {
			// Check if value is an object or array, serialize to JSON
			switch v := val.(type) {
			case map[string]interface{}, []interface{}:
				jsonBytes, err := json.Marshal(v)
				if err == nil {
					valStr = string(jsonBytes)
				} else {
					valStr = fmt.Sprintf("%v", val)
				}
			default:
				valStr = fmt.Sprintf("%v", val)
			}
		}

		if a.tableMode && i < len(a.colWidths) {
			// Pad value to column width
			if i < len(a.displayAttrs)-1 {
				// Not the last column, pad to width
				valStr = fmt.Sprintf("%-*s", a.colWidths[i], valStr)
			}
			// Last column doesn't need padding
		}

		values = append(values, valStr)
	}

	if len(values) == 0 {
		return ""
	}

	if a.tableMode {
		return strings.Join(values, "  ")
	}
	return strings.Join(values, a.separator)
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

func (a *App) getMaxDisplayWidth() int {
	maxWidth := 0
	for _, idx := range a.filtered {
		obj := a.objects[idx]
		displayVal := a.getDisplayValue(obj)
		if len(displayVal) > maxWidth {
			maxWidth = len(displayVal)
		}
	}
	return maxWidth
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

	// Calculate max display width for uniform background highlighting
	maxDisplayWidth := a.getMaxDisplayWidth()

	// Check if any lines will wrap (only if not truncating)
	hasWrappingLines := false
	if !a.truncate {
		effectiveWidth := a.width - 2 // Account for "> " or "  " prefix
		for i := start; i < end; i++ {
			idx := a.filtered[i]
			obj := a.objects[idx]
			displayVal := a.getDisplayValue(obj)
			if len(displayVal) > effectiveWidth {
				hasWrappingLines = true
				break
			}
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

		// Pad display value for uniform highlighting, but only if:
		// - truncate mode is enabled, OR
		// - no lines are wrapping
		var renderVal string
		if a.truncate || !hasWrappingLines {
			renderVal = fmt.Sprintf("%-*s", maxDisplayWidth, displayVal)
		} else {
			renderVal = displayVal
		}

		isSelected := a.selected[idx]
		if i == a.cursor {
			if isSelected {
				fmt.Fprintf(a.tty, "%s%s> %s%s\r\n", colorReverse, colorSelected, renderVal, colorReset)
			} else {
				fmt.Fprintf(a.tty, "%s> %s%s\r\n", colorReverse, renderVal, colorReset)
			}
		} else {
			if isSelected {
				fmt.Fprintf(a.tty, "%s  %s%s\r\n", colorSelected, renderVal, colorReset)
			} else {
				fmt.Fprintf(a.tty, "  %s\r\n", displayVal)
			}
		}
	}

	if len(a.filtered) == 0 {
		fmt.Fprint(a.tty, "  (no matches)\r\n")
	}
}

func (a *App) toggleSelection() {
	if len(a.filtered) > 0 && a.cursor < len(a.filtered) {
		idx := a.filtered[a.cursor]
		a.selected[idx] = !a.selected[idx]
		if a.cursor < len(a.filtered)-1 {
			a.cursor++
		}
	}
}

func (a *App) getSelection() []int {
	if len(a.filtered) == 0 || a.cursor >= len(a.filtered) {
		return nil
	}

	if len(a.selected) > 0 {
		result := make([]int, 0, len(a.selected))
		for idx := range a.selected {
			result = append(result, idx)
		}
		sort.Ints(result)
		return result
	}

	return []int{a.filtered[a.cursor]}
}

func (a *App) handleBackspace() {
	if len(a.filter) > 0 {
		a.filter = a.filter[:len(a.filter)-1]
		a.updateFilter()
	}
}

func (a *App) handleCharacter(ch byte) {
	if ch >= 32 && ch < 127 {
		a.filter += string(ch)
		a.updateFilter()
	}
}

func (a *App) moveCursorUp() {
	if a.cursor > 0 {
		a.cursor--
	}
}

func (a *App) moveCursorDown() {
	if a.cursor < len(a.filtered)-1 {
		a.cursor++
	}
}

func (a *App) handleInput(buf []byte, n int) (done bool, result []int) {
	if n == 1 {
		switch buf[0] {
		case 0: // Ctrl+Space
			a.toggleSelection()
			a.render()
		case 3, 27: // Ctrl+C or ESC
			return true, nil
		case 10, 13: // Enter
			return true, a.getSelection()
		case 127: // Backspace
			a.handleBackspace()
			a.render()
		default:
			a.handleCharacter(buf[0])
			a.render()
		}
	} else if n == 3 && buf[0] == 27 && buf[1] == 91 {
		switch buf[2] {
		case 65: // Up arrow
			a.moveCursorUp()
			a.render()
		case 66: // Down arrow
			a.moveCursorDown()
			a.render()
		}
	}

	return false, nil
}

func (a *App) run() ([]int, error) {
	ttyFd := a.tty.Fd()
	oldState, err := setRawMode(ttyFd)
	if err != nil {
		return nil, err
	}
	defer restoreTerminal(ttyFd, oldState)

	fmt.Fprint(a.tty, altScreenOn+hideCursor)
	defer fmt.Fprint(a.tty, showCursor+altScreenOff)

	a.render()

	buf := make([]byte, 3)
	for {
		n, err := a.tty.Read(buf)
		if err != nil {
			return nil, err
		}

		if done, result := a.handleInput(buf, n); done {
			return result, nil
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

type config struct {
	outputAttr   string
	displayAttrs []string
	truncate     bool
	tableMode    bool
	lineMode     bool
	allAttrs     bool
	filename     string
	separator    string
}

func outputUsage() {
	fmt.Fprintln(os.Stderr, "Usage: qjp [filename] [-d display-attribute] [-o output-attribute] [-s separator] [-t] [-T] [-l] [-a]")
	fmt.Fprintln(os.Stderr, "       qjp [-d display-attribute] [-o output-attribute] [-s separator] [-t] [-T] [-l] [-a] < input")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Input can be provided via stdin or filename, but not both.")
	fmt.Fprintln(os.Stderr, "If no display-attribute is provided, the whole object is displayed.")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Options:")
	fmt.Fprintln(os.Stderr, "  -d <attr>  Display specific attribute in list (can be used multiple times)")
	fmt.Fprintln(os.Stderr, "  -o <attr>  Output specific attribute from selected object(s)")
	fmt.Fprintln(os.Stderr, "  -s <sep>   Separator for multiple display attributes (default: \" - \")")
	fmt.Fprintln(os.Stderr, "  -t         Truncate long lines instead of wrapping")
	fmt.Fprintln(os.Stderr, "  -T         Table mode: align attributes in columns")
	fmt.Fprintln(os.Stderr, "  -l         Line mode: treat input as plain text lines (like percol)")
	fmt.Fprintln(os.Stderr, "  -a         Display all attributes (cannot be used with -d)")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Controls:")
	fmt.Fprintln(os.Stderr, "  Arrow Keys    Navigate up/down")
	fmt.Fprintln(os.Stderr, "  Ctrl+Space    Toggle selection (multi-select)")
	fmt.Fprintln(os.Stderr, "  Enter         Confirm selection")
	fmt.Fprintln(os.Stderr, "  ESC/Ctrl+C    Cancel")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Examples:")
	fmt.Fprintln(os.Stderr, "  qjp yourfile.json -d display_attribute -o output_attribute")
	fmt.Fprintln(os.Stderr, "  qjp -d name -d id -T < data.json")
	fmt.Fprintln(os.Stderr, "  cat file.txt | qjp -l")
}

func parseArgs() config {
	cfg := config{
		separator: " - ",
	}

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-d":
			if i+1 < len(args) {
				cfg.displayAttrs = append(cfg.displayAttrs, args[i+1])
				i++
			}
		case "-o":
			if i+1 < len(args) {
				cfg.outputAttr = args[i+1]
				i++
			}
		case "-s":
			if i+1 < len(args) {
				cfg.separator = args[i+1]
				i++
			}
		case "-t":
			cfg.truncate = true
		case "-T":
			cfg.tableMode = true
		case "-l":
			cfg.lineMode = true
		case "-a":
			cfg.allAttrs = true
		case "-h", "--help":
			outputUsage()
			os.Exit(0)
		default:
			if !strings.HasPrefix(args[i], "-") {
				cfg.filename = args[i]
			}
		}
	}

	return cfg
}

func validateConfig(cfg config) error {
	if cfg.allAttrs && len(cfg.displayAttrs) > 0 {
		return fmt.Errorf("cannot use both -a and -d")
	}

	if cfg.lineMode {
		if len(cfg.displayAttrs) > 0 {
			return fmt.Errorf("cannot use -d in line mode")
		}
		if cfg.allAttrs {
			return fmt.Errorf("cannot use -a in line mode")
		}
		if cfg.outputAttr != "" {
			return fmt.Errorf("cannot use -o in line mode")
		}
		if cfg.separator != " - " {
			return fmt.Errorf("cannot use -s in line mode")
		}
		if cfg.truncate {
			return fmt.Errorf("cannot use -t in line mode")
		}
		if cfg.tableMode {
			return fmt.Errorf("cannot use -T in line mode")
		}
	}

	return nil
}

func readInput(filename string) ([]byte, error) {
	stdinStat, _ := os.Stdin.Stat()
	hasStdin := (stdinStat.Mode() & os.ModeCharDevice) == 0

	if hasStdin && filename != "" {
		return nil, fmt.Errorf("cannot use both stdin and filename input")
	}

	if !hasStdin && filename == "" {
		return nil, fmt.Errorf("no input provided")
	}

	if filename != "" {
		return os.ReadFile(filename)
	}

	return io.ReadAll(os.Stdin)
}

func parseObjects(input []byte, lineMode bool) ([]map[string]interface{}, error) {
	var objects []map[string]interface{}

	if lineMode {
		scanner := bufio.NewScanner(bytes.NewReader(input))
		for scanner.Scan() {
			objects = append(objects, map[string]interface{}{
				"line": scanner.Text(),
			})
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("error reading lines: %w", err)
		}
	} else {
		if err := json.Unmarshal(input, &objects); err != nil {
			return nil, fmt.Errorf("error parsing JSON: %w", err)
		}
	}

	if len(objects) == 0 {
		return nil, fmt.Errorf("no objects found in input")
	}

	return objects, nil
}

func getAllAttributes(objects []map[string]interface{}) []string {
	attrMap := make(map[string]bool)
	for _, obj := range objects {
		for key := range obj {
			attrMap[key] = true
		}
	}

	attrs := make([]string, 0, len(attrMap))
	for key := range attrMap {
		attrs = append(attrs, key)
	}
	sort.Strings(attrs)

	return attrs
}

func formatOutputValue(val interface{}) (string, error) {
	switch v := val.(type) {
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v)), nil
		}
		return fmt.Sprintf("%v", v), nil
	case string:
		return v, nil
	case []interface{}, map[string]interface{}:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("error marshaling output: %w", err)
		}
		return string(jsonBytes), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func outputSelectedObjects(objects []map[string]interface{}, indices []int, outputAttr string) error {
	for _, idx := range indices {
		selectedObj := objects[idx]

		if outputAttr != "" {
			val, ok := selectedObj[outputAttr]
			if !ok {
				return fmt.Errorf("attribute '%s' not found in selected object", outputAttr)
			}

			formatted, err := formatOutputValue(val)
			if err != nil {
				return err
			}
			fmt.Println(formatted)
		} else {
			jsonBytes, err := json.Marshal(selectedObj)
			if err != nil {
				return fmt.Errorf("error marshaling output: %w", err)
			}
			fmt.Println(string(jsonBytes))
		}
	}

	return nil
}

func fatalError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

func main() {
	cfg := parseArgs()

	if err := validateConfig(cfg); err != nil {
		fatalError(err.Error())
	}

	input, err := readInput(cfg.filename)
	if err != nil {
		if err.Error() == "no input provided" {
			outputUsage()
		}
		fatalError(err.Error())
	}

	objects, err := parseObjects(input, cfg.lineMode)
	if err != nil {
		fatalError(err.Error())
	}

	displayAttrs := cfg.displayAttrs
	outputAttr := cfg.outputAttr

	if cfg.lineMode {
		displayAttrs = []string{"line"}
		outputAttr = "line"
	} else if cfg.allAttrs {
		displayAttrs = getAllAttributes(objects)
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		fatalError("opening /dev/tty: %v", err)
	}
	defer tty.Close()

	app := newApp(objects, displayAttrs, outputAttr, tty, cfg.truncate, cfg.tableMode, cfg.separator)
	selectedIndices, err := app.run()
	if err != nil {
		fatalError("%v", err)
	}

	if len(selectedIndices) > 0 {
		if err := outputSelectedObjects(app.objects, selectedIndices, outputAttr); err != nil {
			fatalError("%v", err)
		}
	}
}
