# Agent Workspace CLI Plan

Source spec: `docs/agent-workspace-cli/spec.md`

Shared decisions for all tasks:

- The executable name is `pair`.
- Commands must be usable non-interactively by agents.
- The CLI talks to the existing PAIR Rails API under `/api/v1`.
- Default output is human-readable; `--json` is the stable machine interface.
- Tokens must never appear in normal output, errors, or logs.

## Task 1: Bootstrap the Go CLI Project

### Goal

Create the initial Go module, binary entrypoint, and Cobra command skeleton for `pair`.

### Context

The repository currently has only a README, `.gitignore`, and planning docs. The spec calls for a single Go binary with a conventional package layout.

### Relevant files or references

- `docs/agent-workspace-cli/spec.md`
- `cmd/pair/main.go`
- `internal/cli`
- `go.mod`

### Proposed approach

Initialize the module, add Cobra, create `cmd/pair/main.go`, and implement a root command with global `--base-url`, `--token`, and `--json` flags. Add placeholder subcommand groups for `auth`, `config`, `tasks`, `docs`, and `collections` with help text, but avoid implementing API behavior in this task.

### Acceptance criteria

- `go run ./cmd/pair --help` prints help for the root command.
- `go run ./cmd/pair tasks --help`, `docs --help`, and `collections --help` work.
- The binary builds with `go build ./cmd/pair`.
- The module path matches the GitHub repository.
- No command prints sensitive values.

### Source reference

Spec sections: What, CLI shape, Package layout, Versions.

### Verify

```sh
go test ./...
go build ./cmd/pair
go run ./cmd/pair --help
```

### Out of scope

- Real config persistence.
- HTTP client implementation.
- Resource command behavior.

## Task 2: Implement Config Resolution and Auth Status

### Goal

Add configuration loading, saving, and `pair auth status`.

### Context

The CLI needs to work for agents in local shells and automation. Configuration must resolve from flags, environment variables, and an XDG config file, with predictable precedence.

### Relevant files or references

- `docs/agent-workspace-cli/spec.md`
- `internal/config`
- `internal/cli`

### Proposed approach

Implement a config package that resolves `base_url` and `token` from flags, `PAIR_BASE_URL`, `PAIR_TOKEN`, and `$XDG_CONFIG_HOME/pair/config.json` or `~/.config/pair/config.json`. Add `pair config set base-url`, `pair config set token`, and `pair auth status`. Store config as JSON and write token-bearing files with mode `0600` on Unix-like systems.

### Acceptance criteria

- Flags override env vars, and env vars override config file values.
- `pair config set base-url <url>` persists the base URL.
- `pair config set token <token>` persists the token without echoing it back.
- `pair auth status` reports whether base URL and token are configured, without revealing the token.
- Tests cover config precedence and file write behavior.

### Source reference

Spec sections: Configuration, Error Behavior, Invariants.

### Verify

```sh
go test ./...
PAIR_BASE_URL=http://localhost:3000 PAIR_TOKEN=pair_test go run ./cmd/pair auth status
```

### Out of scope

- Calling the PAIR API to validate the token.
- OS keychain support.

## Task 3: Build the HTTP API Client Foundation

### Goal

Create the reusable API client, request helpers, response decoding, and safe error handling.

### Context

Every resource command depends on a consistent HTTP layer that attaches bearer auth, serializes JSON, handles markdown endpoints, and preserves API error details.

### Relevant files or references

- `docs/agent-workspace-cli/spec.md`
- `pair/app/controllers/api/v1/base_controller.rb`
- `pair/app/controllers/api/v1/contents_controller.rb`
- `internal/api`

### Proposed approach

Implement `api.Client` with `BaseURL`, `Token`, and an injectable `HTTPClient`. Add helpers for JSON requests, raw markdown requests, URL path escaping, and error decoding. Define `APIError` with status code, optional machine code, message, and sanitized request context.

### Acceptance criteria

- Requests include `Authorization: Bearer <token>`.
- JSON requests send and accept JSON.
- Markdown updates send `Content-Type: text/markdown; charset=utf-8`.
- 204 responses are treated as success.
- JSON API errors decode `error` and `message`.
- Text API errors decode readable content endpoint errors.
- Tokens are redacted from all error strings.

### Source reference

Spec sections: HTTP client, API mapping, Error Behavior, Invariants.

### Verify

```sh
go test ./...
```

### Out of scope

- Resource-specific methods beyond small fixtures needed for tests.
- Retry or backoff behavior.

## Task 4: Add Output, Input, and Confirmation Utilities

### Goal

Implement reusable terminal IO behavior for tables, JSON, stdin/file markdown, and destructive confirmations.

