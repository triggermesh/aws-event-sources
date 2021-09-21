# Amazon SNS event source for Knative Eventing

This event source subscribes to notifications from a Amazon SNS topic and sends them as CloudEvents to an arbitrary
event sink.

Each instance of the SNS source is backed by a Knative Service that exposes a unique public HTTP(S) endpoint. This
endpoint is used to subscribe to the desired SNS topic on behalf of the user.

## Contents

1. [Prerequisites](#prerequisites)
1. [Deployment to Kubernetes](#deployment-to-kubernetes)

## Prerequisites

* Register an AWS account
* Create an [Access Key][doc-accesskey] in your AWS IAM dashboard.
* Create a Standard [SNS topic][doc-sns].

## Deployment to Kubernetes

The _Amazon SNS event source_ can be deployed to Kubernetes as an `AWSSNSSource` object, to a cluster where the
TriggerMesh _AWS Event Sources Controller_ is running.

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

Copy the sample manifest from `config/samples/awssnssource.yaml` and replace the pre-filled `spec` attributes with the
values corresponding to your _Amazon SNS_ topic. Then, create that `AWSSNSSource` object in your Kubernetes cluster:

```console
$ kubectl -n <my_namespace> create -f my-awssnssource.yaml
```

[doc-accesskey]: https://docs.aws.amazon.com/general/latest/gr/aws-sec-cred-types.html#access-keys-and-secret-access-keys
[doc-sns]: https://docs.aws.amazon.com/sns/latest/dg/sns-getting-started.html
