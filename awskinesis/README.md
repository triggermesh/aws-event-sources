## AWS Kinesis Event source for knative eventing

This event source is meant to be used as a Container Source with a Knative cluster to consume messages from a AWS Kinesis stream and send them to a Knative service/function.

### Local build

```
dep ensure -v
go build .
```

### Local Usage

Define a few environment variables:

```
export STREAM=default
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=<>
export AWS_SECRET_ACCESS_KEY=<>

```

Then just run the local binary in your shell and send messages via the AWS Kinesis producer from the separate folder.

### Local Docker Usage

If you don't have a local Go environment, use Docker:

```
docker run -ti -e STREAM="stream_name" \
               -e AWS_REGION="us-east-1" \
               -e AWS_ACCESS_KEY_ID="fgfdgsdfg" \
               -e AWS_SECRET_ACCESS_KEY="dsgdgsfgsfdgdsf" \
               gcr.io/triggermesh/awskinesis:latest
```

### Knative usage

Create secret called awscreds with the creds file:

```
 kubectl create secret generic awscreds --from-literal=aws_access_key_id=<replace_with_key> \
                                        --from-literal=aws_secret_access_key=<replace_with_key> \
```

Edit the Container source manifest and apply it:

```
kubectl apply -f kinesis-source.yaml
```