### Context

The resource commands should share consistent behavior. Markdown commands must preserve bytes, `--json` must stay machine-safe, and destructive operations must not happen accidentally.

### Relevant files or references

- `docs/agent-workspace-cli/spec.md`
- `internal/output`
- `internal/cli`

### Proposed approach

Add output helpers for table rendering and JSON rendering. Add input helpers for `--body`, `--file <path>`, and `--file -`. Add confirmation helpers that require `--yes` for non-interactive destructive commands and prompt interactively only when stdin is a terminal.

### Acceptance criteria

- `--json` writes valid JSON to stdout only.
- Errors are written to stderr.
- `--file -` reads stdin without trimming or line ending normalization.
- `--body` and `--file` are mutually exclusive where both are supported.
- Delete/unlink helpers refuse to proceed non-interactively unless `--yes` is present.
- Tests cover markdown byte preservation and confirmation behavior.

### Source reference

Spec sections: Output, Requirements, Invariants, Error Behavior.

### Verify

```sh
go test ./...
```

### Out of scope

- Command-specific table design beyond examples required by tests.

## Task 5: Implement Document API Methods and Commands

### Goal

Add `pair docs` behavior for listing, searching, reading, creating, replacing, showing, and deleting documents.

### Context

Documents are the core PAIR primitive. Metadata is JSON, but content is raw markdown and must be preserved by `docs read` and `docs replace`.

### Relevant files or references

- `docs/agent-workspace-cli/spec.md`
- `pair/app/controllers/api/v1/documents_controller.rb`
- `pair/app/controllers/api/v1/contents_controller.rb`
- `pair/test/integration/api/documents_api_test.rb`
- `pair/test/integration/api/content_api_test.rb`

### Proposed approach

Define document types and API methods for list, show, create, read content, replace content, and delete. Wire these to `pair docs list`, `show`, `read`, `create`, `replace`, and `delete`. Support `--query`, `--tag`, `--since`, `--tag` repeat flags, `--output`, `--body`, `--file`, `--json`, and `--yes` where specified.

### Acceptance criteria

- `docs list` supports query, tag, and since filters.
- `docs read <id>` writes raw markdown to stdout by default and to a file with `--output`.
- `docs create` accepts markdown from body, file, or stdin and returns document metadata.
- `docs replace` replaces content byte-for-byte with the provided markdown body.
- `docs delete` requires confirmation or `--yes`.
- JSON and table output are both covered by tests.

### Source reference

Spec sections: CLI shape, API mapping, Requirements.

### Verify

```sh
go test ./...
go run ./cmd/pair docs --help
```

### Out of scope

- Public share toggling.
- Local document caching.

## Task 6: Implement Task List, Show, Create, Status, and Delete

### Goal

Add the foundational `pair tasks` commands for managing task records.

### Context

PAIR tasks are the board items agents operate. The CLI should make it easy to inspect work, create work, move work through statuses, and delete tasks when explicitly requested.

### Relevant files or references

- `docs/agent-workspace-cli/spec.md`
- `pair/app/controllers/api/v1/tasks_controller.rb`
- `pair/app/models/task.rb`
- `pair/test/integration/api/tasks_api_test.rb`

### Proposed approach

Define task types and API methods for list, show, create, update, and delete. Wire them to `pair tasks list`, `show`, `create`, `status`, and `delete`. Validate statuses locally against `todo`, `doing`, `review`, and `done`.

### Acceptance criteria

- `tasks list` supports optional `--status`.
- `tasks show` includes comments and linked documents when returned by the API.
- `tasks create` supports title, description, and status.
- `tasks status` rejects invalid statuses before making an HTTP request.
- `tasks delete` requires confirmation or `--yes`.
- Human-readable output keeps IDs visible and copyable.

### Source reference

Spec sections: CLI shape, API mapping, Requirements, Decisions.

### Verify

```sh
go test ./...
go run ./cmd/pair tasks --help
```

### Out of scope

- Task comments.
- Task document linking and publishing.

## Task 7: Implement Task Comments and Task Document Publishing

### Goal

Add agent collaboration commands for commenting on tasks and attaching produced documents.

### Context

The PAIR API has explicit agent paths for task comments and task-produced resources. These are central to the "agent workspace" workflow.

### Relevant files or references

- `docs/agent-workspace-cli/spec.md`
- `pair/app/controllers/api/v1/task_comments_controller.rb`
- `pair/app/controllers/api/v1/task_documents_controller.rb`
- `pair/test/integration/api/tasks_api_test.rb`

### Proposed approach

