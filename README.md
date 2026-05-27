# pair-cli

`pair` is a Go CLI for operating a PAIR agent workspace from a terminal or automation environment.

## Install

Build a local binary:

```sh
make build
./pair --help
```

Or run without installing:

```sh
go run ./cmd/pair --help
```

Cross-platform release builds are available through:

```sh
make cross-build
```

That produces binaries in `dist/` for:

- `darwin/arm64`
- `darwin/amd64`
- `linux/arm64`
- `linux/amd64`

## Configure

The CLI resolves configuration in this order:

1. Flags: `--base-url`, `--token`
2. Environment variables: `PAIR_BASE_URL`, `PAIR_TOKEN`
3. Config file: `$XDG_CONFIG_HOME/pair/config.json` or `~/.config/pair/config.json`

Set config values:

```sh
pair config set base-url http://localhost:3000
pair config set token pair_your_token
pair auth status
```

Environment variables are convenient for CI and agents:

```sh
export PAIR_BASE_URL=http://localhost:3000
export PAIR_TOKEN=pair_your_token
```

Tokens are never printed by normal command output.

## Usage

Inspect tasks:

```sh
pair tasks list
pair tasks show task_123 --json
```

Claim a task, publish work, and move it to review:

```sh
pair tasks status task_123 doing
pair tasks publish task_123 --file notes.md --tag demo
pair tasks comment task_123 --body "Published draft notes."
pair tasks status task_123 review
```

Work with documents:

```sh
pair docs list --tag demo
pair docs create --file notes.md --tag demo
pair docs read doc_123 --output notes.md
pair docs replace doc_123 --file notes.md
```

Work with collections:

```sh
pair collections create --name "Demo Bundle"
pair collections link-doc col_123 doc_123
pair collections publish col_123 --file summary.md --tag demo
```

Destructive commands require explicit confirmation in non-interactive environments:

```sh
pair tasks delete task_123 --yes
pair docs delete doc_123 --yes
pair collections delete col_123 --yes
```

## Development

Run tests:

```sh
make test
```

Build target platforms manually:

```sh
GOOS=darwin GOARCH=arm64 go build ./cmd/pair
GOOS=linux GOARCH=amd64 go build ./cmd/pair
```
