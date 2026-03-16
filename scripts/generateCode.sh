#!/usr/bin/env bash
set -euo pipefail

if [ ! $(command -v mockery) ]; then
  echo -e "\033[0;31mCannot find mockery executable required for generating mocks!"
  echo -e "\033[0;31mPlease install it using any of the instructions found here:"
  echo -e "  https://vektra.github.io/mockery/latest/installation/"
  exit 1
fi

if [ ! $(command -v protoc) ]; then
  echo -e "\033[0;31mCannot find protoc executable required for generating protobuf definitions!"
  echo -e "\033[0;31mPlease install it using any of the instructions found here:"
  echo -e "  https://protobuf.dev/downloads/"
  exit 1
fi

go list -f '{{.Dir}}' -m | while read module; do
  pushd "$module" >/dev/null

  mockery

  popd >/dev/null
done

protoc -I "marauder-proto/src" --go_out "marauder-proto/pkg" "marauder-proto/src/servers.proto"
