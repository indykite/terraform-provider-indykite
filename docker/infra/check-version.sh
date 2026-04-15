#!/bin/bash
# Pre-commit hook: auto-bump infra image versions based on what changed.
#
# Image/file mapping (mirrors .github/workflows/docker-build-tf-plugin-tests.yaml):
#   pipeline image      → docker/infra/Dockerfile, docker/infra/pipeline/startscript.sh
#                         → docker/infra/pipeline/.version
#   private cloud image → docker/infra/Dockerfile, docker/infra/private_cloud/startscript.sh, tests/**
#                         → docker/infra/private_cloud/.version
#
# Bump rules (per image, highest-level bump wins):
#   Dockerfile change    → minor bump (X.Y+1.0) for BOTH images
#   startscript change   → patch bump (X.Y.Z+1) for the owning image
#   tests/** change      → patch bump for the private-cloud image only
#
# If the user already staged a change to a .version file, that image is left
# alone — this preserves the override path for manual major/minor bumps.

set -o errexit -o nounset -o pipefail

# Bump a version file. $1 = file path, $2 = "minor" | "patch".
# Preserves an optional leading `v`. Git-adds the modified file.
bump_version() {
    local file=$1
    local kind=$2
    local current
    current=$(tr -d '[:space:]' <"${file}")

    local ver=${current}
    local prefix=""
    if [[ "${ver}" == v* ]]; then
        prefix="v"
        ver=${ver#v}
    fi

    if [[ ! "${ver}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "ERROR: ${file} does not contain an X.Y.Z version (got: '${current}')" >&2
        return 1
    fi

    local major minor patch
    IFS='.' read -r major minor patch <<<"${ver}"

    case "${kind}" in
    minor)
        minor=$((minor + 1))
        patch=0
        ;;
    patch)
        patch=$((patch + 1))
        ;;
    *)
        echo "ERROR: unknown bump kind '${kind}'" >&2
        return 1
        ;;
    esac

    local new="${prefix}${major}.${minor}.${patch}"
    echo "${new}" >"${file}"
    git add "${file}"
    echo "  ${file}: ${current} → ${new} (${kind})"
}

staged_files=$(git diff --cached --name-only)

pipeline_bump=""
private_cloud_bump=""
pipeline_version_staged=false
private_cloud_version_staged=false

# Rank: "minor" > "patch" > "". Upgrade but never downgrade.
upgrade_bump() {
    local current=$1
    local new=$2
    if [[ "${current}" == "minor" ]]; then
        echo "${current}"
    elif [[ "${new}" == "minor" ]]; then
        echo "${new}"
    elif [[ -n "${new}" ]]; then
        echo "${new}"
    else
        echo "${current}"
    fi
}

while IFS= read -r file; do
    [[ -z "${file}" ]] && continue
    case "${file}" in
    docker/infra/Dockerfile)
        pipeline_bump=$(upgrade_bump "${pipeline_bump}" "minor")
        private_cloud_bump=$(upgrade_bump "${private_cloud_bump}" "minor")
        ;;
    docker/infra/pipeline/startscript.sh)
        pipeline_bump=$(upgrade_bump "${pipeline_bump}" "patch")
        ;;
    docker/infra/private_cloud/startscript.sh)
        private_cloud_bump=$(upgrade_bump "${private_cloud_bump}" "patch")
        ;;
    tests/*)
        private_cloud_bump=$(upgrade_bump "${private_cloud_bump}" "patch")
        ;;
    docker/infra/pipeline/.version)
        pipeline_version_staged=true
        ;;
    docker/infra/private_cloud/.version)
        private_cloud_version_staged=true
        ;;
    *)
        : # other staged files do not affect image versions
        ;;
    esac
done <<<"${staged_files}"

bumped=0

if [[ -n "${pipeline_bump}" && "${pipeline_version_staged}" == "false" ]]; then
    [[ "${bumped}" -eq 0 ]] && echo "check-version: bumping infra image versions:"
    bump_version docker/infra/pipeline/.version "${pipeline_bump}"
    bumped=1
fi

if [[ -n "${private_cloud_bump}" && "${private_cloud_version_staged}" == "false" ]]; then
    [[ "${bumped}" -eq 0 ]] && echo "check-version: bumping infra image versions:"
    bump_version docker/infra/private_cloud/.version "${private_cloud_bump}"
    bumped=1
fi

if [[ "${bumped}" -eq 1 ]]; then
    echo "check-version: .version file(s) updated and staged — re-run the commit to include them."
    exit 1
fi

exit 0
