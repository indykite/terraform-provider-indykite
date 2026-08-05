#!/bin/sh
# pipefail - BASH only, not supported in POSIX Shell
set -o errexit -o nounset # -o pipefail

# default values and check for the mandatory args
: "${APPUSER_HOME:?User home path is required}"
: "${GITHUB_BRANCH:?Branch is required}"

# Materialize the SSH deploy key at runtime; the image itself contains no key material.
# Pass it with 'docker run -e SSH_PRIVATE_KEY=...' or mount it as ~/.ssh/id_ed25519.
if [ ! -f "${APPUSER_HOME}/.ssh/id_ed25519" ]; then
    : "${SSH_PRIVATE_KEY:?SSH private key is required (env SSH_PRIVATE_KEY or mounted ~/.ssh/id_ed25519)}"
    # printf adds the trailing newline OpenSSH requires, in case the env var lost it
    printf '%s\n' "${SSH_PRIVATE_KEY}" >"${APPUSER_HOME}/.ssh/id_ed25519"
    chmod 600 "${APPUSER_HOME}/.ssh/id_ed25519"
fi

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
