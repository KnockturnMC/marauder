#!/usr/bin/env bash
set -euo pipefail

if [ ! $(command -v gofumpt) ]; then
  echo -e "\033[0;31mCannot find gofumpt executable required for formatting!"
  echo -e "\033[0;31mPlease install it using:"
  echo -e "  go install mvdan.cc/gofumpt@latest"
  exit 1
fi

if [ ! $(command -v shfmt) ]; then
  echo -e "\033[0;31mCannot find shfmt executable required for formatting!"
  echo -e "\033[0;31mPlease install it using:"
  echo -e "  go install mvdan.cc/sh/v3/cmd/shfmt@latest"
  exit 1
fi

gofumpt -w -l .
find . -name "*.sh" -exec shfmt -i 2 -l -w {} \;
