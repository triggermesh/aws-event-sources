#!/usr/bin/env sh

RELEASE=${GIT_TAG:-$1}

if [ -z "${RELEASE}" ]; then
	echo "Usage:"
	echo "./scripts/release-notes.sh v0.1.0"
	exit 1
fi

if ! git rev-list ${RELEASE} >/dev/null 2>&1; then
	echo "${RELEASE} does not exist"
	exit
fi

PREV_RELEASE=$(git describe --always --tags --abbrev=0 ${RELEASE}^)
NOTABLE_CHANGES=$(git cat-file -p ${RELEASE} | sed '/-----BEGIN PGP SIGNATURE-----/,//d' | tail -n +6)
CHANGELOG=$(git log --no-merges --pretty=format:'- [%h] %s (%aN)' ${PREV_RELEASE}..${RELEASE})
if [ $? -ne 0 ]; then
	echo "Error creating changelog"
	exit 1
fi

cat <<EOF
${NOTABLE_CHANGES}

## Installation

Download Knative event sources for AWS ${RELEASE}

- [awscodecommit (linux/amd64)](https://github.com/triggermesh/aws-event-sources/releases/download/${RELEASE}/awscodecommit-linux-amd64)
- [awscodecommit (macos/amd64)](https://github.com/triggermesh/aws-event-sources/releases/download/${RELEASE}/awscodecommit-darwin-amd64)
- [awscognito (linux/amd64)](https://github.com/triggermesh/aws-event-sources/releases/download/${RELEASE}/awscognito-linux-amd64)
- [awscognito (macos/amd64)](https://github.com/triggermesh/aws-event-sources/releases/download/${RELEASE}/awscognito-darwin-amd64)
- [awsdynamodb (linux/amd64)](https://github.com/triggermesh/aws-event-sources/releases/download/${RELEASE}/awsdynamodb-linux-amd64)
- [awsdynamodb (macos/amd64)](https://github.com/triggermesh/aws-event-sources/releases/download/${RELEASE}/awsdynamodb-darwin-amd64)
- [awskinesis (linux/amd64)](https://github.com/triggermesh/aws-event-sources/releases/download/${RELEASE}/awskinesis-linux-amd64)
- [awskinesis (macos/amd64)](https://github.com/triggermesh/aws-event-sources/releases/download/${RELEASE}/awskinesis-darwin-amd64)
- [awssqs (linux/amd64)](https://github.com/triggermesh/aws-event-sources/releases/download/${RELEASE}/awssqs-linux-amd64)
- [awssqs (macos/amd64)](https://github.com/triggermesh/aws-event-sources/releases/download/${RELEASE}/awssqs-darwin-amd64)

## Changelog

${CHANGELOG}
EOF
