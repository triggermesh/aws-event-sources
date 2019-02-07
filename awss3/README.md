## AWS S3 Event source for knative eventing

This event source is meant to be used as a Container Source with a Knative cluster to consume events from an AWS S3 bucket via a SQS queue. As such the user needs to configure the S3 to SQS connection in the AWS console and then use the TriggerMesh SQS source.

### Setup

In SQS Service: 
1. Create a queue to recieve events from your S3 service.
2. Select created queue and navigate to `permissions` tab. Add a permission to send messages to your queue from your S3 service with your aws account number. If you do not know this number, you can allow `Everybody`, but use it for testing only purposes. 

You now have properly configured queue to get messages from S3. 

In S3 Service: 
1. Create a bucket in S3 service
2. Enter your bucket and select `Properties` tab. 
3. Scroll down to `Advanced settings` and find `Events` there you can add notification for events that happen in your Bucket. 
4. Configure which Events you would like to track and in the section `Send to` select `SQS Queue`. Select the Queue configured to get S3 events to send events to. 

### Local build

```
dep ensure -v
go build .
```

### Local Usage

- Register AWS account
- Get your account credentials. Navigate to "My Security Credentials" (tab)[https://console.aws.amazon.com/iam/home#/security_credential] in account and select "Access keys (access key ID and secret access key)" section to view your credentials


Define a few environment variables:

```
export QUEUE=your_queue_name
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=<>
export AWS_SECRET_ACCESS_KEY=<>
```

Then just run the local binary in your shell, open up 

```
$ ./awss3
...
```

### Local Docker build

If you don't have a local Go environment, use Docker:

```
docker run -ti -e QUEUE="your_queue_name" \
               -e AWS_REGION="us-east-1" \
               -e AWS_ACCESS_KEY_ID="fgfdgsdfg" \
               -e AWS_SECRET_ACCESS_KEY="dsgdgsfgsfdgdsf" \
               gcr.io/triggermesh/awss3:latest
```

### Knative usage

Create secret called awscreds with the creds file:

```
 kubectl create secret generic awscreds --from-literal=aws_access_key_id=<replace_with_key> \
                                        --from-literal=aws_secret_access_key=<replace_with_key> \
```

Edit the Container source manifest and apply it:

```
kubectl apply -f awss3-source.yaml
```
