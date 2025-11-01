# Script Utilities

This directory contains utility scripts for the Agent SDK Go project.

## Go Report Card Fixer

The `fix_goreportcard.sh` script identifies and fixes common issues reported by [Go Report Card](https://goreportcard.com/report/github.com/Muhammadhamd/agent-sdk-go/).

### Usage

```bash
# Make the script executable (if not already)
chmod +x scripts/fix_goreportcard.sh

# Run the script
./scripts/fix_goreportcard.sh
```

### What it Does

The script:

1. Installs necessary Go linting tools
2. Formats code with `gofmt`
3. Fixes common misspellings
4. Identifies ineffectual assignments
5. Finds unchecked errors
6. Runs standard `go vet` checks
7. Applies Go linting rules
8. Checks cyclomatic complexity
9. Runs static analysis
10. Finds unconverted types

### Troubleshooting

If you encounter path-related errors when running the tools:

1. Make sure your `GOPATH/bin` directory is in your PATH:
   ```bash
   export PATH="$(go env GOPATH)/bin:$PATH"
   ```

2. Or run the tools with their full path:
   ```bash
   $(go env GOPATH)/bin/misspell -w your_file.go
   ```

## Pre-commit Setup

We recommend setting up pre-commit hooks to automatically check code quality before committing:

1. Install pre-commit:
   ```bash
   pip install pre-commit
   ```

2. Install the hooks:
   ```bash
   pre-commit install
   ```

3. Now pre-commit will run automatically on `git commit`

## Other Scripts

(Other script documentation will be added here as more scripts are developed) 