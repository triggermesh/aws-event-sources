This document briefly describes the process of creating a new release of the aws-event-sources operator.

## Pre-requisites

- [operator-sdk](https://github.com/operator-framework/operator-sdk/)
- [operator-courier](https://github.com/operator-framework/operator-courier/)

## Assumptions

This document assumes we are preparing version `0.0.2` of the operator.

## Step 1: Update chart sources

The operator is a Helm-based operator created from the `aws-event-sources` chart. A copy of the chart exists in the `deploy/helm-charts` directory and has minor changes in accordance with https://redhat-connect.gitbook.io/certified-operator-guide/helm-operators/building-a-helm-operator/using-a-single-image-variable

The first step in the maintenance of the operator is to apply the upstream chart updates to the local copy in `deploy/helm-charts`.

## Step 2: Build the operator image

After updating the chart manifests, a new version of the operator should be released.

First update the `version` LABEL in the `build/Dockerfile`

Next, build the operator using the `operator-sdk` tool and push it to GCR:

```bash
operator-sdk build gcr.io/triggermesh/aws-event-sources-operator:v0.0.2
docker push gcr.io/triggermesh/aws-event-sources-operator:v0.0.2
```

### Step 3: Prepare the OpenShift operator update

Edit the `deploy/operator.yaml` and `deploy/crds/sources.triggermesh.com_v1alpha1_awseventsources_cr.yaml` files and update version of the operator image and the versions of the controller and adapter images.

### Step 4: Generate the new `ClusterServiceVersion`

Finally generate the `ClusterServiceVersion` YAML file for the operator using the following command:

```bash
operator-sdk generate csv --csv-version 0.0.2
```

### Step 5: Prepare the Operator metadata

Create the directory `bundle/0.0.2` and copy the `deploy/olm-catalog/aws-sources-operator/manifests/aws-sources-operator.clusterserviceversion.yaml` and `deploy/olm-catalog/aws-sources-operator/manifests/sources.triggermesh.com_awseventsources_crd.yaml` into this directory with the names `aws-sources-operator.v0.0.2.clusterserviceversion.yaml` and `aws-sources-operator.crd.yaml` respectively.

> **IMPORTANT**: Please update the version number in the file names

Next, edit the `bundle/aws-sources-operator.package.yaml` file and update the `currentCSV` field accordingly.

## Step 6: Verify the operator metadata

The metadata should pass the following:

```bash
operator-courier verify bundle/
operator-courier verify --ui_validate_io bundle/
```

## Step 7: Submit Operator metadata

In the final step, zip the contents of `bundle/` and submit the Operator metadata to https://connect.redhat.com/project/5061751/operator-metadata

## References

- [Building a Helm Operator](https://redhat-connect.gitbook.io/certified-operator-guide/helm-operators/building-a-helm-operator)
- [Operator Metadata](https://redhat-connect.gitbook.io/certified-operator-guide/ocp-deployment/operator-metadata)
