## AWS Kinesis producer for populating kinesis stream with testing data

This event producer is meant to be used as a producer of messages to Kinesis stream to test working source code.

### Build

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

Then just run the local binary in your shell to send messages to a defined kinesis stream