#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${REPO_ROOT}"

GO_CPU_TEST="${GO_CPU_TEST:-$(getconf _NPROCESSORS_ONLN)}"
HASHICORP_PREFIX="$(brew --prefix hashicorp/tap/terraform 2>/dev/null || true)"
TERRAFORM_PATH="${TF_ACC_TERRAFORM_PATH:-}"

if [[ -z "${TERRAFORM_PATH}" && -n "${HASHICORP_PREFIX}" && -x "${HASHICORP_PREFIX}/bin/terraform" ]]; then
    TERRAFORM_PATH="${HASHICORP_PREFIX}/bin/terraform"
fi

if [[ -z "${TERRAFORM_PATH}" ]]; then
    TERRAFORM_PATH="$(command -v "${TERRAFORM_BIN:-terraform}" 2>/dev/null || true)"
fi

if [[ -z "${TERRAFORM_PATH}" ]]; then
    TERRAFORM_PATH="$(command -v terraform 2>/dev/null || true)"
fi

if [[ -z "${TERRAFORM_PATH}" ]]; then
    TERRAFORM_PATH="$(command -v tofu 2>/dev/null || true)"
fi

if [[ -z "${TERRAFORM_PATH}" ]]; then
    echo "Unable to resolve terraform/tofu binary for TF_ACC_TERRAFORM_PATH"
    exit 1
fi

echo "==> Using Terraform binary at: ${TERRAFORM_PATH}"
"${TERRAFORM_PATH}" version

if "${TERRAFORM_PATH}" version | grep -q "OpenTofu"; then
    echo "Acceptance tests require HashiCorp Terraform CLI, but OpenTofu was detected at ${TERRAFORM_PATH}. Set TF_ACC_TERRAFORM_PATH to a HashiCorp Terraform binary path."
    exit 1
fi

TF_ACC_TERRAFORM_PATH="${TERRAFORM_PATH}" go test \
    -v \
    -cpu "${GO_CPU_TEST}" \
    -covermode=count \
    -coverpkg github.com/indykite/terraform-provider-indykite/... \
    -coverprofile=coverage.out \
    ./...
