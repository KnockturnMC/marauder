#!/usr/bin/env bash
set -euo pipefail

if [ ! $(command -v golangci-lint) ]; then
  echo -e "\033[0;31mCannot find golangci-lint executable required for linting!"
  echo -e "\033[0;31mPlease install it using any of the instructions found here:"
  echo -e "  https://golangci-lint.run/usage/install/#local-installation"
  exit 1
fi

if [ ! $(command -v staticcheck) ]; then
  echo -e "\033[0;31mCannot find staticcheck executable required for linting!"
  echo -e "\033[0;31mPlease install it using any of the instructions found here:"
  echo -e "  https://staticcheck.io/docs/getting-started/#installation"
  exit 1
fi

go list -f '{{.Dir}}' -m | while read module; do
  pushd "$module" >/dev/null

  golangci-lint --timeout 3m0s run
  staticcheck ./...

  popd >/dev/null
done
