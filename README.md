# SnipSnap

A Go CLI utility for safe file operations. `snipsnap` provides atomic file creation, buffered stream copying, separator-based combination, and safe deletion with built-in path traversal safeguards and context cancellation.

---

## Features

* **Atomic File Operations**: Prevents partially written or corrupted files on unexpected failures or context cancellations.
* **Stream Copying & Combining**: Handles large files efficiently using buffered streams with low memory overhead.
* **Safety & Security**: Built-in safeguards against directory overwrites and path traversal vulnerabilities.
* **Automated CI/CD Release Engine**: Cross-compiles native, zero-dependency binaries for Linux, macOS, and Windows across multiple architectures (`amd64`, `arm64`).
* **Conventional Commits & Auto-Versioning**: [release-please](https://github.com/googleapis/release-please) computes the next Semantic Version (including **major** bumps from `feat!:` / `BREAKING CHANGE:`) from Conventional Commits on `main` and maintains a standing release PR with the changelog.

---

## Directory Layout

```text
.
├── .github/
│   ├── workflows/
│   │   ├── ci-checks.yaml            # PR validation (linter, unit tests + race detector)
│   │   └── release.yaml              # release-please versioning + GoReleaser binary publishing
│   ├── release-please-config.json      # release-please versioning/changelog rules
│   └── .release-please-manifest.json   # release-please's current tracked version
├── cmd/
│   └── snipsnap/
│       ├── main.go              # CLI entry point
│       └── main_test.go         # CLI integration tests
├── pkg/
│   └── fileops/
│       ├── ops.go               # Core file operation logic
│       ├── ops_test.go          # Unit tests & race condition safeguards
│       ├── version.go           # Version/Commit/BuildDate vars (injected at release time)
│       └── version_test.go      # Version defaults test
├── .goreleaser.yaml             # Cross-compilation matrix configuration
├── .pre-commit-config.yaml      # Pre-commit hook configuration for conventional commits
├── Makefile                     # Build, test, lint, and hook installation targets
├── go.mod                       # Go module definition
└── README.md
```

---

## Versioning Strategy

Versioning is driven by two vendor-neutral standards, enforced by tooling rather than convention:

* **[Conventional Commits](https://www.conventionalcommits.org/)** — the commit-message format that encodes intent (`feat:`, `fix:`, `feat!:`, `BREAKING CHANGE:`).
* **[Semantic Versioning](https://semver.org/)** — the `MAJOR.MINOR.PATCH` rules that those intents map to.

[release-please](https://github.com/googleapis/release-please) is the tool that implements both: it reads the commit history and computes the correct bump, so a `feat!:` or `BREAKING CHANGE:` commit becomes a real **major** release instead of being silently downgraded.

---

## Quick Start & Installation

### Option 1: Build from Source

**Prerequisites**: Go 1.25+ installed.

```bash
# Clone repository
git clone https://github.com/dchittibala/snipsnap.git
cd snipsnap

# Initialize git pre-commit hooks for conventional commit checks
make init-hooks

# Build binary with version injection
make build

# Verify installation
./bin/snipsnap version
```

### Option 2: Install on Debian / Ubuntu (.deb)

Download the `.deb` for your architecture from the [Releases](https://github.com/dchittibala/snipsnap/releases) page and install it with `dpkg`:

```bash
# Set the version to install (see the Releases page for the latest)
VERSION=1.1.0

curl -LO https://github.com/dchittibala/snipsnap/releases/download/v${VERSION}/snipsnap_${VERSION}_linux_amd64.deb
sudo dpkg -i snipsnap_${VERSION}_linux_amd64.deb

# Verify (installs to /usr/bin/snipsnap)
snipsnap version
# -> snipsnap v1.2.3 (commit: abc1234, built at: 2026-09-01T12:00:00Z)
```

---

## CLI Usage Examples

### 1. Create a File
```bash
# Create an empty file
snipsnap create path/to/file.txt

# Create a file with initial text content
snipsnap create -text="Hello, World!" path/to/file.txt

# Overwrite an existing file safely using force flag
snipsnap create -force -text="New Content" path/to/file.txt
```

### 2. Copy a File
```bash
# Copy file to a target destination
snipsnap copy source.txt destination.txt

# Force overwrite if target exists
snipsnap copy -force source.txt destination.txt
```

### 3. Combine Multiple Files
```bash
# Combine two files into a destination file using a custom separator
snipsnap combine -sep=" --- " fileA.txt fileB.txt output.txt

# Combine in-place using force flag
snipsnap combine -force -sep="\n" fileA.txt fileB.txt fileA.txt
```

### 4. Delete a File
```bash
# Safely remove a file (refuses to delete directories)
snipsnap delete path/to/file.txt
```

### 5. Print Version Information
```bash
# Show the version, commit, and build date injected at release time
snipsnap version
# -> snipsnap v1.2.3 (commit: abc1234, built at: 2026-09-01T12:00:00Z)

---

## Local Development & Testing

All operational tasks are managed via the included `Makefile`:

| Target | Command | Description |
| :--- | :--- | :--- |
| **`make build`** | `go build ...` | Compiles local binary to `bin/snipsnap` with Git metadata injected. |
| **`make test`** | `go test -v ./...` | Executes standard unit test suite. |
| **`make test-race`** | `go test -v -race -cover ./...` | Runs test suite with Go race detection and coverage mapping. |
| **`make lint`** | `golangci-lint run ./...` | Executes standard static code analysis checks. |
| **`make init-hooks`** | `pre-commit install ...` | Installs `pre-commit` binary and activates git hooks locally. |
| **`make verify-hooks`** | `pre-commit run --all-files` | Runs conventional commit validation against all tracked files. |
| **`make clean`** | `rm -rf bin/ dist/` | Removes local build and distribution artifacts. |

---

## Conventional Commits & Pre-commit Hooks

This repository uses `pre-commit` to enforce [Conventional Commits](https://www.conventionalcommits.org/) locally before code is committed.

### Installation & Setup

To install `pre-commit` (via `brew` or `pip`) and set up the Git hooks:

```bash
make init-hooks
```

### Format Rule
```text
<type>(<optional scope>): <description>
```

### Allowed Types
* `feat`: A new feature (triggers a **MINOR** release bump)
* `fix`: A bug fix (triggers a **PATCH** release bump)
* `docs`: Documentation changes
* `style`: Code style/formatting changes
* `refactor`: Code refactoring without behavioral changes
* `perf`: Performance improvements
* `test`: Adding or modifying unit tests
* `build`: Build system or dependency updates
* `ci`: CI configuration changes
* `chore`: Maintenance tasks
* `revert`: Reverting a previous commit

To introduce a **MAJOR** breaking change, add `!` after the type (e.g., `feat!: change CLI flag syntax`) or include `BREAKING CHANGE:` in the body.

---

## CI/CD Pipeline Architecture

### 1. Pull Request Stage (`.github/workflows/ci-checks.yaml`)
Runs automatically on PRs targeting `main`:
* Runs `golangci-lint` to enforce code quality.
* Runs `go test -race -cover ./...` to detect race conditions and verify unit test coverage.

### 2. Release Stage (`.github/workflows/release.yaml`)
A single workflow, run on every push to `main` (i.e. on every merge), with two jobs:

**`release-please` job** — [release-please](https://github.com/googleapis/release-please) evaluates Conventional Commit messages since the last release and opens or updates a standing "release PR" with the computed next version and generated `CHANGELOG.md`.
* Bump detection correctly implements the Conventional Commits spec: `feat!:` / `fix!:` / etc. (the `!` breaking-change marker) or a `BREAKING CHANGE:` footer → **major**; `feat:` → **minor**; everything else (including `fix:`) → **patch**.
* Merging that release PR is what actually cuts the SemVer tag and GitHub Release — this is also the point to force an on-demand release of a specific version regardless of the computed bump, via a `Release-As: X.Y.Z` commit footer.

**`publish` job** — runs in the same workflow run, but only when the `release-please` job reports `release_created == true` (i.e. the merge just cut a release):
* Checks out the exact released tag and triggers **GoReleaser** to build static binaries across Linux, macOS, and Windows (`amd64`, `arm64`).
* Generates SHA-256 checksums and attaches the distribution archives to that same GitHub Release (`release.mode: keep-existing` in `.goreleaser.yaml`, so it never overwrites release-please's changelog or creates a duplicate release).

> The publish job is gated on release-please's output rather than a separate `on: release` trigger on purpose: release-please creates the release with the default `GITHUB_TOKEN`, and GitHub does not fire workflow events for `GITHUB_TOKEN` actions, so a standalone release-triggered workflow would never run.