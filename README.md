# pair-cli
A CLI tool for managing your agent workspace.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/owainlewis/pair-cli/main/install.sh | bash
```

This downloads the latest release binary for your platform and installs it to
`~/.local/bin/pair`. Make sure that directory is on your `PATH`.

Overrides:

- `PAIR_VERSION=v1.2.3` — install a specific tag instead of the latest release.
- `PAIR_BIN_DIR=/usr/local/bin` — install somewhere else (may need `sudo`).

Prefer to inspect before running? Download it first:

```sh
curl -fsSL https://raw.githubusercontent.com/owainlewis/pair-cli/main/install.sh -o install.sh
less install.sh
bash install.sh
```
