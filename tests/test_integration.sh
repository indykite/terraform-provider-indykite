#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
TESTS_PROVIDER_DIR="${REPO_ROOT}/tests/provider"
TESTS_TERRAFORM_DIR="${REPO_ROOT}/tests/terraform"

cd "${REPO_ROOT}"

TERRAFORM_BIN="${TERRAFORM_BIN:-terraform}"

if ! command -v "${TERRAFORM_BIN}" >/dev/null 2>&1; then
    echo "Unable to resolve terraform binary '${TERRAFORM_BIN}' for integration tests"
    exit 1
fi

cleanup() {
    if [[ -f "${TESTS_PROVIDER_DIR}/test.tf.bak" ]]; then
        mv -f "${TESTS_PROVIDER_DIR}/test.tf.bak" "${TESTS_PROVIDER_DIR}/test.tf"
    fi
    rm -f "${TESTS_PROVIDER_DIR}/terraform.tfstate" "${TESTS_PROVIDER_DIR}/terraform.tfstate.backup"
}

trap cleanup EXIT

"${TERRAFORM_BIN}" -chdir="${TESTS_PROVIDER_DIR}" plan
"${TERRAFORM_BIN}" -chdir="${TESTS_PROVIDER_DIR}" apply -input=false -auto-approve

cd "${TESTS_TERRAFORM_DIR}"
go test --tags=integration ./...

cd "${TESTS_PROVIDER_DIR}"
cp test.tf test.tf.bak
sed -i \
    -e 's/description\(\s*=\s*"\)/description\1Updated - /g' \
    -e 's/provider_display_name\(\s*=\s*"\)/provider_display_name\1Updated - /g' \
    -e 's/route_display_name\(\s*=\s*"\)/route_display_name\1Updated - /g' \
    test.tf
"${TERRAFORM_BIN}" apply -input=false -auto-approve
"${TERRAFORM_BIN}" destroy -input=false -auto-approve
