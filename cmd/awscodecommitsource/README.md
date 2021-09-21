# AWS CodeCommit event source for Knative Eventing

This event source consumes messages from a AWS CodeCommit repository and sends them as CloudEvents to an arbitrary event
sink.

## Contents

1. [Prerequisites](#prerequisites)
1. [Deployment to Kubernetes](#deployment-to-kubernetes)

## Prerequisites

* Register an AWS account
* Create an [Access Key][doc-accesskey] in your AWS IAM dashboard.
* Create a [CodeCommit repository][doc-codecommit].

## Deployment to Kubernetes

The _AWS CodeCommit event source_ can be deployed to Kubernetes as an `AWSCodeCommitSource` object, to a cluster where
the TriggerMesh _AWS Event Sources Controller_ is running.

> :information_source: The sample manifest below references AWS credentials (Access Key) from a Kubernetes Secret object
> called `awscreds`. This Secret can be generated with the following command:
>
> ```console
> $ kubectl -n <my_namespace> create secret generic awscreds \
>   --from-literal=aws_access_key_id=<my_key_id> \
>   --from-literal=aws_secret_access_key=<my_secret_key>
> ```
>
> Alternatively, credentials can be used as literal strings instead of references to Kubernetes Secrets by replacing
> `valueFrom` attributes with `value` inside API objects' manifests.

Copy the sample manifest from `config/samples/awscodecommitsource.yaml` and replace the pre-filled `spec` attributes
with the values corresponding to your _AWS CodeCommit_ repository. Then, create that `AWSCodeCommitSource` object in
your Kubernetes cluster:

```console
$ kubectl -n <my_namespace> create -f my-awscodecommitsource.yaml
```

[doc-accesskey]: https://docs.aws.amazon.com/general/latest/gr/aws-sec-cred-types.html#access-keys-and-secret-access-keys
[doc-codecommit]: https://docs.aws.amazon.com/codecommit/latest/userguide/how-to-create-repository.html
