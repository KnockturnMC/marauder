#!/usr/bin/env bash
set -euo pipefail

FLAGS=(
  -r
  -p
  --randomize-all
  --randomize-suites
  -fail-on-pending
  --poll-progress-after=30s
  -nodes=4
  -compilers=4
  -race
  -trace
  --label-filter="unittest"
)

if [ ! -z ${COVERAGE+x} ]; then
  FLAGS+=("--cover")
fi

GO111MODULE=on go run github.com/onsi/ginkgo/v2/ginkgo "${FLAGS[@]}" ./...
