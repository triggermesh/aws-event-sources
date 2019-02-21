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
