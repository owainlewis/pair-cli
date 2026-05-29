#!/usr/bin/env bash
#
# pair-cli installer
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/owainlewis/pair-cli/main/install.sh | bash
#
# Environment overrides:
#   PAIR_VERSION   Tag to install (e.g. v1.2.3). Default: latest release.
#   PAIR_BIN_DIR   Install directory. Default: $HOME/.local/bin
#
set -euo pipefail

REPO="owainlewis/pair-cli"
BIN_NAME="pair"
BIN_DIR="${PAIR_BIN_DIR:-$HOME/.local/bin}"
VERSION="${PAIR_VERSION:-latest}"

info() { printf '\033[1;34m==>\033[0m %s\n' "$1"; }
warn() { printf '\033[1;33mwarning:\033[0m %s\n' "$1" >&2; }
err()  { printf '\033[1;31merror:\033[0m %s\n' "$1" >&2; exit 1; }

require() {
  command -v "$1" >/dev/null 2>&1 || err "'$1' is required but not installed."
}

# Sets globals OS and ARCH.
detect_platform() {
  local os arch
  os="$(uname -s)"
  arch="$(uname -m)"

  case "$os" in
    Darwin) os="darwin" ;;
    Linux)  os="linux" ;;
    *) err "unsupported operating system: $os" ;;
  esac

  case "$arch" in
    x86_64 | amd64) arch="amd64" ;;
    arm64 | aarch64) arch="arm64" ;;
    *) err "unsupported architecture: $arch" ;;
  esac

  # Guard against combos the release pipeline does not build.
  case "${os}-${arch}" in
    darwin-amd64 | darwin-arm64 | linux-amd64) ;;
    *) err "no prebuilt binary for ${os}/${arch}. Build from source with 'go install ./cmd/pair'." ;;
  esac

  OS="$os"
  ARCH="$arch"
}

# Prints the SHA-256 of a file, using whichever tool is available
# (sha256sum on Linux, shasum on macOS).
sha256_of() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print $1}'
  else
    err "need 'sha256sum' or 'shasum' to verify the download."
  fi
}

# Verifies $1 (downloaded file) against the release's SHA256SUMS for asset $2.
# Hard-fails on a mismatch or a missing entry. If the release has no
# SHA256SUMS published at all (e.g. older releases), warns and continues.
verify_checksum() {
  local file="$1" asset="$2"
  local sums_url="https://github.com/${REPO}/releases/download/${VERSION}/SHA256SUMS"

  if ! curl -fsSL "$sums_url" -o "${TMP}/SHA256SUMS" 2>/dev/null; then
    warn "no SHA256SUMS published for ${VERSION}; skipping checksum verification."
    return
  fi

  local expected actual
  expected="$(awk -v a="$asset" '$2 == a {print $1}' "${TMP}/SHA256SUMS")"
  [ -n "$expected" ] || err "no checksum for ${asset} in SHA256SUMS."

  actual="$(sha256_of "$file")"
  if [ "$expected" != "$actual" ]; then
    err "checksum mismatch for ${asset}: expected ${expected}, got ${actual}."
  fi
  info "Checksum verified."
}

# Sets global VERSION to a concrete tag.
resolve_version() {
  [ "$VERSION" != "latest" ] && return

  local api="https://api.github.com/repos/${REPO}/releases/latest"
  local tag
  tag="$(curl -fsSL "$api" | grep -m1 '"tag_name"' | sed -E 's/.*"tag_name" *: *"([^"]+)".*/\1/')"
  [ -n "$tag" ] || err "could not determine latest release tag from GitHub."
  VERSION="$tag"
}

main() {
  require curl
  require tar

  detect_platform
  resolve_version
  info "Installing ${BIN_NAME} ${VERSION} for ${OS}/${ARCH}"

  local pkg="${BIN_NAME}-${VERSION}-${OS}-${ARCH}"
  local asset="${pkg}.tar.gz"
  local url="https://github.com/${REPO}/releases/download/${VERSION}/${asset}"
  local inner_binary="${BIN_NAME}-${OS}-${ARCH}"

  # Global (not local) so the EXIT trap can still see it for cleanup.
  TMP="$(mktemp -d)"
  trap 'rm -rf "${TMP:-}"' EXIT

  info "Downloading ${url}"
  curl -fsSL "$url" -o "${TMP}/${asset}" \
    || err "download failed. Check that release ${VERSION} has asset ${asset}."

  verify_checksum "${TMP}/${asset}" "${asset}"

  tar -xzf "${TMP}/${asset}" -C "$TMP" \
    || err "failed to extract ${asset}."

  local extracted="${TMP}/${pkg}/${inner_binary}"
  [ -f "$extracted" ] || err "expected binary ${inner_binary} not found in archive."

  mkdir -p "$BIN_DIR"
  install -m 0755 "$extracted" "${BIN_DIR}/${BIN_NAME}"

  info "Installed to ${BIN_DIR}/${BIN_NAME}"

  case ":${PATH}:" in
    *":${BIN_DIR}:"*) ;;
    *)
      warn "${BIN_DIR} is not on your PATH."
      printf '  Add this to your shell profile:\n\n    export PATH="%s:$PATH"\n\n' "$BIN_DIR"
      ;;
  esac

  info "Done. Run '${BIN_NAME} --help' to get started."
}

main "$@"
