# aws-event-sources

Installs TriggerMesh AWS event sources controller. The following event sources are currently supported by the controller

- Amazon CodeCommit (`kind: AWSCodeCommitSource`)
- Amazon Cognito Identity (`kind: AWSCognitoIdentitySource`)
- Amazon Cognito UserPool (`kind: AWSCognitoUserPoolSource`)
- Amazon DynamoDB (`kind: AWSDynamoDBSource`)
- Amazon Kinesis (`kind: AWSKinesisSource`)
- Amazon Simple Notification Service (`kind: AWSSNSSource`)
- Amazon Simple Queue Service (`kind: AWSSQSSource`)

Refer to [aws-event-sources/config/samples](https://github.com/triggermesh/aws-event-sources/tree/master/config/samples) for examples that make use of the controller.

## TL;DR;

```console
$ helm repo add triggermesh https://storage.googleapis.com/triggermesh-charts
$ helm install triggermesh/aws-event-sources
```

To report bugs and for feedback and support please [create a new issue](https://github.com/triggermesh/aws-event-sources/issues/new).

## Introduction

This chart installs the [aws-event-sources](https://github.com/triggermesh/aws-event-sources) controller on a Kubernetes cluster.

## Prerequisites
  - Kubernetes 1.16+ with Beta APIs
  - Helm 3.0+
  - Knative v0.14+

## Installing the Chart

Add the TriggerMesh chart repository to Helm:

```console
$ helm repo add triggermesh https://storage.googleapis.com/triggermesh-charts
```

To install the chart with the release name `my-release`:

```console
$ helm install --name my-release triggermesh/aws-event-sources
```

The command deploys the aws-event-sources controller in the default configuration. Refer to the [configuration](#configuration) section for the complete list of parameters that can be specified to customize the deployment of the controller.

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```console
$ helm delete my-release
```

The Kubernetes resources associated with chart will be removed and the Helm release will be deleted.

However note that the `CustomResourceDefinition` resources created by the chart will not be removed by the delete command and therefore need to be removed manually.

```console
$ kubectl delete crd awscodecommitsources.sources.triggermesh.io
$ kubectl delete crd awscognitoidentitysources.sources.triggermesh.io
$ kubectl delete crd awscognitouserpoolsources.sources.triggermesh.io
$ kubectl delete crd awsdynamodbsources.sources.triggermesh.io
$ kubectl delete crd awsiotsources.sources.triggermesh.io
$ kubectl delete crd awskinesissources.sources.triggermesh.io
$ kubectl delete crd awssnssources.sources.triggermesh.io
$ kubectl delete crd awssqssources.sources.triggermesh.io
```

## Configuration

|                  Parameter                  |                     Description                     |                  Default                   |
|---------------------------------------------|-----------------------------------------------------|--------------------------------------------|
| `nameOverride`                              | Override the name for controller resources          | `""`                                       |
| `fullnameOverride`                          | Override the fullname for controller resources      | `""`                                       |
| `rbac.create`                               | Create RBAC resources                               | `true`                                     |
| `serviceAccount.create`                     | Create service account for the controller           | `true`                                     |
| `serviceAccount.annotations`                | Annotations to add to controller service account    | `{}`                                       |
| `serviceAccount.name`                       | Override the name for the service account           | `nil`                                      |
| `imagePullSecrets`                          | Specify image pull secrets                          | `[]`                                       |
| `image.registry`                            | Image registry name                                 | `gcr.io`                                   |
| `image.repository`                          | Image repository name                               | `triggermesh/aws-event-sources-controller` |
| `image.tag`                                 | Image tag                                           | `{TAG_NAME}`                               |
| `image.pullPolicy`                          | Image pull policy                                   | `IfNotPresent`                             |
| `adapter.awscloudwatch.repository`          | AWS adapter Cloudwatch image name                   | `triggermesh/awscloudwatchsource`          |
| `adapter.awscloudwatch.tag`                 | AWS adapter Cloudwatch image tag                    | `_defaults to value of `.image.tag`_`      |
| `adapter.awscloudwatchlogs.repository`      | AWS adapter Cloudwatch Logs image name              | `triggermesh/awscloudwatchlogssource`      |
| `adapter.awscloudwatchlogs.tag`             | AWS adapter Cloudwatch Logs image tag               | `_defaults to value of `.image.tag`_`      |
| `adapter.awscodecommit.repository`          | AWS adapter CodeCommit image name                   | `triggermesh/awscodecommitsource`          |
| `adapter.awscodecommit.tag`                 | AWS adapter CodeCommit image tag                    | `_defaults to value of `.image.tag`_`      |
| `adapter.awscognitoidentity.repository`     | AWS adapter Cognito Identity image name             | `triggermesh/awscognitoidentitysource`     |
| `adapter.awscognitoidentity.tag`            | AWS adapter Cognito Identity image tag              | `_defaults to value of `.image.tag`_`      |
| `adapter.awscognitouserpool.repository`     | AWS adapter Cognito Userpool image name             | `triggermesh/awscognitouserpoolsource`     |
| `adapter.awscognitouserpool.tag`            | AWS adapter Cognito Userpool image tag              | `_defaults to value of `.image.tag`_`      |
| `adapter.awsdynamodb.repository`            | AWS adapter DynamoDB image name                     | `triggermesh/awsdynamodbsource`            |
| `adapter.awsdynamodb.tag`                   | AWS adapter DynamoDB image tag                      | `_defaults to value of `.image.tag`_`      |
| `adapter.awskinesis.repository`             | AWS adapter Kinesis image name                      | `triggermesh/awskinesissource`             |
| `adapter.awskinesis.tag`                    | AWS adapter Kinesis image tag                       | `_defaults to value of `.image.tag`_`      |
| `adapter.awsperformanceinsights.repository` | AWS adapter Performance Insights image name         | `triggermesh/awsperformanceinsightssource` |
| `adapter.awsperformanceinsights.tag`        | AWS adapter Performance Insights image tag          | `_defaults to value of `.image.tag`_`      |
| `adapter.awssns.repository`                 | AWS adapter SNS image name                          | `triggermesh/awssnssource`                 |
| `adapter.awssns.tag`                        | AWS adapter SNS image tag                           | `_defaults to value of `.image.tag`_`      |
| `adapter.awssqs.repository`                 | AWS adapter SQS image name                          | `triggermesh/awssqssource`                 |
| `adapter.awssqs.tag`                        | AWS adapter SQS image tag                           | `_defaults to value of `.image.tag`_`      |
| `podAnnotations`                            | Annotations to add to the controller pod            | `{}``                                      |
| `podSecurityContext`                        | Security context for controller pods                | `{}`                                       |
| `securityContext`                           | Security context for controller containers          | `{}`                                       |
| `resources`                                 | Resource requests/limits for the controller         | `{requests: {cpu: 20m, memory: 20Mi}}`     |
| `nodeSelector`                              | Controller node selector                            | `{}`                                       |
| `tolerations`                               | Tolerations for use with node taints                | `[]`                                       |
| `affinity`                                  | Assign custom affinity rules to the controller pods | `{}`                                       |
