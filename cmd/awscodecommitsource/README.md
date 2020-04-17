# AWS CodeCommit event source for Knative Eventing

This event source consumes messages from a AWS CodeCommit repository and sends them as CloudEvents to an arbitrary event
sink.

## Contents

1. [Prerequisites](#prerequisites)
1. [Deployment to Kubernetes](#deployment-to-kubernetes)
   * [As a AWSCodeCommitSource object](#as-a-awscodecommitsource-object)
   * [As a ContainerSource object](#as-a-containersource-object)
1. [Running locally](#running-locally)
   * [In the shell](#in-the-shell)
   * [In a Docker container](#in-a-docker-container)

## Prerequisites

* Register an AWS account
* Create an [Access Key][doc-accesskey] in your AWS IAM dashboard.
* Create a [CodeCommit repository][doc-codecommit].

## Deployment to Kubernetes

The _AWS CodeCommit event source_ can be deployed to Kubernetes in different manners:

* As an `AWSCodeCommitSource` object, to a cluster where the TriggerMesh _AWS Sources Controller_ is running.
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

### As a AWSCodeCommitSource object

Copy the sample manifest from `config/samples/awscodecommitsource.yaml` and replace the pre-filled `spec` attributes
with the values corresponding to your _AWS CodeCommit_ repository. Then, create that `AWSCodeCommitSource` object in
your Kubernetes cluster:

```console
$ kubectl -n <my_namespace> create -f my-awscodecommitsource.yaml
```

### As a ContainerSource object

Copy the sample manifest from `config/samples/awscodecommit-containersource.yaml` and replace the pre-filled environment
variables under `env` with the values corresponding to your _AWS CodeCommit_ repository. Then, create that
`ContainerSource` object in your Kubernetes cluster:

```console
$ kubectl -n <my_namespace> create -f my-awscodecommit-containersource.yaml
```

## Running locally

Running the event source on your local machine can be convenient for development purposes.

### In the shell

Ensure the following environment variables are exported to your current shell's environment:

```sh
export REPO=<my_codecommit_repo>
export BRANCH=<my_git_branch>
export EVENT_TYPES=push,pull_request
export AWS_REGION=<my_repo_region>
export AWS_ACCESS_KEY_ID=<my_key_id>
export AWS_SECRET_ACCESS_KEY=<my_secret_key>
export NAME=my-awscodecommitsource
export NAMESPACE=default
export K_LOGGING_CONFIG='{"level":"info"}'
export K_METRICS_CONFIG='{"domain":"triggermesh.io/sources", "component":"awscodecommitsource", "configMap":{}}'
```

Then, run the event source with:

```console
$ go run ./cmd/awscodecommitsource
```

### In a Docker container

Using one of TriggerMesh's release images:

```console
$ docker run --rm \
  -e REPO=<my_codecommit_repo> \
  -e BRANCH=<my_git_branch> \
  -e EVENT_TYPES=push,pull_request \
  -e AWS_REGION=<my_repo_region> \
  -e AWS_ACCESS_KEY_ID=<my_key_id> \
  -e AWS_SECRET_ACCESS_KEY=<my_secret_key> \
  -e NAME=my-awscodecommitsource \
  -e NAMESPACE=default \
  -e K_LOGGING_CONFIG='{"level":"info"}' \
  -e K_METRICS_CONFIG='{"domain":"triggermesh.io/sources", "component":"awscodecommitsource", "configMap":{}}' \
  gcr.io/triggermesh/awscodecommitsource:latest
```

[doc-accesskey]: https://docs.aws.amazon.com/general/latest/gr/aws-sec-cred-types.html#access-keys-and-secret-access-keys
[doc-codecommit]: https://docs.aws.amazon.com/codecommit/latest/userguide/how-to-create-repository.html
