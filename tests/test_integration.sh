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

APPLIED=0

cleanup() {
    local rc=$?
    trap - EXIT
    if [[ "${APPLIED}" -eq 1 ]]; then
        if [[ -f "${TESTS_PROVIDER_DIR}/test.tf.bak" ]]; then
            mv -f "${TESTS_PROVIDER_DIR}/test.tf.bak" "${TESTS_PROVIDER_DIR}/test.tf" ||
                echo ">>> Cleanup: failed to restore test.tf from backup" >&2
        fi
        echo ">>> Cleanup: running terraform destroy (main rc=${rc})"
        if ! "${TERRAFORM_BIN}" -chdir="${TESTS_PROVIDER_DIR}" destroy -input=false -auto-approve; then
            echo ">>> Cleanup: terraform destroy failed" >&2
            [[ "${rc}" -eq 0 ]] && rc=1
        fi
        # Non-fatal so a filesystem error cannot overwrite the real exit code.
        rm -f "${TESTS_PROVIDER_DIR}/terraform.tfstate" "${TESTS_PROVIDER_DIR}/terraform.tfstate.backup" ||
            echo ">>> Cleanup: failed to remove local tfstate files" >&2
    fi
    exit "${rc}"
}

trap cleanup EXIT

"${TERRAFORM_BIN}" -chdir="${TESTS_PROVIDER_DIR}" plan
APPLIED=1
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
# terraform destroy runs automatically via the EXIT trap above
