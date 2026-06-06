#!/bin/bash -xe

MODE="${1:-finalize}"

VERSION=$(grep 'RELEASE_VERSION[[:space:]]*=' version.go  | awk -F= '{print $2}' | sed -e 's_"__g' -e 's/[[:space:]]//g')

if [[ ! "${VERSION}" =~ ^([0-9]+[.][0-9]+)[.]([0-9]+)(-(alpha|beta)[.]([0-9]+))?$ ]]; then
  echo "Version ${VERSION} must be 'X.Y.Z', 'X.Y.Z-alpha.N', or 'X.Y.Z-beta.N'"
  exit 1
fi

if [ "${MODE}" = "--validate-only" ]; then
  exit 0
fi

if [ "${MODE}" != "finalize" ]; then
  echo "Usage: $0 [--validate-only]"
  exit 1
fi

MINOR=${BASH_REMATCH[1]}
RELEASE_BRANCH="release-${MINOR}"
CURRENT_COMMIT=$(git rev-parse HEAD)

remote_ref_sha() {
  local ref="$1"

  git ls-remote origin "${ref}" | awk '{print $1}' | head -n1
}

ensure_remote_tag() {
  local tag_name="v${VERSION}"
  local remote_tag_sha
  local remote_peeled_sha
  local local_peeled_sha

  remote_tag_sha=$(remote_ref_sha "refs/tags/${tag_name}")

  if [ -n "${remote_tag_sha}" ]; then
    remote_peeled_sha=$(remote_ref_sha "refs/tags/${tag_name}^{}")

    if [ -z "${remote_peeled_sha}" ]; then
      echo "Remote tag ${tag_name} exists, but its peeled commit could not be resolved"
      exit 1
    fi

    if [ "${remote_peeled_sha}" != "${CURRENT_COMMIT}" ]; then
      echo "Remote tag ${tag_name} points to ${remote_peeled_sha}, not ${CURRENT_COMMIT}"
      exit 1
    fi

    echo "Tag ${tag_name} already exists on origin"
    return
  fi

  if git rev-parse -q --verify "refs/tags/${tag_name}" >/dev/null; then
    local_peeled_sha=$(git rev-parse "${tag_name}^{}")

    if [ "${local_peeled_sha}" != "${CURRENT_COMMIT}" ]; then
      echo "Local tag ${tag_name} points to ${local_peeled_sha}, not ${CURRENT_COMMIT}"
      exit 1
    fi
  else
    git tag -a -m "Release ${VERSION}" "${tag_name}"
  fi

  git push origin "refs/tags/${tag_name}"
}

ensure_release_branch() {
  local remote_branch_sha
  local local_branch_sha

  if [[ ! "${VERSION}" =~ .0-beta.1$ ]]; then
    return
  fi

  remote_branch_sha=$(remote_ref_sha "refs/heads/${RELEASE_BRANCH}")

  if [ -n "${remote_branch_sha}" ]; then
    if [ "${remote_branch_sha}" != "${CURRENT_COMMIT}" ]; then
      echo "Remote branch ${RELEASE_BRANCH} points to ${remote_branch_sha}, not ${CURRENT_COMMIT}"
      exit 1
    fi

    echo "Branch ${RELEASE_BRANCH} already exists on origin"
    return
  fi

  if git rev-parse -q --verify "refs/heads/${RELEASE_BRANCH}" >/dev/null; then
    local_branch_sha=$(git rev-parse "${RELEASE_BRANCH}")

    if [ "${local_branch_sha}" != "${CURRENT_COMMIT}" ]; then
      echo "Local branch ${RELEASE_BRANCH} points to ${local_branch_sha}, not ${CURRENT_COMMIT}"
      exit 1
    fi
  else
    git branch "${RELEASE_BRANCH}" "${CURRENT_COMMIT}"
  fi

  git push origin "refs/heads/${RELEASE_BRANCH}:refs/heads/${RELEASE_BRANCH}"
}

ensure_remote_tag
ensure_release_branch
