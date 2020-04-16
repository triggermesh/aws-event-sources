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

PREV_RELEASE=${PREV_RELEASE:-$(git describe --always --tags --abbrev=0 ${RELEASE}^)}
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

- awscodecommitsource ([container](https://gcr.io/triggermesh/awscodecommitsource:${RELEASE})) ([linux/amd64](https://github.com/triggermesh/aws-event-sources/releases/download/${RELEASE}/awscodecommitsource-linux-amd64)) ([macos/amd64](https://github.com/triggermesh/aws-event-sources/releases/download/${RELEASE}/awscodecommitsource-darwin-amd64))
- awscognitosource ([container](https://gcr.io/triggermesh/awscognitosource:${RELEASE})) ([linux/amd64](https://github.com/triggermesh/aws-event-sources/releases/download/${RELEASE}/awscognitosource-linux-amd64)) ([macos/amd64](https://github.com/triggermesh/aws-event-sources/releases/download/${RELEASE}/awscognitosource-darwin-amd64))
- awsdynamodbsource ([container](https://gcr.io/triggermesh/awsdynamodbsource:${RELEASE})) ([linux/amd64](https://github.com/triggermesh/aws-event-sources/releases/download/${RELEASE}/awsdynamodbsource-linux-amd64)) ([macos/amd64](https://github.com/triggermesh/aws-event-sources/releases/download/${RELEASE}/awsdynamodbsource-darwin-amd64))
- awskinesis ([container](https://gcr.io/triggermesh/awskinesis:${RELEASE})) ([linux/amd64](https://github.com/triggermesh/aws-event-sources/releases/download/${RELEASE}/awskinesis-linux-amd64)) ([macos/amd64](https://github.com/triggermesh/aws-event-sources/releases/download/${RELEASE}/awskinesis-darwin-amd64))
- awssqssource ([container](https://gcr.io/triggermesh/awssqssource:${RELEASE})) ([linux/amd64](https://github.com/triggermesh/aws-event-sources/releases/download/${RELEASE}/awssqssource-linux-amd64)) ([macos/amd64](https://github.com/triggermesh/aws-event-sources/releases/download/${RELEASE}/awssqssource-darwin-amd64))

## Changelog

${CHANGELOG}
EOF
