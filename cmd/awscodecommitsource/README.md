## AWS CodeCommit Events source for knative eventing

This event source is meant to be used as a Container Source with a Knative cluster to consume messages from a AWS CodeCommit events and send them to a Knative service/function.

### Local build

```
go build .
```

### Local Usage

- Register AWS account
- Get your account credentials. Navigate to "My Security Credentials" (tab)[https://console.aws.amazon.com/iam/home#/security_credential] in account and select "Access keys (access key ID and secret access key)" section to view your credentials
- Create repo in CodeCommit in Developer Tools and create simple file.txt so that it would create a master branch. Then create another branch from master, add file there and make PR to master

Define a few environment variables:

```
export REPO=triggermeshtest
export BRANCH=master
export EVENTS=pull_request,push
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=<>
export AWS_SECRET_ACCESS_KEY=<>
```

Then just run the local binary in your shell and send PR through Developer Tools.

```
$ ./awscodecommitsource
```

### Local Docker Usage

If you don't have a local Go environment, use Docker:

```
docker run -ti -e REPO="your_repo_name" \
               -e BRANCH="your_branch_name" \
               -e EVENTS="pull_request,push" \
               -e AWS_REGION="us-east-1" \
               -e AWS_ACCESS_KEY_ID="fgfdgsdfg" \
               -e AWS_SECRET_ACCESS_KEY="dsgdgsfgsfdgdsf" \
               gcr.io/triggermesh/awscodecommit:latest
```

### Knative usage

Create secret called awscreds with the creds file:

```
kubectl create secret generic awscreds --from-literal=aws_access_key_id=<replace_with_key> \
                                        --from-literal=aws_secret_access_key=<replace_with_key> \
```

Edit the Container source manifest and apply it:

```
kubectl apply -f codecommit-source.yaml
```
