#!/bin/sh
export REPO_NAME="terraform-provider-indykite"
export LOCALREPO="${APPUSER_HOME}/github/${REPO_NAME}"
export LOCALREPO_VC_DIR="${LOCALREPO}/.git"
export GITHUB_REPO="${GITHUB_REPO:=master}"

if [ ! -d ${LOCALREPO_VC_DIR} ]
then
    git clone --branch ${GITHUB_BRANCH} "${GITHUB_REPO}" "${LOCALREPO}"
    cd ./${REPO_NAME} || exit
else
    cd ${LOCALREPO}
    git pull origin ${GITHUB_BRANCH}
fi

make upgrade_test_provider integration
