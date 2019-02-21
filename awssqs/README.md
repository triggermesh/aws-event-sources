## AWS SQS Event source for knative eventing

This event source is meant to be used as a Container Source with a Knative cluster to consume messages from a AWS SQS queue and send them to a Knative service/function.

### Local build

```
dep ensure -v
go build .
```

### Local Usage

Define a few environment variables:

```
export QUEUE=default
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=<>
export AWS_SECRET_ACCESS_KEY=<>
```

Then just run the local binary in your shell and send messages via the AWS SQS console.

```
$ ./awssqs 
INFO[0000] Beginning to listen at URL: https://sqs.us-east-1.amazonaws.com/587264368683/triggermesh 
INFO[0025] Processing message with ID: 7168a015-09fa-4802-bca9-63e3df66753a 
INFO[0025] {
...
```

### Local Docker Usage

If you don't have a local Go environment, use Docker:

```
docker run -ti -e QUEUE="queue_name" \
               -e AWS_REGION="us-east-1" \
               -e AWS_ACCESS_KEY_ID="fgfdgsdfg" \
               -e AWS_SECRET_ACCESS_KEY="dsgdgsfgsfdgdsf" \
               gcr.io/triggermesh/sqssource:latest
```

### Knative usage

Create secret called awscreds with the creds file:

```
 kubectl create secret generic awscreds --from-literal=aws_access_key_id=<replace_with_key> \
                                        --from-literal=aws_secret_access_key=<replace_with_key> \
```

Edit the Container source manifest and apply it:

```
kubectl apply -f sqs-source.yaml
```
