### SNS Source

This source expects external-dns of some flavor to be installed in the k8s cluster. If you look at the deployment.yaml file, you'll see that a service is created that has a hostname annotation. This hostname must be of the form `sns.$TOPIC_NAME.$CHANNEL_NAME.$NAMESPACE.$DOMAIN`. Domain can be anything you like. I used a subdomain in my personal env `sources.rsmitty.cloud`. These same vars are passed into the SNS deployment as env variables.

Upon starting, a background task is launched before running the webserver. This background task attempts to subscribe to the SNS topic mentioned. It will wait until the hostname above resolves in DNS before attempting to do so. Once subscription has occurred, events are pushed like other sources.

Usage:

- Define AWS creds file like

```
[default]
aws_access_key_id=xxx
aws_secret_access_key=yyy
```

- Create secret called snscreds with the creds file:

```
 kubectl create secret generic snscreds --from-file=/path/to/.aws/credentials
```

- Create deployment with `kubectl create -f deployment.yaml`

TODO: 
  - Support HTTPS
  - Marshal up notification data better