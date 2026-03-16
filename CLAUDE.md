# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`gh-ap` is a GitHub CLI extension that adds Issues or Pull Requests to GitHub Projects (V2) with interactive custom field updates. It uses `gh` CLI as a host and communicates with GitHub via GraphQL API.

## Build & Run

```bash
# Build
go build -o gh-ap

# Run (must be in a git repo with gh CLI authenticated)
gh ap
gh ap -issue 123
gh ap -pr 456

# With field values
gh ap -issue 123 -field "Status=Done" -field "Priority=High"

# Run tests
go test ./...
```

## Architecture

All code is in `package main` with no subdirectories (aside from `sandbox/` which contains standalone GraphQL exploration scripts).

| File | Role |
|------|------|
| `main.go` | Entry point, CLI flag parsing, orchestrates the interactive flow |
| `cli.go` | Wraps `gh` CLI commands (repo view, pr view, issue view/list) via `gh.Exec` |
| `query.go` | GraphQL queries: user/org projects, project fields, field types |
| `mutation.go` | GraphQL mutations: add item to project, update field values (text/date/number/select/iteration) |
| `survey.go` | Interactive prompts using `survey/v2` for project/content/field selection |
| `fields.go` | Merges field type info with field options (single-select, iteration) |
| `types.go` | Shared struct definitions (Project, Content, Option, ProjectField, Repository) |

## Key Libraries

- `github.com/cli/go-gh` — gh CLI host library (REST client, GQL client, exec)
- `github.com/cli/shurcooL-graphql` — GraphQL query/mutation construction via Go struct tags
- `github.com/AlecAivazis/survey/v2` — interactive terminal prompts
- `github.com/shurcooL/githubv4` — GitHub v4 API types (used for `githubv4.Date`)

## Flow

1. Fetch user projects + org projects (if repo is in an org)
2. User selects a project interactively
3. Fetch project fields and their types/options
4. User selects content type (Current PR / PR / Issue) or uses `-issue`/`-pr` flags
5. Add the content to the project via `addProjectV2ItemById` mutation
6. For each custom field, use CLI `-field` value if provided, otherwise prompt user for input, and update via `updateProjectV2ItemFieldValue` mutation

## Release

Tags matching `v*` trigger `.github/workflows/release.yml` which uses `cli/gh-extension-precompile` to build and publish.
