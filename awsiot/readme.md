## AWS IoT source for knative eventing

This event source is meant to be used as a Container Source with a Knative cluster to consume messages from a AWS IoT Topic and send them to a Knative service/function.

### Local Usage

- install `python 3` & `pip3`
- run ```pip3 install AWSIoTPythonSDK```
- copy the following files inside your project's folder: `Root CA`, `Certificate for the Thing`, `Private Key`. You can obtain them following the official AWS Guide for IoT [setup tutorial](https://docs.aws.amazon.com/iot/latest/developerguide/iot-gs.html)

copy your `AWS IoT custom endpoint`:
1. Open up your `Thing`
2. Look for `interact` item on the left pannel
3. Inside `interact` copy HTTPS endpoint for your Thing 


Define a few environment variables:
```
export THING_SHADOW_ENDPOINT=example-ats.iot.us-west-2.amazonaws.com
export ROOT_CA_PATH=rootca.crt
export CERTIFICATE_PATH=certificate.pem.crt
export PRIVATE_KEY_PATH=private.pem.key
export TOPIC=yourtopic
```

- run ``` python3 main.py ```

### Local Docker Usage

If you don't have a local Python environment, use Docker:

```
docker run -ti -e THING_SHADOW_ENDPOINT="example-ats.iot.us-west-2.amazonaws.com" \
               -e ROOT_CA_PATH="rootca.crt" \
               -e CERTIFICATE_PATH="certificate.pem.crt" \
               -e PRIVATE_KEY_PATH="private.pem.key" \
               -e TOPIC="triggertopic" \
               gcr.io/triggermesh/awsiot:latest
```

### Knative usage

Create secret called awsiot with all the needed files:

```
kubectl create secret generic awsiot --from-file=rootca.crt --from-file=certificate.pem.crt --from-file=private.pem.key
```

Edit the Container source manifest and apply it:

```
kubectl apply -f awsiot-source.yaml
```
