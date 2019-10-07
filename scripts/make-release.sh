#!/usr/bin/env bash

set -e

SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ]; do
  DIR="$( cd -P "$( dirname "$SOURCE" )" >/dev/null 2>&1 && pwd )"
  SOURCE="$(readlink "$SOURCE")"
  [[ $SOURCE != /* ]] && SOURCE="$DIR/$SOURCE"
done
BASEDIR="$( cd -P "$( dirname "$SOURCE" )/.." >/dev/null 2>&1 && pwd )"

function upload_artifact() {
  repo_id=${1}
  asset_file=${2}
  echo "Uploading artifact $f to ${GITHUB_ORG}/${REPO_NAME} release ${RELEASE_VERSION} with id: ${repo_id}"

  if [[ ! -f ${asset_file} ]]; then
    echo "file not found $f"
    exit 1
  fi

  curl -H "Authorization: token $GITHUB_TOKEN" -H "Content-Type: $(file -b --mime-type "$asset_file")" --data-binary @"$asset_file" "https://uploads.github.com/repos/${GITHUB_ORG}/${REPO_NAME}/releases/${repo_id}/assets?name=$(basename "$asset_file")"
}

GITHUB_ORG="TheNatureOfSoftware"
REPO_NAME="k3pi"

set +e
RELEASE_VERSION=$(git describe --tags --dirty)
if [[ $? -gt 0 ]]; then
  echo "no release tag found, nothing to do"$?
  exit 0
fi
set -e

if [[ -z "${RELEASE_VERSION}" ]]; then
  echo "No release tag found"
  exit 0
fi

echo "Making release ${RELEASE_VERSION}"

API_JSON=$(printf '{"tag_name": "%s","target_commitish": "master","name": "%s","body": "Release of version %s","draft": false,"prerelease": false}' $RELEASE_VERSION $RELEASE_VERSION $RELEASE_VERSION)
repo_id=$(curl --data "$API_JSON" -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/repos/${GITHUB_ORG}/${REPO_NAME}/releases | jq -r '.id')

if [[ "$repo_id" == "" ]]; then
  echo "Failed: repository id is empty"
  exit 1
fi

if [[ -d ${BASEDIR}/bin ]]; then
  for bin_file in ${BASEDIR}/bin/*; do
    upload_artifact ${repo_id} ${bin_file}
  done
fi


