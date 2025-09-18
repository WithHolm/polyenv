# Project Workflow Automation

This document outlines the automated CI/CD workflows configured in this project's `.github/workflows` directory.

## Pull Request to `dev` Branch Workflow

When a pull request is opened targeting the `dev` branch, the following automated workflows are triggered to validate the changes:

### 1. Labeling (`pull_request_labeler.yml`)

- **Trigger:** On any pull request.
- **Action:** Automatically applies labels (e.g., `documentation`, `go`, `ci`) to the pull request based on the paths of the files that were modified.

### 2. Code Validation (`pr-to-dev-test-and-lint.yml`)

- **Trigger:** On pull requests to the `dev` branch.
- **Action:**
    - Checks out the code from the pull request.
    - Runs the full suite of Go tests (`go test`).
    - Runs the linter to check for code style and quality issues.
    - This is a required check for merging.

### 3. Changelog Notes (`pr-to-dev-grab-update-notes.yml`)

- **Trigger:** On pull requests to the `dev` branch.
- **Action:**
    - Validates if a changelog fragment (`changelog/<pr-number>.md`) already exists.
    - If it doesn't, it extracts the changelog description from the PR body (from under the `## Description` heading).
    - It then creates a new changelog fragment file and commits it back to the PR branch.

### 4. Dependabot Auto-Merge (`pr-to-dev-dependabot-handler.yml`)

- **Trigger:** On pull requests to the `dev` branch.
- **Action:** Specifically for PRs created by Dependabot. If all status checks (like the test and lint workflow) pass, this workflow will automatically merge the pull request.

## Post-Merge and Other Workflows

### Development Release (`release-dev.yml`)

- **Trigger:** On every push to the `dev` branch (i.e., after a PR is merged).
- **Action:** Creates a "pre-release" version using `goreleaser`. This is used for testing and development versions.

### Main Release (`release-main.yml`)

- **Trigger:** On every push to the `main` branch.
- **Action:** Creates an official new release of the application using `goreleaser`.

### Manual Changelog Creation (`create-cl-from-fragments.yml`)

- **Trigger:** Manually via the GitHub Actions UI (`workflow_dispatch`).
- **Action:** Gathers all the individual changelog fragments and compiles them into the main `changelog.md` file.
