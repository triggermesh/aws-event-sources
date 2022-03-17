[![Release](https://img.shields.io/github/v/release/triggermesh/aws-event-sources?label=release)](https://github.com/triggermesh/aws-event-sources/releases) [![Downloads](https://img.shields.io/github/downloads/triggermesh/aws-event-sources/total?label=downloads)](https://github.com/triggermesh/aws-event-sources/releases) [![CircleCI](https://circleci.com/gh/triggermesh/aws-event-sources/tree/master.svg?style=shield)](https://circleci.com/gh/triggermesh/aws-event-sources/tree/master) [![Go Report Card](https://goreportcard.com/badge/github.com/triggermesh/aws-event-sources)](https://goreportcard.com/report/github.com/triggermesh/aws-event-sources) [![License](https://img.shields.io/github/license/triggermesh/aws-event-sources?label=license)](LICENSE)

THIS REPOSITORY IS NOW ARCHIVED IN FAVOR OF (https://github.com/triggermesh/triggermesh)[https://github.com/triggermesh/triggermesh]

![Sources for Amazon Web Services](./images/saws.png "Sources for Amazon Web Services")

TriggerMesh Sources for Amazon Web Services (SAWS) allow you to quickly and easily consume events from your AWS services
and send them to workloads running in your cluster.

Other Knative Sources maintained by TriggerMesh are available in the following repositories:

- [Knative Sources][knsrc]
- [GitLab Source][knsrc-gitlab] (Knative sandbox project)

## Installation

### Kubernetes

Using Helm:

```bash
$ helm repo add triggermesh https://storage.googleapis.com/triggermesh-charts
$ helm install triggermesh/aws-event-sources
```

Refer to the [aws-event-sources chart documentation](chart/README.md) for all available configuration options.

### OpenShift

Login to the OpenShift Container Platform console and install the **AWS Sources Operator** from the **OperatorHub**.
Refer to the [documentation][operator] of the AWS Event Sources Operator for the complete guide in getting up and
running on the OpenShift Container Platform.

## Getting Started

The following table lists the AWS services currently supported by TriggerMesh Sources for AWS and their support level.

| AWS Service                                                       | Documentation                                    | Support Level |
|-------------------------------------------------------------------|--------------------------------------------------|---------------|
| [CloudWatch](https://aws.amazon.com/cloudwatch/)                  |                                                  | alpha         |
| [CloudWatch Logs](https://aws.amazon.com/cloudwatch/)             |                                                  | alpha         |
| [CodeCommit](https://aws.amazon.com/codecommit/)                  | [README](cmd/awscodecommitsource/README.md)      | alpha         |
| [Cognito Identity Pool](https://aws.amazon.com/cognito/)          | [README](cmd/awscognitoidentitysource/README.md) | alpha         |
| [Cognito User Pool](https://aws.amazon.com/cognito/)              | [README](cmd/awscognitouserpoolsource/README.md) | alpha         |
| [DynamoDB](https://aws.amazon.com/dynamodb/)                      | [README](cmd/awsdynamodbsource/README.md)        | alpha         |
| [Kinesis](https://aws.amazon.com/kinesis/)                        | [README](cmd/awskinesissource/README.md)         | alpha         |
| [Simple Cloud Storage (S3)](https://aws.amazon.com/s3/)           |                                                  | alpha         |
| [Simple Notifications Service (SNS)](https://aws.amazon.com/sns/) | [README](cmd/awssnssource/README.md)             | alpha         |
| [Simple Queue Service (SQS)](https://aws.amazon.com/sqs/)         | [README](cmd/awssqssource/README.md)             | alpha         |

For detailed usage instructions about a particular source, please refer to its linked `README.md` file, or to the
[TriggerMesh documentation][tm-docs].

## Contributions and support

We would love to hear your feedback on these sources. Please don't hesitate to submit bug reports and suggestions by
[filing issues][gh-issue], or contribute by [submitting pull-requests][gh-pr].

Refer to [DEVELOPMENT.md](./DEVELOPMENT.md) in order to get started.

## TriggerMesh Cloud Early Access

TriggerMesh Sources for Amazon Web Services can be used as is from this repo. You can also use them along with other
components from our Cloud at <https://cloud.triggermesh.io>, which has a web UI to configure and run them.

## Commercial Support

TriggerMesh Inc. supports those sources commercially. Email us at <info@triggermesh.com> to get more details.

## Code of Conduct

Although this project is not part of the [CNCF][cncf], we abide by its [code of conduct][cncf-conduct].

[operator]: https://github.com/triggermesh/aws-sources-operator/blob/master/README.md
[tm-docs]: https://docs.triggermesh.io/sources/

[knsrc]: https://github.com/triggermesh/knative-sources
[knsrc-gitlab]: https://github.com/knative-sandbox/eventing-gitlab

[gh-issue]: https://github.com/triggermesh/knative-sources/issues
[gh-pr]: https://github.com/triggermesh/knative-sources/pulls

[cncf]: https://www.cncf.io/
[cncf-conduct]: https://github.com/cncf/foundation/blob/master/code-of-conduct.md
