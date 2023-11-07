#!/usr/bin/env bash

set -euo pipefail

pluginDir=".semrel/$(go env GOOS)_$(go env GOARCH)/changelog-generator-llm/0.0.0-dev/"
[[ ! -d "$pluginDir" ]] && {
  echo "creating $pluginDir"
  mkdir -p "$pluginDir"
}

go build -o "$pluginDir/changelog-generator-llm" ./
