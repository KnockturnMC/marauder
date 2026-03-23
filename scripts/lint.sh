#!/usr/bin/env bash
set -euo pipefail

if [ ! $(command -v golangci-lint) ]; then
  echo -e "\033[0;31mCannot find golangci-lint executable required for linting!"
  echo -e "\033[0;31mPlease install it using any of the instructions found here:"
  echo -e "  https://golangci-lint.run/docs/welcome/install/local/"
  exit 1
fi

go list -f '{{.Dir}}' -m | while read module; do
  pushd "$module" >/dev/null

  echo "linting ${module}..."
  golangci-lint --timeout 3m0s run
  go run honnef.co/go/tools/cmd/staticcheck@latest ./...

  popd >/dev/null
done
