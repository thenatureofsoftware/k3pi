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
  f=${1}
  echo "Uploading artifact $f to ${GITHUB_ORG}/${REPO_NAME} release v${RELEASE_VERSION}"

  if [[ ! -f ${f} ]]; then
    echo "file not found $f"
    exit 1
  fi

  curl \
-H "Authorization: token $GITHUB_TOKEN" \
-H "Content-Type: $(file -b --mime-type "$f")" \
--data-binary @"$f" "https://uploads.github.com/repos/${GITHUB_ORG}/${REPO_NAME}/releases/v${RELEASE_VERSION}/assets?name=$(basename "$f")"
}

GITHUB_ORG="TheNatureOfSoftware"
REPO_NAME="k3pi"
RELEASE_VERSION=$(git describe --tags --dirty)

if [[ -z "${RELEASE_VERSION}" ]]; then
  echo "No release tag found"
  exit 0
fi

echo "Making release ${RELEASE_VERSION}"

API_JSON=$(printf '{"tag_name": "v%s","target_commitish": "master","name": "v%s","body": "Release of version %s","draft": false,"prerelease": false}' $RELEASE_VERSION $RELEASE_VERSION $RELEASE_VERSION)
curl --data "$API_JSON" \
-H "Authorization: token $GITHUB_TOKEN" \
https://api.github.com/repos/${GITHUB_ORG}/${REPO_NAME}/releases

if [[ -d ${BASEDIR}/bin ]]; then
  for binFile in ${BASEDIR}/bin/*; do
    upload_artifact ${binFile}
  done
fi


