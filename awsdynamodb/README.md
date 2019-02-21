## AWS Dynamo DB Stream Event source for knative eventing

This event source is meant to be used as a Container Source with a Knative cluster to consume messages from a AWS DynamoDB Stream and send them to a Knative service/function.

### Local build

```
dep ensure -v
go build .
```

### Local Usage

- Register AWS account
- Get your account credentials. Navigate to "My Security Credentials" (tab)[https://console.aws.amazon.com/iam/home#/security_credential] in account and select "Access keys (access key ID and secret access key)" section to view your credentials
- Create table in Dynamo DB (with Stream Option - Enabled)
- Save Region and Table Name values

Define a few environment variables:

```
export TABLE=default
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=<>
export AWS_SECRET_ACCESS_KEY=<>
```

Then just run the local binary in your shell, open up DynamoDB table and insert record to the table via the AWS DynamoDB console.

```
$ ./awsdynamodb 
INFO[0000] Begin listening for Dynamo DB Stream         
INFO[0020] Processing record ID: 0xc0004463d0 
...
```

### Local Docker Usage

If you don't have a local Go environment, use Docker:

```
docker run -ti -e TABLE="your_table_name" \
               -e AWS_REGION="us-east-1" \
               -e AWS_ACCESS_KEY_ID="fgfdgsdfg" \
               -e AWS_SECRET_ACCESS_KEY="dsgdgsfgsfdgdsf" \
               gcr.io/triggermesh/awsdynamodb:latest
```

### Knative usage

Create secret called awscreds with the creds file:

```
 kubectl create secret generic awscreds --from-literal=aws_access_key_id=<replace_with_key> \
                                        --from-literal=aws_secret_access_key=<replace_with_key> \
```

Edit the Container source manifest and apply it:

```
kubectl apply -f dynamodb-source.yaml
```
