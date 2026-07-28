#!/usr/bin/env sh
# install.sh — build and install the linear CLI into a local bin dir.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/scmmishra/linear-cli/main/install.sh | sh
#   ./install.sh            # from a checkout: builds the local source
#
# Environment:
#   LINEAR_VERSION      pin a module version, e.g. "v0.1.0" (default: latest)
#   LINEAR_INSTALL_DIR  install location (default: $HOME/.local/bin)
#   NO_COLOR            set to disable color output
#
# Requires Go 1.25+ (builds from source; no prebuilt binaries yet).

set -eu

MODULE="github.com/scmmishra/linear-cli"
BINARY="linear"
INSTALL_DIR="${LINEAR_INSTALL_DIR:-$HOME/.local/bin}"
VERSION="${LINEAR_VERSION:-latest}"

if [ -t 1 ] && [ -z "${NO_COLOR:-}" ]; then
  C_RESET=$(printf '\033[0m')
  C_DIM=$(printf '\033[2m')
  C_BOLD=$(printf '\033[1m')
  C_CYAN=$(printf '\033[36m')
  C_GREEN=$(printf '\033[32m')
  C_YELLOW=$(printf '\033[33m')
  C_RED=$(printf '\033[31m')
else
  C_RESET=''; C_DIM=''; C_BOLD=''; C_CYAN=''; C_GREEN=''; C_YELLOW=''; C_RED=''
fi

step() { printf '  %s→%s %s\n' "$C_CYAN" "$C_RESET" "$*"; }
ok()   { printf '  %s✓%s %s\n' "$C_GREEN" "$C_RESET" "$*"; }
warn() { printf '  %s!%s %s\n' "$C_YELLOW" "$C_RESET" "$*"; }
err()  { printf '  %s✗%s %s\n' "$C_RED" "$C_RESET" "$*" >&2; exit 1; }
has()  { command -v "$1" >/dev/null 2>&1; }

has go || err "Go is required (https://go.dev/dl). No prebuilt binaries are published yet."

mkdir -p "$INSTALL_DIR"

# From a checkout (script sits next to go.mod for this module), build the
# local source; otherwise fetch the module through the Go proxy.
script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
if [ -f "$script_dir/go.mod" ] && grep -q "^module $MODULE$" "$script_dir/go.mod"; then
  step "Building from local checkout ${C_DIM}${script_dir}${C_RESET}"
  (cd "$script_dir" && go build -o "$INSTALL_DIR/$BINARY" ./cmd/linear/)
  ok "Installed ${C_BOLD}${BINARY}${C_RESET} (local build) to ${INSTALL_DIR}/${BINARY}"
else
  step "Installing ${C_BOLD}${MODULE}@${VERSION}${C_RESET}"
  GOBIN="$INSTALL_DIR" go install "${MODULE}/cmd/linear@${VERSION}"
  ok "Installed ${C_BOLD}${BINARY}@${VERSION}${C_RESET} to ${INSTALL_DIR}/${BINARY}"
fi

case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *)
    printf '\n'
    warn "${INSTALL_DIR} is not on your PATH"
    printf '    add this to your shell profile (~/.zshrc, ~/.bashrc, etc.):\n\n'
    # shellcheck disable=SC2016  # literal $PATH is intentional here
    printf '      %sexport PATH="%s:$PATH"%s\n\n' "$C_DIM" "$INSTALL_DIR" "$C_RESET"
    ;;
esac

printf '\n  %sNext steps%s\n' "$C_BOLD" "$C_RESET"
printf '    %slinear auth login%s          save your Linear API key\n' "$C_CYAN" "$C_RESET"
printf '    %slinear --help%s              see all commands\n\n' "$C_CYAN" "$C_RESET"
