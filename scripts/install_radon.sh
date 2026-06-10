#!/usr/bin/env bash
# install_radon.sh — installs the Radon Python complexity tool.
# Run this once before using pr-complexity outside Docker for Python analysis.
set -euo pipefail

MIN_PYTHON="3.8"

check_python() {
  if ! command -v python3 &>/dev/null; then
    echo "❌  python3 not found. Install Python ${MIN_PYTHON}+ first." >&2
    exit 1
  fi

  version=$(python3 -c 'import sys; print(f"{sys.version_info.major}.{sys.version_info.minor}")')
  if ! python3 - <<PY
import sys
minimum = tuple(map(int, "${MIN_PYTHON}".split(".")))
current = sys.version_info[:2]
raise SystemExit(0 if current >= minimum else 1)
PY
  then
    echo "❌  Python ${version} found, but Python ${MIN_PYTHON}+ is required." >&2
    exit 1
  fi

  echo "✓  Python ${version} found"
}

install_radon() {
  if command -v radon &>/dev/null; then
    echo "✓  radon already installed ($(radon --version))"
    return
  fi

  echo "→  Installing radon via pip..."
  python3 -m pip install --quiet 'radon==6.*'
  echo "✓  radon installed ($(radon --version))"
}

verify() {
  echo "→  Verifying radon works..."
  tmp=$(mktemp /tmp/pr-complexity-radon-XXXXXX.py)
  trap 'rm -f "$tmp"' EXIT
  printf 'def f(x):\n    return x\n' > "$tmp"
  radon cc --json "$tmp" >/dev/null
  echo "✓  radon verification passed"
}

check_python
install_radon
verify

echo ""
echo "All done. You can now run:"
echo "  pr-complexity analyze --base main --head HEAD"
