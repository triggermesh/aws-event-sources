#!/usr/bin/env sh

RELEASE=${1:-${GIT_TAG}}
RELEASE=${RELEASE:-${CIRCLE_TAG}}

if [ -z "${RELEASE}" ]; then
	echo "Usage:"
	echo "release-notes.sh VERSION"
	exit 1
fi

if ! git rev-list ${RELEASE} >/dev/null 2>&1; then
	echo "${RELEASE} does not exist"
	exit
fi

KREPO="aws-event-sources"
BASE_URL="https://github.com/triggermesh/${KREPO}/releases/download/${RELEASE}"
PREV_RELEASE=${PREV_RELEASE:-$(git describe --tags --abbrev=0 ${RELEASE}^ 2>/dev/null)}
PREV_RELEASE=${PREV_RELEASE:-$(git rev-list --max-parents=0 ${RELEASE}^ 2>/dev/null)}
NOTABLE_CHANGES=$(git cat-file -p ${RELEASE} | sed '/-----BEGIN PGP SIGNATURE-----/,//d' | tail -n +6)
CHANGELOG=$(git log --no-merges --pretty=format:'- [%h] %s (%aN)' ${PREV_RELEASE}..${RELEASE})
if [ $? -ne 0 ]; then
	echo "Error creating changelog"
	exit 1
fi

COMMANDS=$(sed -n -e "s/^\(COMMANDS[[:space:]]*=[[:space:]]*\)\(.*\)$/\2/p" Makefile)
PLATFORMS=$(sed -n -e "s/^\(TARGETS[[:space:]]*?=[[:space:]]*\)\(.*\)$/\2/p" Makefile)
RELEASE_ASSETS_TABLE=$(
  echo -n "|"; for command in ${COMMANDS}; do echo -n " ${command} |"; done ; echo
  echo -n "|"; for command in ${COMMANDS}; do echo -n "--|"; done ; echo
  echo -n "|"
  for command in ${COMMANDS}; do
    echo -n " ([container](https://gcr.io/triggermesh/${command}:${RELEASE}))"
    for platform in ${PLATFORMS}; do
      echo -n " ([${platform}](${BASE_URL}/${command}-${platform%/*}-${platform#*/}))"
    done
    echo -n " |"
  done
  echo
)

cat <<EOF
${NOTABLE_CHANGES}

## Installation

Download TriggerMesh Knative event sources for AWS ${RELEASE}

${RELEASE_ASSETS_TABLE}

## Changelog

${CHANGELOG}
EOF
