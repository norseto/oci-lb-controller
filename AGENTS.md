# AGENTS.md

## General
- Respond in Japanese
- Do only what is instructed.
- Keep responses concise.
- If you are unsure about something, ask for more context.
- DO NOT ASSUME YOU KNOW EVERYTHING, ASK THE USER ABOUT THEIR REASONING.
- Ask users to provide more context (for example imported files etc) when needed.

## Basic Code Style & Guidelines
- Write code comments and PR comments in English.
- Make comments on lines by themselves.
- Format code with `go fmt ./...` or `make fmt`

## Toolchain Versioning
- Keep the Go major.minor series aligned across `go.mod`, the root `Dockerfile`, and `.devcontainer/devcontainer.json`.
- Use a major.minor version for the devcontainer Go Feature (for example, `1.26`); specify the required patch version in the `go.mod` toolchain and the root `Dockerfile` Go builder (for example, `1.26.9`). Update the Feature only when changing series.
- With `GOTOOLCHAIN=auto`, the toolchain directive suggests the project Go version; a newer installed toolchain may be selected. It does not strictly pin the version or guarantee the same binary.

## Git Usage
- Use only English for branch name
- Use only English for commit messages

## Testing
- Run `make vet` before every commit
- Run `make test` on every commit
- Run `make lint` on every commit
- Run `make vulcheck` on every commit
- Run `make seccheck` on every commit
- Coverrage: 80%
- Do not create PR if test fails

### Test Coverage
- Generate coverage report with `go test ./... -coverprofile=coverage.out`
- Inspect coverage with `go tool cover -func coverage.out`
- Enforce a minimum overall coverage of 80%
- CI should fail if coverage falls below threshold
- Ensure coverage includes critical paths and edge cases

## Commit Message
- Follow Conventional Commits
- Use only English for commit messages
