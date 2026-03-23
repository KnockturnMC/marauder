#!/usr/bin/env bash
set -euo pipefail

go run mvdan.cc/gofumpt@latest -w -l .
find . -name "*.sh" -exec go run mvdan.cc/sh/v3/cmd/shfmt@latest -i 2 -l -w {} \;
