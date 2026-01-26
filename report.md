# Polyenv Project Improvement Report

## Executive Summary

This report outlines key findings and recommendations for improving the cleanliness, maintainability, and user experience of the `polyenv` project. The analysis focused on code structure, architectural patterns, and documentation.

The project is a Go-based CLI for managing secrets. Its core feature is a dynamic command structure where `!{env}` commands are generated at runtime based on the presence of `{env}.polyenv.toml` configuration files.

While the project is functional, several architectural weaknesses were identified that make the code brittle, difficult to test, and harder to maintain. The key areas for improvement are:

1.  **Error Handling**: Widespread use of `log.Fatal` or `os.Exit` within internal packages makes the code untestable and non-reusable.
2.  **Global State**: The dynamic commands rely on a global variable, creating tight coupling and making the code harder to reason about.
3.  **Configuration**: Inconsistent handling of the `--debug` flag.

This report provides specific, actionable recommendations to address these issues.

## Key Architectural Issues

### 1. Improper Error Handling

The codebase frequently uses `log.Error` followed by `os.Exit(1)` inside command and internal packages. This is a significant anti-pattern in Go applications.

*   **Problem**: When a function calls `os.Exit`, it immediately terminates the program. This prevents proper cleanup, makes the function impossible to test (as the test will exit), and prevents the function from being reused in other contexts. Errors should be returned up the call stack to `main`, which is the only place the program should exit.
*   **Example**: In `internal/polyenvfile/main.go`, the `OpenFile` function calls `os.Exit(1)` if the file cannot be read.

### 2. Use of Global State

The dynamic `!{env}` commands rely on a global variable (`cmd.PolyenvFile`) that is set in a `PersistentPreRun` hook and read by the command's `Run` function.

*   **Problem**: This reliance on a global variable creates a hidden dependency between the `PersistentPreRun` function and the command's `Run` function. This makes the code difficult to understand, test, and maintain. It's not clear where the state is coming from without tracing the execution flow of Cobra commands.
*   **Example**: In `cmd/current-pull.go`, the `pull` command's `Run` function directly accesses `PolyenvFile`, which is set in `cmd/current.go`.

### 3. Inconsistent Configuration

The `--debug` flag is parsed manually in `main.go` from `os.Args`.

*   **Problem**: The Cobra library, which is already used in the project, has robust support for flag parsing. Manually parsing `os.Args` is redundant and can lead to inconsistencies. All flags should be defined and parsed using Cobra for a consistent user experience.
*   **Example**: In `main.go`, there is a loop that checks for `--debug`. This should be handled by Cobra in the `cmd` package.

## Recommendations

1.  **Refactor Error Handling**:
    *   Replace all instances of `log.Fatal`, `log.Fatalln`, `log.Fatalf`, and `os.Exit` in `internal/` and `cmd/` packages with `return err`.
    *   Propagate errors up to the `main` function in `cmd/main.go`.
    *   In `main`, handle the returned error by printing it to `stderr` and exiting with a non-zero status code.

2.  **Eliminate Global State**:
    *   Refactor the dynamic command generation to avoid global variables.
    *   One approach is to use a closure when creating the commands, so that the `polyenvfile` is captured by the `Run` function.

3.  **Centralize Flag Parsing**:
    *   Remove the manual parsing of the `--debug` flag from `main.go`.
    *   Define the `--debug` flag as a persistent flag on the `rootCmd` in `cmd/main.go` using Cobra.

By addressing these architectural issues, the `polyenv` project will be more robust, easier to test, and more maintainable in the long run.

## Implementation Plan

Here is a multi-step plan to implement the recommendations:

### Step 1: Refactor Error Handling

This step will be broken down into smaller sub-steps to address each part of the codebase.

1.  **`internal/polyenvfile/main.go`**:
    *   Modify `OpenFile` to return `(*File, error)` instead of `*File`.
    *   In `OpenFile`, change `log.Errorf(...)` and `os.Exit(1)` to `return nil, err`.
    *   Modify `File.Save()` to return an `error`.
    *   Update all places where `OpenFile` and `File.Save()` are called to handle the returned error.
2.  **`cmd/current.go`**:
    *   In the `init` function, when `OpenFile` is called, handle the error. If there is an error, `log.Fatal` is acceptable here as it's in the `init` phase.
3.  **`cmd/current-pull.go`**:
    *   Change the `Run` function for the `pullCmd` to return an `error`.
    *   In the `Run` function, where `vault.Pull` is called, if there is an error, return it instead of calling `log.Error` and `os.Exit(1)`.
4.  **`cmd/current-export.go`**:
    *   Change the `Run` function for the `exportCmd` to return an `error`.
    *   Where `ExportEnv` is called, return the error if it's not nil.
5.  **`cmd/main.go`**:
    *   Update the `Execute` function to handle the errors returned from the commands.
    *   If an error is returned, print it to `os.Stderr` and call `os.Exit(1)`.

### Step 2: Eliminate Global State

1.  **`cmd/current.go`**:
    *   Modify the `init` function. The loop that creates the dynamic commands needs to be changed.
    *   Instead of `rootCmd.AddCommand(currentCmd)`, create a new command for each environment.
    *   The `RunE` (or `Run`) function for these new commands should be a closure that captures the `polyenvFile` for the specific environment.
    *   This will involve creating a function that returns a `func(cmd *cobra.Command, args []string) error`.
2.  **`cmd/current-pull.go`, `cmd/current-export.go`, etc.**:
    *   The `PolyenvFile` global variable will no longer be used.
    *   The `polyenvFile` will be passed to the functions that need it, or be available within the closure.

### Step 3: Centralize Flag Parsing

1.  **`main.go`**:
    *   Remove the `for` loop that parses `os.Args` for the `--debug` flag.
    *   Remove the `init` function that sets up logging.
2.  **`cmd/main.go`**:
    *   In the `init` function, add `rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")`.
    *   Add a `PersistentPreRun` to the `rootCmd` to configure the logging based on the `debug` flag.

This plan provides a clear path to refactor the codebase. I recommend tackling one step at a time.