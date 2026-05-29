#!/usr/bin/env bash
# install_radon.sh — installs the Radon Python complexity tool.
# Run this once before using pr-complexity outside Docker.
set -euo pipefail

MIN_PYTHON="3.8"

check_python() {
  if ! command -v python3 &>/dev/null; then
    echo "❌  python3 not found. Install Python ${MIN_PYTHON}+ first." >&2
    exit 1
  fi
  version=$(python3 -c 'import sys; print(f"{sys.version_info.major}.{sys.version_info.minor}")')
  echo "✓  Python ${version} found"
}

install_radon() {
  if command -v radon &>/dev/null; then
    echo "✓  radon already installed ($(radon --version))"
    return
  fi

  echo "→  Installing radon via pip..."
  pip3 install --quiet radon
  echo "✓  radon installed ($(radon --version))"
}

verify() {
  echo "→  Verifying radon works..."
  echo "def f(x): return x" | radon cc --json - >/dev/null
  echo "✓  radon verification passed"
}

check_python
install_radon
verify

echo ""
echo "All done. You can now run:"
echo "  pr-complexity analyze --base main --head HEAD"
