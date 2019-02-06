## AWS S3 Event source for knative eventing

This event source is meant to be used as a Container Source with a Knative cluster to consume messages from a AWS S3 Event Stream and send them to a Knative service/function.

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
export AWS_BUCKET=your_bucket_name
export AWS_OBJECT_KEY=your_object_key
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
docker run -ti -e AWS_BUCKET="your_bucket_name" \
               -e AWS_OBJECT_KEY="your_object_key" \
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
