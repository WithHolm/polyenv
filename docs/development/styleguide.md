# Development Style Guide

This document outlines the coding and documentation style for the Polyenv project.

## General Principles

*   Clarity and simplicity are paramount.
*   Follow existing conventions in the codebase.
*   When in doubt, ask for clarification.

## Go Programming Language

We follow the official Go style guides. Please familiarize yourself with them:

*   [Effective Go](https://go.dev/doc/effective_go)
*   [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### Naming Conventions

*   **Variable Names:** Variable names should be at least 3 characters long. Single-letter variables are discouraged unless they are in a very small scope (e.g., `i` in a for loop).

### Code Formatting

*   **Line Length:** The maximum line length is 120 characters. However, it is preferred to keep lines at 80 characters for better readability.

### TUI Development (`huh` library)

*   **Group Comments:** When creating a new form with `huh.NewForm`, any `huh.NewGroup` should have a comment above it that describes the purpose and contents of the group.

    ```go
    // This group contains fields for configuring the database connection.
    huh.NewGroup(
        huh.NewInput().
            Title("Host").
            Value(&db.Host),
        // ...
    ),
    ```

## Git and Commit Messages

*   **Commit Messages:** We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification. This helps in automating changelog generation and makes the commit history more readable.

## Markdown and Documentation

*   **Markdown Style:** We use standard GitHub Flavored Markdown for all our documentation.
