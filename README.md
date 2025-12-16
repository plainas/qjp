# qjp

`qjp` (quick json picker) is an interactive command-line tool for filtering and selecting JSON objects or plain text lines. It provides a quick unix-pipeline friendly way to add an interactive menu to your shellscripts.

Feed it a JSON array via stdin or from a file (or plain text with `-l`), optionally specify which field(s) to display, and qjp will present an interactive list.
Type to filter, use arrow keys to navigate, press Ctrl+Space to multi-select, press Enter to output your selection - either as complete JSON objects or just specific field values.

## Features

- Interactive filtering and selection of JSON objects or plain text lines
- Control over what gets displayed and what gets output.
- Real-time filtering as you type
- Multi-select support with Ctrl+Space
- Read from stdin or directly from a file
- Display one or multiple attributes while browsing
- Table mode, displaying attributes vertically aligned for readability
- Line mode. Ignore json, behave like percol
- Optional line truncate for long content. Wraps lines otherwise.
- Output the entire selected object(s) or a specific attribute
- Arrays and objects output as single-line JSON

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

#### Install the manpage on your system (optional)

```bash
sudo wget -O /usr/local/share/man/man1/qjp.1 "https://github.com/plainas/qjp/releases/latest/download/qjp.1"
sudo mandb
```


### Build from Source

```bash
go build -o qjp
```

## Usage

```
qjp [filename] [-d display-attribute] [-o output-attribute] [-s separator] [-t] [-T] [-l] [-a]
qjp [-d display-attribute] [-o output-attribute] [-s separator] [-t] [-T] [-l] [-a] < input
```

### Arguments

- `filename`: (optional) JSON file to read (or plain text with `-l`). If not provided, reads from stdin.
- `-d <attribute>`: Display specific attribute(s) in list (can be used multiple times for multiple attributes)
- `-o <attribute>`: Output specific attribute from selected object(s). Arrays and objects are output as single-line JSON.
- `-s <separator>`: Separator for multiple display attributes (default: " - ")
- `-t`: Truncate long lines instead of wrapping
- `-T`: Table mode - align attributes in columns
- `-l`: Line mode - treat input as plain text lines (like percol). Cannot be used with `-d`, `-o`, `-s`, `-t`, `-T`, or `-a`.
- `-a`: Display all attributes - automatically discover and display all unique attributes from all objects in alphabetical order. Cannot be used with `-d` or `-l`. Particularly useful with `-T` for a structured overview.
- `-h, --help`: Show help message

**Note:** Input can be provided via stdin or filename, but not both.

For detailed usage information, see the man page:
```bash
man ./qjp.1
# Or after installation:
man qjp
```

## Keyboard Controls

- **Type**: Filter the list in real-time
- **Up/Down arrows**: Navigate through the list
- **Ctrl+Space**: Toggle selection (multi-select mode - selected items shown with green background)
- **Enter**: Confirm selection (outputs selected item(s))
- **Backspace**: Delete the last character from the filter
- **Esc** or **Ctrl+C**: Exit without selecting

## Examples

Basic example below use the sample `cars.json` included in the source.


```bash
##############################################################
## Basic usage
##############################################################

# Read from file and display entire objects.
# Outputs entire objects one per line
qjp cars.json

# Read from file and display only car models name when selecting
qjp cars.json -d model

# Display car models and output only the ID
qjp cars.json -d model -o id

# Display multiple attributes (year, make and model) with default separator " - "
qjp cars.json -d year -d make -d model -o price

# Display multiple attributes with custom separator
qjp cars.json -d model -d year -s " | "

# Display multiple attributes in table mode (aligned columns)
qjp cars.json -d make -d model -d year -T

# Display all attributes (discovers all keys automatically)
qjp cars.json -a

# Display all attributes in table mode
# Useful for quickly inspecting the contents of a file
qjp cars.json -a -T

# Truncate long lines.
# this can be usefull while working with large objects
qjp cars.json -t

# Read from stdin (traditional pipe usage)
cat cars.json | qjp -d model

# Display car make and model and output the price
cat cars.json | qjp -d make -d model -o price

# Line mode: select from plain text lines (like percol)
ls -la | qjp -l
cat file.txt | qjp -l
ps aux | qjp -l


##############################################################
## Advanced examples
##############################################################

# Select one of Linus Torvalds repositories on github and output its number of stars
curl -s "https://api.github.com/users/torvalds/repos" | qjp -d name -d description -o stargazers_count

# A country picker that display common names and outpouts two leter country codes.
curl -s "https://www.apicountries.com/countries" | qjp -d name -o alpha2Code

# Minimal lobste.rs reased on your terminal
curl -s "https://lobste.rs/hottest.json" | ./qjp -d title -o url | xargs lynx

# Get the ID of a docker container by name. Useful to feed into other commands
docker ps --format json | jq -s '.' | qjp -d Names -d Status -o ID

# a quick way to find and install npm packages
npm search --json typescript | qjp -d name -d description -o name | xargs npm install

# Pick a cryptocurrency and retrieve its details from coingecko
curl -s "https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=50" | qjp -d name -d symbol -d current_price | jq .
```



## Development

### GitHub Actions Workflows

This project includes GitHub Actions workflows for automated building and releasing:

1. **CI Workflow** (`.github/workflows/ci.yml`): Runs on every push/PR to test builds on multiple platforms
2. **Release Workflow** (`.github/workflows/release.yml`): Manual release workflow with custom build matrix
3. **GoReleaser Workflow** (`.github/workflows/goreleaser.yml`): Automated releases using GoReleaser (recommended)

To create a new release:
```bash
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
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


# TODO

 * Support jq syntax
 * Add support for jsonlines input
 * Add a classifier to automatically detect input format
 * output as json array
 * add option to output single values as json encoded

## License

MIT License - see [LICENSE](LICENSE) file for details.