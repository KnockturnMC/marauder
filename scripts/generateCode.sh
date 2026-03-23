#!/usr/bin/env bash
set -euo pipefail

if [ ! $(command -v protoc) ]; then
  echo -e "\033[0;31mCannot find protoc executable required for generating protobuf definitions!"
  echo -e "\033[0;31mPlease install it using any of the instructions found here:"
  echo -e "  https://protobuf.dev/downloads/"
  exit 1
fi

go list -f '{{.Dir}}' -m | while read module; do
  pushd "$module" >/dev/null

  go run github.com/vektra/mockery/v3@v3.7.0

  popd >/dev/null
done

protoc -I "marauder-proto/src/main" --go_out "marauder-proto/src/main/golang" "marauder-proto/src/main/proto/servers.proto"
