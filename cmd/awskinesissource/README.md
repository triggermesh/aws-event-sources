# Amazon Kinesis event source for Knative Eventing

This event source consumes records from a Amazon Kinesis stream and sends them as CloudEvents to an arbitrary event
sink.

## Contents

1. [Prerequisites](#prerequisites)
1. [Deployment to Kubernetes](#deployment-to-kubernetes)

## Prerequisites

* Register an AWS account
* Create an [Access Key][doc-accesskey] in your AWS IAM dashboard.
* Create a [Kinesis stream][doc-kinesis].

## Deployment to Kubernetes

The _Amazon Kinesis event source_ can be deployed to Kubernetes as an `AWSKinesisSource` object, to a cluster where the
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

Copy the sample manifest from `config/samples/awskinesissource.yaml` and replace the pre-filled `spec` attributes with
the values corresponding to your _Amazon Kinesis_ stream. Then, create that `AWSKinesisSource` object in your Kubernetes
cluster:

```console
$ kubectl -n <my_namespace> create -f my-awskinesissource.yaml
```

[doc-accesskey]: https://docs.aws.amazon.com/general/latest/gr/aws-sec-cred-types.html#access-keys-and-secret-access-keys
[doc-kinesis]: https://docs.aws.amazon.com/streams/latest/dev/amazon-kinesis-streams.html
