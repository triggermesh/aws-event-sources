## AWS Cognito Event source for knative eventing

This event source is meant to be used as a Container Source with a Knative cluster to consume messages from a AWS Cognito and send them to a Knative service/function.

### Local build

```
GO111MODULE=on go mod download
go build .
```

### Local Usage

Define a few environment variables:

```
export IDENTITY_POOL_ID=identity_pool
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=<>
export AWS_SECRET_ACCESS_KEY=<>
```

Then just run the local binary in your shell and send datasets via the AWS Cognito console. You need to have existing identities to do so.
```
$ ./awscognito 
INFO[0001] {
  AllowUnauthenticatedIdentities: true,
  IdentityPoolId: "us-west-2:b403deb6-c49d-477f-bfc8-ade279e15af2",
  IdentityPoolName: "triggermesh"
} 

```

### Local Docker Usage

If you don't have a local Go environment, use Docker:

```
docker run -ti -e IDENTITY_POOL_ID="your_identity_pool_id" \
               -e AWS_REGION="us-east-1" \
               -e AWS_ACCESS_KEY_ID="fgfdgsdfg" \
               -e AWS_SECRET_ACCESS_KEY="dsgdgsfgsfdgdsf" \
               gcr.io/triggermesh/awscognito:latest
```

### Knative usage

Create secret called awscreds with the creds file:

```
kubectl create secret generic awscreds --from-literal=aws_access_key_id=<replace_with_key> \
                                        --from-literal=aws_secret_access_key=<replace_with_key> \
```

Edit the Container source manifest and apply it:

```
kubectl apply -f cognito-source.yaml
```
