#!/usr/bin/env bash
set -euo pipefail

if [ ! $(command -v ginkgo) ]; then
  echo -e "\033[0;31mCannot find ginkgo executable required for testing!"
  echo -e "\033[0;31mPlease install it using:"
  echo -e "  go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo"
  exit 1
fi

FLAGS=(
  -r
  -p
  --randomize-all
  --randomize-suites
  -fail-on-pending
  --poll-progress-after=30s
  -nodes=1
  -compilers=4
  -race
  -trace
  --label-filter="functiontest"
)

if [ ! -z ${COVERAGE+x} ]; then
  FLAGS+=("--cover")
fi

GO111MODULE=on ginkgo "${FLAGS[@]}" ./...
