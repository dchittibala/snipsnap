# SnipSnap

A Go CLI utility for safe file operations. `snipsnap` provides atomic file creation, buffered stream copying, separator-based combination, and safe deletion with built-in path traversal safeguards and context cancellation.

---

## Features

* **Atomic File Operations**: Prevents partially written or corrupted files on unexpected failures or context cancellations.
* **Stream Copying & Combining**: Handles large files efficiently using buffered streams with low memory overhead.
* **Safety & Security**: Built-in safeguards against directory overwrites and path traversal vulnerabilities.
* **Automated CI/CD Release Engine**: Cross-compiles native, zero-dependency binaries for Linux, macOS, and Windows across multiple architectures (`amd64`, `arm64`).
* **Conventional Commits & Auto-Versioning**: Automatically calculates Semantic Versioning tags (`v1.0.0`, `v1.0.1`) on `master` branch merges.

---

## Directory Layout

```text
.
├── .github/
│   └── workflows/
│       ├── pr-checks.yml        # PR validation (linter, unit tests + race detector)
│       └── release.yml          # Auto-bump tagger and GoReleaser pipeline
├── cmd/
│   └── snipsnap/
│       ├── main.go              # CLI entry point
│       └── main_test.go         # CLI integration tests
├── pkg/
│   └── fileops/
│       ├── ops.go               # Core file operation logic
│       └── ops_test.go          # Unit tests & race condition safeguards
├── .goreleaser.yaml             # Cross-compilation matrix configuration
├── .pre-commit-config.yaml      # Pre-commit hook configuration for conventional commits
├── Makefile                     # Build, test, lint, and hook installation targets
├── go.mod                       # Go module definition
└── README.md
```

---

## Quick Start & Installation

### Option 1: Build from Source

**Prerequisites**: Go 1.24+ installed.

```bash
# Clone repository
git clone [https://github.com/dchittibala/snipsnap.git](https://github.com/dchittibala/snipsnap.git)
cd snipsnap

# Initialize git pre-commit hooks for conventional commit checks
make init-hooks

# Build binary with version injection
make build

# Verify installation
./bin/snipsnap --version
```

### Option 2: Download Compiled Binaries

Download pre-compiled static binaries for your OS and architecture directly from the [GitHub Releases](https://github.com/dchittibala/snipsnap/releases) page.

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

### 1. Pull Request Stage (`.github/workflows/pr-checks.yml`)
Runs automatically on PRs targeting `main`:
* Runs `golangci-lint` to enforce code quality.
* Runs `go test -race -cover ./...` to detect race conditions and verify unit test coverage.

### 2. Auto-Bump & Release Stage (`.github/workflows/release.yml`)
Runs automatically on code merges to `main`:
* Evaluates Conventional Commit messages since the last git tag.
* Calculates and pushes the next Semantic Version tag (`v1.0.0` → `v1.0.1`).
* Triggers **GoReleaser** to build static binaries across Linux, macOS, and Windows (`amd64`, `arm64`).
* Generates SHA-256 checksums and attaches distribution archives to a new GitHub Release.