Add API methods and commands for `tasks comment`, `tasks link-doc`, `tasks publish`, and `tasks unlink-doc`. Comments should accept `--body` or `--file`. Publishing should create a new document from markdown and link it to the task in one API call. Linking should attach an existing document by ID.

### Acceptance criteria

- `tasks comment <id>` creates an agent comment with body from flag, file, or stdin.
- Blank comments fail locally before making an API request.
- `tasks link-doc <task-id> <document-id>` attaches an existing document.
- `tasks publish <task-id>` creates and links a document from markdown.
- `tasks publish` supports repeatable `--tag`.
- `tasks unlink-doc` requires confirmation or `--yes`.

### Source reference

Spec sections: CLI shape, API mapping, Requirements.

### Verify

```sh
go test ./...
```

### Out of scope

- Creating public share links for produced documents.

## Task 8: Implement Collection Commands

### Goal

Add collection list, show, create, rename, link, publish, unlink, and delete commands.

### Context

PAIR collections are ordered bundles of documents. Agents need to publish multi-artifact work into collections and attach existing docs when assembling a workspace.

### Relevant files or references

- `docs/agent-workspace-cli/spec.md`
- `pair/app/controllers/api/v1/collections_controller.rb`
- `pair/app/controllers/api/v1/collection_documents_controller.rb`
- `pair/test/integration/api/collections_api_test.rb`
- `pair/test/integration/api/collection_documents_api_test.rb`

### Proposed approach

Define collection types and API methods for collection metadata and document membership. Wire them to `pair collections list`, `show`, `create`, `rename`, `link-doc`, `publish`, `unlink-doc`, and `delete`. Reuse markdown input and confirmation helpers from earlier tasks.

### Acceptance criteria

- `collections list` shows collection IDs, names, document counts, and update times.
- `collections show` includes member documents when returned by the API.
- `collections create` and `rename` persist names.
- `collections link-doc` attaches existing documents.
- `collections publish` creates a markdown document and attaches it in one call.
- `collections unlink-doc` and `delete` require confirmation or `--yes`.
- JSON output is available for all commands.

### Source reference

Spec sections: CLI shape, API mapping, Requirements.

### Verify

```sh
go test ./...
go run ./cmd/pair collections --help
```

### Out of scope

- Collection sharing or bundle export endpoints that do not exist in the Rails API yet.

## Task 9: Add End-to-End CLI Tests Against Fake PAIR API

### Goal

Cover representative complete command flows using a local fake HTTP server.

### Context

Unit tests prove individual helpers, but the CLI also needs confidence that config, command parsing, HTTP requests, and output wiring work together.

### Relevant files or references

- `docs/agent-workspace-cli/spec.md`
- `internal/cli`
- `internal/api`

### Proposed approach

Use `httptest.Server` to simulate the Rails API and run CLI commands in-process with test stdout/stderr. Cover one happy path per resource plus key failure paths: missing token, API JSON error, API text error, invalid status, and destructive command without `--yes`.

### Acceptance criteria

- Tests cover at least one document create/read/replace flow.
- Tests cover at least one task status/comment/publish flow.
- Tests cover at least one collection publish flow.
- Tests assert auth headers are present without exposing token values in failures.
- Failure tests assert exit-code classification behavior.

### Source reference

Spec sections: Testing Strategy, Error Behavior.

### Verify

```sh
go test ./...
```

### Out of scope

- Tests that require a running Rails server.

## Task 10: Add Release Builds, Documentation, and Install Notes

### Goal

Make the CLI straightforward to build, install, and verify across target platforms.

### Context

The spec requires a single binary for macOS and Linux on ARM64 and AMD64. The README should teach humans and agents how to configure and use it.

### Relevant files or references

- `docs/agent-workspace-cli/spec.md`
- `README.md`
- `.github/workflows`

### Proposed approach

Add a build script or Makefile target for local builds, a GitHub Actions workflow that runs tests and cross-compiles release artifacts, and README usage docs covering install, config, auth, and representative task/document/collection workflows.

### Acceptance criteria

- `go test ./...` runs in CI.
- CI builds `pair` for `darwin/arm64`, `darwin/amd64`, `linux/arm64`, and `linux/amd64`.
- README documents environment variables, config file behavior, and basic commands.
- README includes examples for an agent claiming a task, publishing a document, and moving the task to review.
- Local build instructions produce a runnable `pair` binary.

### Source reference

Spec sections: Requirements, Versions, Testing Strategy.

### Verify

```sh
go test ./...
GOOS=darwin GOARCH=arm64 go build ./cmd/pair
GOOS=linux GOARCH=amd64 go build ./cmd/pair
```

### Out of scope

- Homebrew formula.
- Signed releases or package manager distribution.
