#!/bin/bash

set -euo pipefail

TEST_NAME="run-mount"

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
  "$DOPPLER_BINARY" run clean --max-age=0s --silent
}

beforeEach() {
  "$DOPPLER_BINARY" run clean --max-age=0s --silent
  rm -f ./secrets.json ./secrets.env
}

afterAll() {
  echo "INFO: Completed '$TEST_NAME' tests"
  "$DOPPLER_BINARY" run clean --max-age=0s --silent
  rm -f ./secrets.json ./secrets.env
}

beforeAll

beforeEach

# verify content of mounted secrets file
EXPECTED_SECRETS='{"DOPPLER_CONFIG":"prd_e2e_tests","DOPPLER_ENCLAVE_CONFIG":"prd_e2e_tests","DOPPLER_ENCLAVE_ENVIRONMENT":"prd","DOPPLER_ENCLAVE_PROJECT":"cli","DOPPLER_ENVIRONMENT":"prd","DOPPLER_PROJECT":"cli"}'
actual="$("$DOPPLER_BINARY" run --mount secrets.json --command "cat \$DOPPLER_CLI_SECRETS_PATH")"
if [ "$actual" != "$EXPECTED_SECRETS" ]; then
 echo "ERROR: mounted secrets file has invalid contents"
 exit 1
fi

beforeEach

# verify secrets aren't injected into environment
# this will succeed
"$DOPPLER_BINARY" run --command "printenv DOPPLER_ENVIRONMENT" && \
  # this should fail
  "$DOPPLER_BINARY" run --mount secrets.json --command "printenv DOPPLER_ENVIRONMENT" && \
  (echo "ERROR: secrets injected into environment despite using mounted secrets file" && exit 1)

beforeEach

# verify specified mount path is used and converted to absolute path
expected="$(pwd)/secrets.json"
actual="$("$DOPPLER_BINARY" run --mount secrets.json --command "echo \$DOPPLER_CLI_SECRETS_PATH")"
if [ "$actual" != "$expected" ]; then
 echo "ERROR: secrets are not mounted to specified path"
 exit 1
fi

beforeEach

# verify format is auto detected
EXPECTED_SECRETS='DOPPLER_CONFIG="prd_e2e_tests"\nDOPPLER_ENCLAVE_CONFIG="prd_e2e_tests"\nDOPPLER_ENCLAVE_ENVIRONMENT="prd"\nDOPPLER_ENCLAVE_PROJECT="cli"\nDOPPLER_ENVIRONMENT="prd"\nDOPPLER_PROJECT="cli"'
actual="$("$DOPPLER_BINARY" run --mount secrets.env --command "cat \$DOPPLER_CLI_SECRETS_PATH")"
if [[ "$actual" != "$(echo -e "$EXPECTED_SECRETS")" ]]; then
 echo "ERROR: mounted secrets file with auto-detected env format has invalid contents"
 exit 1
fi

beforeEach

# verify specified format is used
EXPECTED_SECRETS='{"DOPPLER_CONFIG":"prd_e2e_tests","DOPPLER_ENCLAVE_CONFIG":"prd_e2e_tests","DOPPLER_ENCLAVE_ENVIRONMENT":"prd","DOPPLER_ENCLAVE_PROJECT":"cli","DOPPLER_ENVIRONMENT":"prd","DOPPLER_PROJECT":"cli"}'
actual="$("$DOPPLER_BINARY" run --mount secrets.env --mount-format json --command "cat \$DOPPLER_CLI_SECRETS_PATH")"
if [[ "$actual" != "$(echo -e "$EXPECTED_SECRETS")" ]]; then
 echo "ERROR: mounted secrets file with json format has invalid contents"
 exit 1
fi

beforeEach

# verify specified name transformer is used
EXPECTED_SECRETS='{"TF_VAR_doppler_config":"prd_e2e_tests","TF_VAR_doppler_enclave_config":"prd_e2e_tests","TF_VAR_doppler_enclave_environment":"prd","TF_VAR_doppler_enclave_project":"cli","TF_VAR_doppler_environment":"prd","TF_VAR_doppler_project":"cli"}'
actual="$("$DOPPLER_BINARY" run --mount secrets.json --name-transformer tf-var --command "cat \$DOPPLER_CLI_SECRETS_PATH")"
if [[ "$actual" != "$EXPECTED_SECRETS" ]]; then
 echo "ERROR: mounted secrets file with name transformer has invalid contents"
 exit 1
fi

afterAll
