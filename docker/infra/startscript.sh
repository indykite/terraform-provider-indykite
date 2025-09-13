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

cd ./*/ || exit

# setup reporting
export RELEASE_VERSION="${RELEASE_VERSION:=unknown}"
run_date=$(date +%Y%m%d-%H%M)
results_file_name="${RELEASE_VERSION}_results_terraform_${RUN_ENV}_${run_date}_report.txt"

# setup GCloud
export BUCKET_NAME="${BUCKET_NAME:=terraform_results_deploy}"
export SECRET_NAME=${SECRET_NAME:=terraformPluginTests}
storage="https://storage.cloud.google.com/${BUCKET_NAME}/${results_file_name}"

export RUN_ENV="${RUN_ENV:=staging}"

# getting the test setup from the secrets
secret_json=$(gcloud secrets versions access latest --secret="${SECRET_NAME}")
CUSTOMER_ID=$(jq -r ".terraformPluginTests.${RUN_ENV}.customerID" <<<"${secret_json}")
TF_VAR_LOCATION_ID=$(jq -r ".terraformPluginTests.${RUN_ENV}.locationID" <<<"${secret_json}")
INDYKITE_APPLICATION_CREDENTIALS=$(jq -r ".terraformPluginTests.${RUN_ENV}.serviceAccountCredentials" <<<"${secret_json}")
INDYKITE_SERVICE_ACCOUNT_CREDENTIALS=$(jq -r ".terraformPluginTests.${RUN_ENV}.serviceAccountCredentials" <<<"${secret_json}")
export CUSTOMER_ID TF_VAR_LOCATION_ID INDYKITE_APPLICATION_CREDENTIALS INDYKITE_SERVICE_ACCOUNT_CREDENTIALS

make upgrade_test_provider integration 2>output.txt
retVal=$?

# we are moving away of this kind of slack messaging, so it is optional for now
if [[ -n "${SLACK_WEBHOOK_URL}" ]]; then
    app_name="indykite_${RUN_ENV}"
    github_sha=$(git rev-parse --short HEAD)
    if [[ ${retVal} -ne 0 ]]; then
        echo "There was an error during terraform run: $(cat output.txt || true)"
        message="Test errors: ${retVal}"
        attachment_message=":alert: Tests failed :alert:"
        repair_message="To see the errors, open the logs or go to Github and launch the tests manually"
        colour="#FF0000"
        blocks=$(jq -n \
            --arg sha "${github_sha}" \
            --arg app "${app_name}" \
            --arg storage "${storage}" \
            --arg message "${message}" \
            --arg repair "${repair_message}" \
            --arg colour "${colour}" \
            --arg attach "${attachment_message}" \
            '{
    blocks: [
      {type: "divider"},
      {type: "section", text: {type: "mrkdwn", text: "Test results - *Terraform Plugin tests* - \($sha) - triggered by \($app) <\($storage)?authuser=0|Logs>"}},
      {type: "section", fields: [
        {type: "mrkdwn", text: $message},
        {type: "mrkdwn", text: $repair}
      ]},
      {type: "divider"}
    ],
    attachments: [
      {title: $message, color: $colour, fields: [$attach]}
    ]
  }')

    #
    else
        message="All tests passed"
        attachment_message=":heavy_check_mark: All Passed :heavy_check_mark:"
        repair_message="All good"
        colour="#008000"
        blocks=$(jq -n \
            --arg sha "${github_sha}" \
            --arg app "${app_name}" \
            --arg message "${message}" \
            --arg repair "${repair_message}" \
            --arg colour "${colour}" \
            --arg attach "${attachment_message}" \
            '{
    blocks: [
      {type: "divider"},
      {type: "section", text: {type: "mrkdwn", text: "Test results - *Terraform Plugin tests* - \($sha) - triggered by \($app)"}},
      {type: "section", fields: [
        {type: "mrkdwn", text: $message},
        {type: "mrkdwn", text: $repair}
      ]},
      {type: "divider"}
    ],
    attachments: [
      {title: $message, color: $colour, fields: [$attach]}
    ]
  }')
    fi

    curl --header "Content-Type: application/json" --data "${blocks}" -X POST "${SLACK_WEBHOOK_URL}"
fi

echo "Send results to GCP bucket"
BUCKET_PATH="gs://${BUCKET_NAME}"
mv output.txt "${results_file_name}"

echo "Copying the results to google cloud"
echo "Copying ${results_file_name} to ${BUCKET_PATH}"
gsutil -q cp "${results_file_name}" "${BUCKET_PATH}"

echo "Logs: ${storage}?authuser=0"
exit "${retVal}"
