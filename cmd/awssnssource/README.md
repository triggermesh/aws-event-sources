# AWS SNS event source for Knative Eventing

This event source subscribes to messages from a AWS SNS topic and sends them as CloudEvents to an arbitrary event sink.

<!--
This source expects external-dns of some flavor (https://github.com/kubernetes-sigs/external-dns) to be installed in the
k8s cluster. If you look at the deployment.yaml file, you'll see that a service is created that has a hostname
annotation. This hostname must be of the form `sns.$TOPIC_NAME.$CHANNEL_NAME.$NAMESPACE.$DOMAIN`. Domain can be anything
you like. I used a subdomain in my personal env `sources.rsmitty.cloud`. These same vars are passed into the SNS
deployment as env variables.

Upon starting, a background task is launched before running the webserver. This background task attempts to subscribe to
the SNS topic mentioned. It will wait until the hostname above resolves in DNS before attempting to do so. Once
subscription has occurred, events are pushed like other sources.
-->

## Contents

1. [Prerequisites](#prerequisites)
1. [Deployment to Kubernetes](#deployment-to-kubernetes)
   * [As a AWSSNSSource object](#as-a-awssnssource-object)
   * [As a ContainerSource object](#as-a-containersource-object)
   * [As a Deployment object bound by a SinkBinding](#as-a-deployment-object-bound-by-a-sinkbinding)
1. [Running locally](#running-locally)
   * [In the shell](#in-the-shell)
   * [In a Docker container](#in-a-docker-container)

## Prerequisites

* Register an AWS account
* Create an [Access Key][doc-accesskey] in your AWS IAM dashboard.
* Create a [SNS topic][doc-sns].
* Create a [SNS subscription][doc-sns].

## Deployment to Kubernetes

The _AWS SNS event source_ can be deployed to Kubernetes in different manners:

* As an `AWSSNSSource` object, to a cluster where the TriggerMesh _AWS Sources Controller_ is running.
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

### As a AWSSNSSource object

Copy the sample manifest from `config/samples/awssnssource.yaml` and replace the pre-filled `spec` attributes with the
values corresponding to your _AWS SNS_ topic. Then, create that `AWSSNSSource` object in your Kubernetes cluster:

```console
$ kubectl -n <my_namespace> create -f my-awssnssource.yaml
```

### As a ContainerSource object

Copy the sample manifest from `config/samples/awssns-containersource.yaml` and replace the pre-filled environment
variables under `env` with the values corresponding to your _AWS SNS_ topic. Then, create that `ContainerSource` object
in your Kubernetes cluster:

```console
$ kubectl -n <my_namespace> create -f my-awssns-containersource.yaml
```

### As a Deployment object bound by a SinkBinding

Copy the sample manifest from `config/samples/awssns-sinkbinding.yaml` and replace the pre-filled environment variables
under `env` with the values corresponding to your _AWS SNS_ topic. Then, create the `Deployment` and `SinkBinding`
objects in your Kubernetes cluster:

```console
$ kubectl -n <my_namespace> create -f my-awssns-sinkbinding.yaml
```

## Running locally

Running the event source on your local machine can be convenient for development purposes.

### In the shell

Ensure the following environment variables are exported to your current shell's environment:

```sh
export TOPIC=<my_sns_topic>
export AWS_REGION=<my_topic_region>
export AWS_ACCESS_KEY_ID=<my_key_id>
export AWS_SECRET_ACCESS_KEY=<my_secret_key>
export NAME=my-awssnssource
export NAMESPACE=default
export K_LOGGING_CONFIG=''
export K_METRICS_CONFIG='{"domain":"triggermesh.io/sources", "component":"awssnssource", "configMap":{}}'
```

Then, run the event source with:

```console
$ go run ./cmd/awssnssource
```

### In a Docker container

Using one of TriggerMesh's release images:

```console
$ docker run --rm \
  -e TOPIC=<my_sns_topic> \
  -e AWS_REGION=<my_topic_region> \
  -e AWS_ACCESS_KEY_ID=<my_key_id> \
  -e AWS_SECRET_ACCESS_KEY=<my_secret_key> \
  -e NAME=my-awssnssource \
  -e NAMESPACE=default \
  -e K_LOGGING_CONFIG='' \
  -e K_METRICS_CONFIG='{"domain":"triggermesh.io/sources", "component":"awssnssource", "configMap":{}}' \
  gcr.io/triggermesh/awssnssource:latest
```

[doc-accesskey]: https://docs.aws.amazon.com/general/latest/gr/aws-sec-cred-types.html#access-keys-and-secret-access-keys
[doc-sns-topic]: https://docs.aws.amazon.com/sns/latest/dg/sns-getting-started.html
