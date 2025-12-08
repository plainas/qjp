# qjp

`qjp` (quick json picker) is an interactive command-line tool for filtering and selecting JSON objects. It provides a quick unix-pipeline friendly way to add an interactive menu to your shellscripts.

Feed it a JSON array via stdin or from a file, optionally specify which field(s) to display, and qjp will present an interactive list.
Type to filter, use arrow keys to navigate, press Ctrl+Space to multi-select, and press Enter to output your selection - either as complete JSON objects or just specific field values.

## Features

- Interactive filtering and selection of JSON objects
- Navigate with arrow keys
- Multi-select support with Ctrl+Space
- Ctrl+C or ESC to exit without selecting anything
- Read from stdin or directly from a file
- Display one or multiple attributes while browsing
- Customizable display separator for multiple attributes
- Optional line truncation for long content
- Output the entire selected object(s) or a specific attribute
- Real-time filtering as you type


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
qjp [filename] [-d display-attribute] [-o output-attribute] [-s separator] [-t]
qjp [-d display-attribute] [-o output-attribute] [-s separator] [-t] < input.json
```

### Arguments

- `filename`: (optional) JSON file to read. If not provided, reads from stdin.
- `-d <attribute>`: Display specific attribute(s) in list (can be used multiple times for multiple attributes)
- `-o <attribute>`: Output specific attribute from selected object(s)
- `-s <separator>`: Separator for multiple display attributes (default: " - ")
- `-t`: Truncate long lines instead of wrapping
- `-h, --help`: Show help message

**Note:** Input can be provided via stdin or filename, but not both.

For detailed usage information, see the man page:
```bash
man ./qjp.1
# Or after installation:
man qjp
```

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

# Truncate long lines.
# this can be usefull while working with large objects
qjp cars.json -t

# Read from stdin (traditional pipe usage)
cat cars.json | qjp -d model

# Display car make and model and output the price
cat cars.json | qjp -d make -d model -o price


##############################################################
## Advanced examples
##############################################################

# Select one of Linus Torvalds repositories on github and output its number of stars
curl -s "https://api.github.com/users/torvalds/repos" | qjp -d name -d description -o stargazers_count

# Interactively pick one of your EC2 instances and restart it
aws ec2 reboot-instances --instance-ids $(aws ec2 describe-instances | jq '.Reservations[].Instances[]' | qjp -d State.Name -o InstanceId)

# Select a Kubernetes pod to inspect
kubectl get pods -o json | jq '.items[]' | qjp -d metadata.name -d status.phase -o metadata.name

# Choose a Docker container
docker ps --format json | jq -s '.' | qjp -d Names -d Status -o ID

# Browse npm packages
npm search --json typescript | jq -s '.' | qjp -d name -d description -o name

# Pick a cryptocurrency
curl -s "https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=50" | qjp -d name -d symbol -d current_price

# Browse Hacker News stories
curl -s "https://hacker-news.firebaseio.com/v0/topstories.json" | jq '.[0:5]' | xargs -I {} curl -s "https://hacker-news.firebaseio.com/v0/item/{}.json" | jq -s '.' | qjp -d title -d by -o url

# Select a country from REST Countries API
curl -s "https://restcountries.com/v3.1/all" | qjp -d name.common -d capital -o cca2

# Interactive TODO list management
curl -s "https://jsonplaceholder.typicode.com/todos" | qjp -d title -d completed -o id


```

## Keyboard Controls

- **Type**: Filter the list in real-time
- **Up/Down arrows**: Navigate through the list
- **Ctrl+Space**: Toggle selection (multi-select mode - selected items shown with green background)
- **Enter**: Confirm selection (outputs selected item(s))
- **Backspace**: Delete the last character from the filter
- **Esc** or **Ctrl+C**: Exit without selecting

## FAQs
  
**Q: Why did you do this?**

**Q: What can I use qjp for?**

**Q: How do I select more than one attribute to the output?**

**A:** `qjp` is designed to output complete JSON objects or single attributes. If you need to extract multiple attributes in the output, pipe the result through `jq`. For example:
```bash
qjp cars.json -d model | jq '{make, model, price}'
```

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