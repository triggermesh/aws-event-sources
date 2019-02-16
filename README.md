![TriggerMesh Knative Lambda Sources](./images/klass.png "TriggerMesh Knative Lambda Sources")

**What:** Knative Lambda Sources (KLASS) are Knative event sources for AWS services.

**Why:** You maybe be using some Cloud services on AWS but still interested to run workloads within Kubernetes and soon via [Knative](https://github.com/knative/docs) to benefit from features such as scale to zero and source-to-url FaaS functionality. To trigger those workloads when events happen in your AWS service you need to have an event source that can consume AWS events and send them to your workload. This is a key principle in Knative eventing.

**How:** The sources listed in this repo are fully open source and can be used in any Knative cluster. They consist of Go event consumers for various AWS services. Most of them are packaged as `Container Sources` and make use of [CloudEvents](https://cloudevents.io/)

## Sources and Usage

Most sources have the following structure:

```shell
├── awscodecommit
│   ├── Dockerfile
│   ├── Gopkg.lock
│   ├── Gopkg.toml
│   ├── Makefile
│   ├── README.md
│   ├── codecommit-source.yaml
│   ├── main.go
│   └── main_test.go
```

The code is in `main.go`. The `Dockerfile` shows how the source is containerized and the `*-source.yaml` is the `ContainerSource` manifest that you can deploy on your knative cluster.

For example, the following manifest will start the _CodeCommit_ source.

```
apiVersion: sources.eventing.knative.dev/v1alpha1
kind: ContainerSource
metadata:
  name: awscodecommit
spec:
  image: gcr.io/triggermesh/awscodecommit:latest
  sink:
    apiVersion: eventing.knative.dev/v1alpha1
    kind: Channel
    name: default
  env:
    - name: AWS_ACCESS_KEY_ID
      valueFrom:
        secretKeyRef:
          name: awscreds
          key: aws_access_key_id
    - name: AWS_SECRET_ACCESS_KEY
      valueFrom:
        secretKeyRef:
          name: awscreds
          key: aws_secret_access_key
    - name: AWS_REGION
      value: us-west-2
    - name: REPO
      value: triggermeshtest
    - name: BRANCH
      value: master
    - name: EVENTS
      value: pull_request,push
```

Given the above manifest stored as a file named `codecommit-source.yaml`. You can start the source via:

```
kubectl apply -f codecommit-source.yaml
```

For information on what is a knative [Channel](https://github.com/knative/docs/tree/master/eventing) please see the Knative [documentation](https://github.com/knative/docs/tree/master/eventing).

### Credentials

In the example manifests provided in this repo, the AWS credentials are loaded via a Kubernetes secret named `awscreds` which needs to contain the keys `aws_access_key_id` and `aws_secret_access_key`.

## List

| AWS service | Source Type | Support Level|
|-------------|-------------|--------|
|CodeCommit|Container source|alpha|
|Cognito|Container source|alpha|
|DynamoDB|Container source|alpha|
|IoT|Container source|alpha|
|Kinesis|Container source|alpha|
|S3|Container source|alpha|
|SNS|Container source|alpha|
|SQS|Container source|alpha|

## Caveat

AWS Events are very rich. AWS SNS and AWS CloudWatch can be used with almost every AWS services, hence there are many different ways to consume and/or receive AWS events. These sources represent one way of doing it.

## TriggerMesh Cloud Early Access

These container sources can be used as is from this repo. You can also use them from our Cloud [https://cloud.triggermesh.io](https://cloud.triggermesh.io) where we have developed an enjoyable UI to configure them. Check out this snapshot:

![TM cloud sources](./images/sources.png)

## Roadmap

* Add a more generic SNS source using an operator architecture
* Add a CloudWatch source using an operator architecture
* Use goroutines to make the sources more performant

## Support

We would love your feedback and help on these sources, so don't hesitate to let us know what is wrong and how we could improve them, just file an [issue](https://github.com/triggermesh/knative-lambda-sources/issues/new) or join those of use who are maintaining them and submit a [PR](https://github.com/triggermesh/knative-lambda-sources/compare)

## Code of Conduct

This plugin is by no means part of [CNCF](https://www.cncf.io/) but we abide by its [code of conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md)
