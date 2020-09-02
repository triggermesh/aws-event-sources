# End-to-end test suite

Uses the [Kubernetes E2E framework](https://godoc.org/k8s.io/kubernetes/test/e2e/framework)

## Usage examples

Display all available test flags and exit:

```console
go test ./ -args -h
```

Run locally, against the default cluster (context) configured in the referenced kubeconfig file. Equivalent to exporting
the **KUBECONFIG** environment variable:

```console
go test ./ -args -kubeconfig ${HOME}/.kube/config
```

Enable Ginkgo's verbose mode, which prints log outputs immediately regardless of the outcome of tests:

```console
go test ./ -args -ginkgo.v
```
