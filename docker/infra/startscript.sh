#!/bin/bash
# getting the latest release tag so the released terraform is used
# tag=$(git ls-remote --tags --exit-code --refs https://github.com/indykite/terraform-provider-indykite | sed -E 's/^[[:xdigit:]]+[[:space:]]+refs\/tags\/(.+)/\1/g' | sort --version-sort | tail -n1)

# Adding a retry so let the istio start properly before the tests start
export BRANCH="${BRANCH:=master}"
i=0
while [ $i -le 5 ]; do
  git clone --single-branch --branch "${BRANCH}" "https://${GITHUB_USER}:${GITHUB_TOKEN}@${GITHUB}"
  retVal=$?
  if [ $retVal -ne 0 ]; then
    sleep 1
    ((i++))
  else
    break
  fi
done
if [ $retVal -ne 0 ]; then
  echo "Failed to clone https://****:****@$GITHUB"
  exit $retVal
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
export CUSTOMER_ID=$(gcloud secrets versions access latest --secret=${SECRET_NAME}  | jq --raw-output  .terraformPluginTests.${RUN_ENV}.customerID)
export TF_VAR_LOCATION_ID=$(gcloud secrets versions access latest --secret=${SECRET_NAME}  | jq --raw-output  .terraformPluginTests.${RUN_ENV}.locationID)
export INDYKITE_APPLICATION_CREDENTIALS=$(gcloud secrets versions access latest --secret=${SECRET_NAME}  | jq --raw-output  .terraformPluginTests.${RUN_ENV}.serviceAccountCredentials)
export INDYKITE_SERVICE_ACCOUNT_CREDENTIALS=$(gcloud secrets versions access latest --secret=${SECRET_NAME}  | jq --raw-output  .terraformPluginTests.${RUN_ENV}.serviceAccountCredentials)

make upgrade_test_provider integration 2> output.txt
retVal=$?

# we are moving away of this kind of slack messaging, so it is optional for now
if [ ! -z "${SLACK_WEBHOOK_URL}" ]; then
  app_name="indykite_${RUN_ENV}"
  github_sha=$(git rev-parse --short HEAD)
  if [ ${retVal} -ne 0 ]; then
    echo "There was an error during terraform run: $(cat output.txt)"
    message="Test errors: ${retVal}"
    attachment_message=":alert: Tests failed :alert:"
    repair_message="To see the errors, open the logs or go to Github and launch the tests manually"
    colour="#FF0000"
    blocks='{"blocks": [{ "type": "divider" }, {"type": "section", "text": {"type": "mrkdwn", "text": "Test results - *Terraform Plugin tests* - `'${github_sha}'` - triggered by `'${app_name}'` <'${storage}'?authuser=0|Logs>"}}, {"type":"section", "fields":[{"type": "mrkdwn", "text": "'${message}'"},{"type": "mrkdwn", "text": "'${repair_message}'"}]},{"type":"divider"}],"attachments": [{"title": "'${message}'", "color": "'${colour}'", "fields": ["'${attachment_message}'"]}]}'
  else
    message="All tests passed"
    attachment_message=":heavy_check_mark: All Passed :heavy_check_mark:"
    repair_message="All good"
    colour="#008000"
    blocks='{"blocks": [{ "type": "divider" }, {"type": "section", "text": {"type": "mrkdwn", "text": "Test results - *Terraform Plugin tests* - `'${github_sha}'` - triggered by `'${app_name}'` "}}, {"type":"section", "fields":[{"type": "mrkdwn", "text": "'${message}'"},{"type": "mrkdwn", "text": "'${repair_message}'"}]},{"type":"divider"}],"attachments": [{"title": "'${message}'", "color": "'${colour}'", "fields": ["'${attachment_message}'"]}]}'
  fi

  curl --header "Content-Type: application/json" --data "$blocks" -X POST $SLACK_WEBHOOK_URL
fi

echo "Send results to GCP bucket"
BUCKET_PATH="gs://${BUCKET_NAME}"
mv output.txt ${results_file_name}

echo "Copying the results to google cloud"
echo "Copying ${results_file_name} to $BUCKET_PATH"
gsutil -q cp ${results_file_name} $BUCKET_PATH

echo "Logs: ${storage}?authuser=0"
exit $retVal
