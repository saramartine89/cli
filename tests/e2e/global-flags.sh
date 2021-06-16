#!/bin/bash

set -euo pipefail

TEST_NAME="global flags"

cleanup() {
  exit_code=$?
  if [ "$exit_code" -ne 0 ]; then
    echo "ERROR: '$TEST_NAME' tests failed during execution"
    afterAll || echo "ERROR: Cleanup failed"
  fi

  exit "$exit_code"
}
trap cleanup EXIT

beforeAll() {
  echo "INFO: Executing '$TEST_NAME' tests"
}

beforeEach() {
  rm -f ./temp-config
}

afterAll() {
  echo "INFO: Completed '$TEST_NAME' tests"
  beforeEach
}

error() {
  message=$1
  echo "$message"
  exit 1
}

beforeAll

beforeEach

###
# --no-read-env flag
###

# verify initial state
token="$("$DOPPLER_BINARY" configure debug --json --configuration=./temp-config | jq -r ".[\"/\"].token")"
[[ "$token" == "" ]] || error "ERROR: expected blank config value"
# verify env var is read
token="$(DOPPLER_TOKEN=123 "$DOPPLER_BINARY" configure debug --json --configuration=./temp-config | jq -r ".[\"/\"].token")"
[[ "$token" == "123" ]] || error "ERROR: expected token from environment"
# verify env var is ignored
token="$(DOPPLER_TOKEN=123 "$DOPPLER_BINARY" configure debug --json --configuration=./temp-config --no-read-env | jq -r ".[\"/\"].token")"
[[ "$token" == "" ]] || error "ERROR: expected blank config value"

afterAll
