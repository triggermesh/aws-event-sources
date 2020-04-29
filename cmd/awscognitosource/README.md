# AWS Cognito event source for Knative Eventing

This event source consumes messages from a AWS Cognito identity pool and sends them as CloudEvents to an arbitrary event
sink.

## Contents

1. [Prerequisites](#prerequisites)
1. [Deployment to Kubernetes](#deployment-to-kubernetes)
   * [As a AWSCognitoSource object](#as-a-awscognitosource-object)
   * [As a ContainerSource object](#as-a-containersource-object)
   * [As a Deployment object bound by a SinkBinding](#as-a-deployment-object-bound-by-a-sinkbinding)
1. [Running locally](#running-locally)
   * [In the shell](#in-the-shell)
   * [In a Docker container](#in-a-docker-container)

## Prerequisites

* Register an AWS account
* Create an [Access Key][doc-accesskey] in your AWS IAM dashboard.
* Create a [Cognito identity pool][doc-cognito].

## Deployment to Kubernetes

The _AWS Cognito event source_ can be deployed to Kubernetes in different manners:

* As an `AWSCognitoSource` object, to a cluster where the TriggerMesh _AWS Sources Controller_ is running.
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

### As a AWSCognitoSource object

Copy the sample manifest from `config/samples/awscognitosource.yaml` and replace the pre-filled `spec` attributes
with the values corresponding to your _AWS Cognito_ identity pool. Then, create that `AWSCognitoSource` object in
your Kubernetes cluster:

```console
$ kubectl -n <my_namespace> create -f my-awscognitosource.yaml
```

### As a ContainerSource object

Copy the sample manifest from `config/samples/awscognito-containersource.yaml` and replace the pre-filled environment
variables under `env` with the values corresponding to your _AWS Cognito_ identity pool. Then, create that
`ContainerSource` object in your Kubernetes cluster:

```console
$ kubectl -n <my_namespace> create -f my-awscognito-containersource.yaml
```

### As a Deployment object bound by a SinkBinding

Copy the sample manifest from `config/samples/awscognito-sinkbinding.yaml` and replace the pre-filled environment
variables under `env` with the values corresponding to your _AWS Cognito_ identity pool. Then, create the `Deployment`
and `SinkBinding` objects in your Kubernetes cluster:

```console
$ kubectl -n <my_namespace> create -f my-awscognito-sinkbinding.yaml
```

## Running locally

Running the event source on your local machine can be convenient for development purposes.

### In the shell

Ensure the following environment variables are exported to your current shell's environment:

```sh
export IDENTITY_POOL_ID=<my_identity_pool_id>
export AWS_ACCESS_KEY_ID=<my_key_id>
export AWS_SECRET_ACCESS_KEY=<my_secret_key>
export NAME=my-awscognitosource
export NAMESPACE=default
export K_LOGGING_CONFIG=''
export K_METRICS_CONFIG='{"domain":"triggermesh.io/sources", "component":"awscognitosource", "configMap":{}}'
```

Then, run the event source with:

```console
$ go run ./cmd/awscognitosource
```

### In a Docker container

Using one of TriggerMesh's release images:

```console
$ docker run --rm \
  -e IDENTITY_POOL_ID=<my_identity_pool_id> \
  -e AWS_ACCESS_KEY_ID=<my_key_id> \
  -e AWS_SECRET_ACCESS_KEY=<my_secret_key> \
  -e NAME=my-awscognitosource \
  -e NAMESPACE=default \
  -e K_LOGGING_CONFIG='' \
  -e K_METRICS_CONFIG='{"domain":"triggermesh.io/sources", "component":"awscognitosource", "configMap":{}}' \
  gcr.io/triggermesh/awscognitosource:latest
```

[doc-accesskey]: https://docs.aws.amazon.com/general/latest/gr/aws-sec-cred-types.html#access-keys-and-secret-access-keys
[doc-cognito]: https://docs.aws.amazon.com/cognito/latest/developerguide/tutorial-create-identity-pool.html
