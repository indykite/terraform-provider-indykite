#!/usr/bin/env bash
# Pre-commit hook: auto-bump infra image versions based on what changed.
#   - change under docker/infra/  -> bump MINOR of both ci/.version and full/.version
#   - change under tests         -> bump PATCH of full/.version
# When a required bump is missing, the hook writes the new version(s) and fails
# the commit so you can stage and re-commit with the bump included.

set -euo pipefail

VERSION_FILE_FULL="docker/infra/full/.version"
VERSION_FILE_CI="docker/infra/ci/.version"

# Limit detection to staged changes — pre-commit's normal scope.
# Exclude the .version files from the infra check, otherwise bumping a version
# file would itself look like a docker/infra change on the next commit.
STAGED_TESTS=$(git diff --name-only --cached -- 'tests/' || true)
STAGED_INFRA=$(git diff --name-only --cached -- 'docker/infra/' \
    ':(exclude)docker/infra/full/.version' \
    ':(exclude)docker/infra/ci/.version' || true)

if [[ -z "${STAGED_TESTS}" && -z "${STAGED_INFRA}" ]]; then
    exit 0
fi

if [[ ! -f "${VERSION_FILE_FULL}" ]]; then
    echo "❌ ${VERSION_FILE_FULL} not found — cannot bump version."
    exit 1
fi
if [[ ! -f "${VERSION_FILE_CI}" ]]; then
    echo "❌ ${VERSION_FILE_CI} not found — cannot bump version."
    exit 1
fi

# bump_version <file> <minor|patch> — rewrites <file> with the bumped version.
bump_version() {
    local file="$1" part="$2" current major minor patch
    current=$(tr -d '[:space:]' <"${file}")
    # Expect format like v<MAJOR>.<MINOR>.<PATCH>
    if [[ ! "${current}" =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
        echo "❌ Unexpected version format in ${file}: '${current}' (expected vMAJOR.MINOR.PATCH)"
        exit 1
    fi
    major="${BASH_REMATCH[1]}"
    minor="${BASH_REMATCH[2]}"
    patch="${BASH_REMATCH[3]}"
    if [[ "${part}" == "minor" ]]; then
        minor=$((minor + 1))
        patch=0
    else
        patch=$((patch + 1))
    fi
    printf 'v%s.%s.%s\n' "${major}" "${minor}" "${patch}" >"${file}"
}

# staged_path <file> — prints <file> if it's already staged (bumped this commit), else nothing.
staged_path() {
    git diff --name-only --cached -- "$1" || true
}

bumped=0

if [[ -n "${STAGED_INFRA}" ]]; then
    # infra change -> minor bump on BOTH files (takes precedence over a tests/ bump)
    full_staged="$(staged_path "${VERSION_FILE_FULL}")"
    ci_staged="$(staged_path "${VERSION_FILE_CI}")"
    if [[ -z "${full_staged}" ]]; then
        bump_version "${VERSION_FILE_FULL}" minor
        bumped=1
    fi
    if [[ -z "${ci_staged}" ]]; then
        bump_version "${VERSION_FILE_CI}" minor
        bumped=1
    fi
elif [[ -n "${STAGED_TESTS}" ]]; then
    # tests change -> patch bump on full only
    full_staged="$(staged_path "${VERSION_FILE_FULL}")"
    if [[ -z "${full_staged}" ]]; then
        bump_version "${VERSION_FILE_FULL}" patch
        bumped=1
    fi
fi

if [[ "${bumped}" -eq 1 ]]; then
    cat <<EOF
❌ Changes detected under tests/ and/or docker/infra/.
   The version file(s) below have been bumped by this hook:
     - ${VERSION_FILE_FULL}
     - ${VERSION_FILE_CI}
   Stage the change(s) and commit again, e.g.:
     git add ${VERSION_FILE_FULL} ${VERSION_FILE_CI}
EOF
    exit 1
fi

echo "✅ Version file(s) already updated in this commit; changes accepted."
exit 0
