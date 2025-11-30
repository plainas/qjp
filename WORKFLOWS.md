# GitHub Actions Workflows

This project includes three GitHub Actions workflows for automated building, testing, and releasing.

## Workflows Overview

### 1. CI Workflow (`ci.yml`)

**Trigger**: On every push to `main`/`master` branch and pull requests

**Purpose**: Continuous integration testing to ensure the code builds successfully on all platforms

**What it does**:
- Tests building on Ubuntu, macOS, and Windows runners
- Cross-compiles for all supported platforms (Linux, macOS, Windows, FreeBSD) with multiple architectures
- Verifies the binary is created successfully

**Platforms tested**:
- Linux: amd64, arm64
- macOS: amd64, arm64
- Windows: amd64
- FreeBSD: amd64

### 2. Release Workflow (`release.yml`)

**Trigger**:
- When a tag starting with `v` is pushed (e.g., `v1.0.0`)
- Manual workflow dispatch

**Purpose**: Build release binaries for all platforms and create a GitHub release

**What it does**:
- Builds binaries for all supported platforms
- Uploads raw binaries directly (no archives)
- Creates a GitHub release with all binaries attached

**Release naming**: `qjp-{os}-{arch}` (or `.exe` for Windows)

### 3. GoReleaser Workflow (`goreleaser.yml`) - **RECOMMENDED**

**Trigger**: When a tag starting with `v` is pushed (e.g., `v1.0.0`)

**Purpose**: Automated release using GoReleaser (cleaner and more powerful)

**What it does**:
- Uses GoReleaser to handle the entire release process
- Builds for all platforms defined in `.goreleaser.yml`
- Uploads raw binaries directly (no archives)
- Creates checksums
- Generates changelog automatically
- Creates GitHub release with proper formatting

**Configuration**: See `.goreleaser.yml` for platform and binary settings

## How to Create a Release

### Using GoReleaser (Recommended)

1. Make sure your code is committed and pushed
2. Create and push a version tag:
   ```bash
   git tag -a v1.0.0 -m "Release version 1.0.0"
   git push origin v1.0.0
   ```
3. The `goreleaser.yml` workflow will automatically run and create the release

### Using Manual Release Workflow

1. Go to Actions tab on GitHub
2. Select "Release" workflow
3. Click "Run workflow"
4. Select the branch
5. Click "Run workflow" button

Or push a tag:
```bash
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

## Which Workflow Should I Use?

- **For releases**: Use the GoReleaser workflow (most automated and clean)
- **For testing**: CI workflow runs automatically on every push
- **For custom builds**: Manual release workflow if you need more control

## Disabling Unused Workflows

If you want to use only GoReleaser:

1. Delete or disable `release.yml`:
   ```bash
   rm .github/workflows/release.yml
   ```
2. Keep `ci.yml` for testing and `goreleaser.yml` for releases

## Customizing Build Platforms

### For GoReleaser

Edit `.goreleaser.yml`:
```yaml
builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
      # Add more OSes here
    goarch:
      - amd64
      - arm64
      # Add more architectures here
```

### For Manual Release Workflow

Edit `.github/workflows/release.yml`:
```yaml
strategy:
  matrix:
    include:
      - goos: linux
        goarch: amd64
      # Add more combinations here
```

## Supported Platforms

- **Linux**: amd64, arm64, arm (armv7)
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **Windows**: amd64
- **FreeBSD**: amd64

All binaries are statically compiled with `CGO_ENABLED=0` for maximum portability.
