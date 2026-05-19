#!/bin/bash
# getting the latest release tag so the released terraform is used
# tag=$(git ls-remote --tags --exit-code --refs https://github.com/indykite/terraform-provider-indykite | sed -E 's/^[[:xdigit:]]+[[:space:]]+refs\/tags\/(.+)/\1/g' | sort --version-sort | tail -n1)

set -o errexit -o nounset -o pipefail

# default values and check for the mandatory args
: "${TERRAFORM_CONFIG_FILE:?Config file for the test is required}"

export RUN_ENV="${RUN_ENV:=cloud}"

readonly config_prefix=".terraformPluginTests.${RUN_ENV}"

# getting the test setup from the secrets
config_json=$(<"${TERRAFORM_CONFIG_FILE}")
CUSTOMER_ID=$(jq -r "${config_prefix}.customerID" <<<"${config_json}")
TF_VAR_CUSTOMER_NAME=$(jq -r "${config_prefix}.customerName" <<<"${config_json}")
TF_VAR_LOCATION_ID=$(jq -r "${config_prefix}.locationID" <<<"${config_json}")
INDYKITE_APPLICATION_CREDENTIALS=$(jq -r "${config_prefix}.serviceAccountCredentials" <<<"${config_json}")
INDYKITE_SERVICE_ACCOUNT_CREDENTIALS=$(jq -r "${config_prefix}.serviceAccountCredentials" <<<"${config_json}")

export CUSTOMER_ID TF_VAR_LOCATION_ID INDYKITE_APPLICATION_CREDENTIALS INDYKITE_SERVICE_ACCOUNT_CREDENTIALS

# Only export the name if the config actually contains a customerName
if [[ "${TF_VAR_CUSTOMER_NAME}" != "null" ]]; then
    export TF_VAR_CUSTOMER_NAME
fi

PROVIDER_DIR="$(pwd)/provider"
readonly PROVIDER_DIR
APPLIED=0

cleanup() {
    local rc=$?
    trap - EXIT
    if [[ "${APPLIED}" -eq 1 ]]; then
        echo ">>> Cleanup: running terraform destroy (main rc=${rc})"
        cd "${PROVIDER_DIR}" || exit "${rc}"
        [[ -f test.tf.bak ]] && mv -f test.tf.bak test.tf
        if ! terraform destroy -input=false -auto-approve; then
            echo ">>> Cleanup: terraform destroy failed" >&2
            [[ "${rc}" -eq 0 ]] && rc=1
        fi
        rm -f terraform.tfstate terraform.tfstate.backup
    fi
    exit "${rc}"
}
trap cleanup EXIT

cd "${PROVIDER_DIR}"
terraform init -upgrade
terraform plan
# Mark applied BEFORE running apply: partial apply failures still leave state that must be destroyed
APPLIED=1
terraform apply -input=false -auto-approve

cd ../terraform
go test --tags=integration ./...

cd "${PROVIDER_DIR}"
cp test.tf test.tf.bak
sed -i 's/description\(\s*=\s*"\)/description\1Updated - /g' test.tf
terraform apply -input=false -auto-approve
mv test.tf.bak test.tf
# terraform destroy runs automatically via the EXIT trap
