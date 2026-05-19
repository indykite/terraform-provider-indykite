#!/bin/bash
# getting the latest release tag so the released terraform is used
# tag=$(git ls-remote --tags --exit-code --refs https://github.com/indykite/terraform-provider-indykite | sed -E 's/^[[:xdigit:]]+[[:space:]]+refs\/tags\/(.+)/\1/g' | sort --version-sort | tail -n1)

# pipefail - BASH only, not supported in POSIX Shell
set -o errexit -o nounset # -o pipefail

# default values and check for the mandatory args
: "${APPUSER_HOME:?User home path is required}"
: "${BRANCH:="master"}"
: "${GITHUB_USER:?GitHub User is required}"
: "${GITHUB_TOKEN:?GitHub Token is required}"
: "${GITHUB:?GitHub FQDN is required}"
: "${SLACK_WEBHOOK_URL:=""}"
: "${BUCKET_NAME:?Bucket name is required}"

# --------- GET THE CODE, WITHOUT IT USELESS TO SET UP THE REST ---------
i=0
while [[ ${i} -le 5 ]]; do
    git clone --single-branch --branch "${BRANCH}" "https://${GITHUB_USER}:${GITHUB_TOKEN}@${GITHUB}"
    retVal=$?
    if [[ ${retVal} -ne 0 ]]; then
        sleep 1
        ((i++))
    else
        break
    fi
done
if [[ ${retVal} -ne 0 ]]; then
    echo "Failed to clone https://****:****@${GITHUB}"
    exit "${retVal}"
fi
echo "I'll clone the ${BRANCH} version"
# --------- END OF GETTING THE CODE --------

cd ./*/ || exit

# setup reporting
export RUN_ENV="${RUN_ENV:=staging}"
export RELEASE_VERSION="${RELEASE_VERSION:=unknown}"

run_date=$(date +%Y%m%d-%H%M)
readonly BUCKET_PATH="gs://${BUCKET_NAME}"
readonly results_file_name="${RELEASE_VERSION}_results_terraform_${RUN_ENV}_${run_date}_report.txt"
readonly config_prefix=".terraformPluginTests.${RUN_ENV}"

# setup GCloud
export BUCKET_NAME="${BUCKET_NAME:=terraform_results_deploy}"
export SECRET_NAME=${SECRET_NAME:=terraformPluginTests}
readonly storage="https://storage.cloud.google.com/${BUCKET_NAME}/${results_file_name}"

# getting the test setup from the secrets
config_json=$(gcloud secrets versions access latest --secret="${SECRET_NAME}")
CUSTOMER_ID=$(jq -r "${config_prefix}.customerID" <<<"${config_json}")
TF_VAR_LOCATION_ID=$(jq -r "${config_prefix}.locationID" <<<"${config_json}")
INDYKITE_APPLICATION_CREDENTIALS=$(jq -r "${config_prefix}.serviceAccountCredentials" <<<"${config_json}")
INDYKITE_SERVICE_ACCOUNT_CREDENTIALS=$(jq -r "${config_prefix}.serviceAccountCredentials" <<<"${config_json}")
export CUSTOMER_ID TF_VAR_LOCATION_ID INDYKITE_APPLICATION_CREDENTIALS INDYKITE_SERVICE_ACCOUNT_CREDENTIALS

# --------- RUN TESTS --------
if make upgrade_test_provider integration 2>"${results_file_name}"; then
    exit_code=0
else
    exit_code=$?
    echo "tests failed:"
    echo "$(<"${results_file_name}")"
fi
# --------- END RUN TESTS ------

echo "Send results to GCP bucket"
echo "Copying the results to google cloud"
echo "Copying ${results_file_name} to ${BUCKET_PATH}"
gsutil -q cp "${results_file_name}" "${BUCKET_PATH}"

echo "Logs: ${storage}?authuser=0"
exit "${exit_code}"
