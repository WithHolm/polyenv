# Gemini Interaction Guide

This document provides guidelines for the Gemini AI to follow when interacting with this project.

## Project Overview

Polyenv is a command-line interface (CLI) tool designed to manage environment variables and secrets across different "vault" backends. It helps synchronize configurations from sources like local files, Azure Key Vault, and others into a consistent `.env` file for development.

## File & Directory Structure

- `cmd/`: Contains the implementation for each of the CLI commands (e.g., `polyenv init`, `polyenv current-add`).
- `internal/`: Holds the core application logic.
  - `internal/model`: Defines the core data structures like `Secret` and `Vault`.
  - `internal/polyenvfile`: Manages the reading and writing of the `polyenv.yaml` configuration file.
  - `internal/vaults`: Contains the logic for interacting with different vault backends (e.g., Azure Key Vault, local files).
- `docs/`: Contains project documentation, usage guides, and demos.
- `scripts/`: Includes helper scripts for automation and project tasks.

## General Interaction

- Be polite and maintain a neutral but somewhat friendly tone.
- If you think an idea is good or bad, say so and explain why.
- If you are unsure about a task, ask for clarification. Deny requests if you are less than 80% certain you can complete them correctly.
- Before making any changes to files, always ask for user permission, however web fetch and google search is allowed without permission.
  - never try to edit files without user explicitly asking for it. you can say you can edit the file, or offer suggestion of changes if its not too complex.
- when generating new test files, explain you general thought beforehand and ask for concent to start work. you can then create new `*_test.go` files without asking for concent. you can also run `go test` without asking for concent.

## Documentation

- The primary sources of documentation are the root `README.md` and the files within the `./docs` directory.

## Development Workflow

### Project Commands

Use these commands for common development tasks:

- **Build:** `goreleaser release --clean --snapshot --skip publish`
- **Test:** `go test ./internal/... ./cmd/... -coverprofile=coverage.out`

### Git Conventions

- **Commit Messages:** Please follow the Conventional Commits specification.