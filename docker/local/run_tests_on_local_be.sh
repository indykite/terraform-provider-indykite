#!/bin/sh
# pipefail - BASH only, not supported in POSIX Shell
set -o errexit -o nounset # -o pipefail

# default values and check for the mandatory args
: "${APPUSER_HOME:?User home path is required}"
: "${GITHUB_BRANCH:?Branch is required}"

export REPO_NAME="terraform-provider-indykite"
export LOCALREPO="${APPUSER_HOME}/github/${REPO_NAME}"
export LOCALREPO_VC_DIR="${LOCALREPO}/.git"
export GITHUB_REPO="${GITHUB_REPO:=master}"

if [ ! -d "${LOCALREPO_VC_DIR}" ]; then
    git clone --branch "${GITHUB_BRANCH}" "${GITHUB_REPO}" "${LOCALREPO}"
    cd "./${REPO_NAME}" || exit
else
    cd "${LOCALREPO}" || exit
    git pull origin "${GITHUB_BRANCH}"
fi

make upgrade_test_provider integration
