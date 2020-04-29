# AWS SQS event source for Knative Eventing

This event source consumes messages from a AWS SQS queue and sends them as CloudEvents to an arbitrary event sink.

## Contents

1. [Prerequisites](#prerequisites)
1. [Deployment to Kubernetes](#deployment-to-kubernetes)
   * [As a AWSSQSSource object](#as-a-awssqssource-object)
   * [As a ContainerSource object](#as-a-containersource-object)
   * [As a Deployment object bound by a SinkBinding](#as-a-deployment-object-bound-by-a-sinkbinding)
1. [Running locally](#running-locally)
   * [In the shell](#in-the-shell)
   * [In a Docker container](#in-a-docker-container)

## Prerequisites

* Register an AWS account
* Create an [Access Key][doc-accesskey] in your AWS IAM dashboard.
* Create a [SQS queue][doc-sqs].

## Deployment to Kubernetes

The _AWS SQS event source_ can be deployed to Kubernetes in different manners:

* As an `AWSSQSSource` object, to a cluster where the TriggerMesh _AWS Sources Controller_ is running.
* As a Knative `ContainerSource`, to any cluster running Knative Eventing.

> :information_source: The sample manifests below reference AWS credentials (Access Key) from a Kubernetes Secret object
> called `awscreds`. This Secret can be generated with the following command:
>
> ```console
> $ kubectl -n <my_namespace> create secret generic awscreds \
>   --from-literal=aws_access_key_id=<my_key_id> \
>   --from-literal=aws_secret_access_key=<my_secret_key>
> ```
>
> Alternatively, credentials can be used as literal strings instead of references by replacing `valueFrom` attributes
> with `value`.

### As a AWSSQSSource object

Copy the sample manifest from `config/samples/awssqssource.yaml` and replace the pre-filled `spec` attributes with the
values corresponding to your _AWS SQS_ queue. Then, create that `AWSSQSSource` object in your Kubernetes cluster:

```console
$ kubectl -n <my_namespace> create -f my-awssqssource.yaml
```

### As a ContainerSource object

Copy the sample manifest from `config/samples/awssqs-containersource.yaml` and replace the pre-filled environment
variables under `env` with the values corresponding to your _AWS SQS_ queue. Then, create that `ContainerSource` object
in your Kubernetes cluster:

```console
$ kubectl -n <my_namespace> create -f my-awssqs-containersource.yaml
```

### As a Deployment object bound by a SinkBinding

Copy the sample manifest from `config/samples/awssqs-sinkbinding.yaml` and replace the pre-filled environment variables
under `env` with the values corresponding to your _AWS SQS_ queue. Then, create the `Deployment` and `SinkBinding`
objects in your Kubernetes cluster:

```console
$ kubectl -n <my_namespace> create -f my-awssqs-sinkbinding.yaml
```

## Running locally

Running the event source on your local machine can be convenient for development purposes.

### In the shell

Ensure the following environment variables are exported to your current shell's environment:

```sh
export QUEUE=<my_sqs_queue>
export AWS_REGION=<my_queue_region>
export AWS_ACCESS_KEY_ID=<my_key_id>
export AWS_SECRET_ACCESS_KEY=<my_secret_key>
export NAME=my-awssqssource
export NAMESPACE=default
export K_LOGGING_CONFIG=''
export K_METRICS_CONFIG='{"domain":"triggermesh.io/sources", "component":"awssqssource", "configMap":{}}'
```

Then, run the event source with:

```console
$ go run ./cmd/awssqssource
```

### In a Docker container

Using one of TriggerMesh's release images:

```console
$ docker run --rm \
  -e QUEUE=<my_sqs_queue> \
  -e AWS_REGION=<my_queue_region> \
  -e AWS_ACCESS_KEY_ID=<my_key_id> \
  -e AWS_SECRET_ACCESS_KEY=<my_secret_key> \
  -e NAME=my-awssqssource \
  -e NAMESPACE=default \
  -e K_LOGGING_CONFIG='' \
  -e K_METRICS_CONFIG='{"domain":"triggermesh.io/sources", "component":"awssqssource", "configMap":{}}' \
  gcr.io/triggermesh/awssqssource:latest
```

[doc-accesskey]: https://docs.aws.amazon.com/general/latest/gr/aws-sec-cred-types.html#access-keys-and-secret-access-keys
[doc-sqs]: https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-create-queue.html
