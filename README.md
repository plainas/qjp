# qjp

`qjp` (quick json picker) is an interactive command-line tool for filtering and selecting JSON objects. It provides a quick unix-pipeline friendly way to add an interactive menu to your shellscript.

Feed it a JSON array via stdin, specify which field to display, and qjp will present an interactive list.
Type to filter, use arrow keys to navigate, and press Enter to output your selection - either as a complete JSON object or just a specific field value.

## Features

- Interactive filtering and selection of JSON objects
- Display a specific attribute while browsing
- Output the entire selected object or a specific attribute
- Real-time filtering as you type
- Navigate with arrow keys

## Installation

### Download Pre-built Binaries

Download the binary directly for your platform from the [Releases](https://github.com/plainas/qjp/releases) page.


#### Quick install scripts
```sh
# Linux (x86_64)
sudo wget -O /usr/local/bin/qjp "https://github.com/plainas/qjp/releases/latest/download/qjp-linux-x86_64"
sudo chmod +x /usr/local/bin/qjp

# Linux (ARM64 / aarch64)
sudo wget -O /usr/local/bin/qjp "https://github.com/plainas/qjp/releases/latest/download/qjp-linux-arm64"
sudo chmod +x /usr/local/bin/qjp

# Linux (ARMv7)
sudo wget -O /usr/local/bin/qjp "https://github.com/plainas/qjp/releases/latest/download/qjp-linux-armv7"
sudo chmod +x /usr/local/bin/qjp

# macOS (Intel)
sudo wget -O /usr/local/bin/qjp "https://github.com/plainas/qjp/releases/latest/download/qjp-darwin-x86_64"
sudo chmod +x /usr/local/bin/qjp


# macOS (Apple Silicon)
sudo wget -O /usr/local/bin/qjp "https://github.com/plainas/qjp/releases/latest/download/qjp-darwin-arm64"
sudo chmod +x /usr/local/bin/qjp
```


### Build from Source

```bash
# Build only
go build -o qjp
```


## Usage

```
qjp <display-attribute> [-o output-attribute] < input.json

# Arguments

  - `<display-attribute>`: (required) The attribute to display for each object while browsing
  - `-o <output-attribute>`: (optional)The attribute to output when an object is selected.
    If output-attribute is not provided, the whole selected object will be output.
```

For detailed usage information, see the man page:
```bash
man ./qjp.1
# Or after installation:
man qjp
```

## Examples

```bash
# Display car models and output the entire selected object
# This will show you a list of car models. When you select one, it outputs the entire JSON object on a single line.
cat cars.json | ./qjp model

# Display car models and output only the ID
# This will show you a list of car models. When you select one, it outputs only the `id` field value.
cat cars.json | ./qjp model -o id

# Display car make and output the price
cat cars.json | ./qjp make -o price
```

## Keyboard Controls

- **Type**: Filter the list in real-time
- **Up/Down arrows**: Navigate through the list
- **Enter**: Select the current item
- **Backspace**: Delete the last character from the filter
- **Esc** or **Ctrl+C**: Exit without selecting

## Development

### GitHub Actions Workflows

This project includes GitHub Actions workflows for automated building and releasing:

1. **CI Workflow** (`.github/workflows/ci.yml`): Runs on every push/PR to test builds on multiple platforms
2. **Release Workflow** (`.github/workflows/release.yml`): Manual release workflow with custom build matrix
3. **GoReleaser Workflow** (`.github/workflows/goreleaser.yml`): Automated releases using GoReleaser (recommended)

To create a new release:
```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

The GoReleaser workflow will automatically build binaries for all supported platforms and create a GitHub release.

### Cross-compilation

To build for a specific platform locally:

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o qjp-linux-amd64

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o qjp-darwin-arm64

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o qjp-windows-amd64.exe
```

## License

MIT License - see [LICENSE](LICENSE) file for details.