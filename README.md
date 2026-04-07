# Constellix Webhook for Cert Manager

This is a webhook solver for [Constellix](https://constellix.com), for use with cert-manager,
to solve ACME DNS01 challenges.

**Supported environments:** Kubernetes **1.25+**, [cert-manager](https://cert-manager.io/docs/installation/kubernetes/) **1.14+** (tested against cert-manager **1.14**), [Helm](https://helm.sh/) **3.x**.

The container image is built from **Go 1.23** as a static binary and runs on [**distroless**](https://github.com/GoogleContainerTools/distroless) (`nonroot`).

## Prerequisites

- [Helm](https://helm.sh/) 3.x for installing the chart
- [cert-manager](https://cert-manager.io/docs/installation/kubernetes/) 1.14 or newer in the cluster

## Installation

1. Install the chart with helm

We have a [helm repo](https://constellix.github.io/cert-manager-webhook-constellix/) set up,
so you can use that, or you can install directly from source:

```bash
$ helm install --namespace cert-manager cert-manager-webhook-constellix ./deploy/cert-manager-webhook-constellix
```

2. Populate a secret with your Constellix API Key and Secret Key

```bash
$ kubectl --namespace cert-manager create secret generic constellix-credentials --from-literal=apiKey='Your Constellix API Key'
```

3. We need to grant permission for service account to get the secret. Copy the
following and apply with something like:

```bash
$ kubectl --namespace cert-manager apply -f secret_reader.yaml
```

Note that it may make more sense in your setup to use a `ClusterRole` and
`ClusterRoleBinding` here.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: cert-manager-webhook-constellix:secret-reader
rules:
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: ["constellix-credentials"]
  verbs: ["get", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: cert-manager-webhook-constellix:secret-reader
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: cert-manager-webhook-constellix:secret-reader
subjects:
  - apiGroup: ""
    kind: ServiceAccount
    name: cert-manager-webhook-constellix
```

## Configuration

You'll need to edit and apply some resources, with something like:

```bash
$ kubectl --namespace cert-manager apply -f my_resource.yaml
```

Note that we use the `cert-manager` namespace, but it may make more sense in
your setup to same more nuanced namespace management.

1. Create Issuer(s), we'll use `letsencrypt` for example. We'll use
`ClusterIssuer` here, which will be available accross namespaces. You may
prefer to use `Issuer`. This is where `Constellix` API options are set (`endpoint`,
`ignoreSSL`).

Staging issuer (**optional**):

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    # The ACME server URL
    server: https://acme-staging-v02.api.letsencrypt.org/directory

    # Email address used for ACME registration
    email: user@example.com # REPLACE THIS WITH YOUR EMAIL!!!

    # Name of a secret used to store the ACME account private key
    privateKeySecretRef:
      name: letsencrypt-staging

    solvers:
    - dns01:
        webhook:
          groupName: acme.constellix.com
          solverName: constellix
          config:
            apiKeySecretRef:
              key: apiKey
              name: constellix-credentials
            apiSecretSecretRef:
              key: secretKey
              name: constellix-credentials
            zoneId: 1234
            insecure: false
```

For **multiple Constellix domains** under one issuer, use `zones` (longest DNS name match wins) instead of a single `zoneId`:

```yaml
            zones:
              - dnsName: example.com
                zoneId: 111
              - dnsName: sub.example.com
                zoneId: 222
```

Production issuer:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    # The ACME server URL
    server: https://acme-v02.api.letsencrypt.org/directory

    # Email address used for ACME registration
    email: user@example.com # REPLACE THIS WITH YOUR EMAIL!!!

    # Name of a secret used to store the ACME account private key
    privateKeySecretRef:
      name: letsencrypt-prod

    solvers:
    - dns01:
        webhook:
          groupName: acme.constellix.com
          solverName: constellix
          config:
            apiKeySecretRef:
              key: apiKey
              name: constellix-credentials
            apiSecretSecretRef:
              key: secretKey
              name: constellix-credentials
            zoneId: 1234
            insecure: false
```

2. Test things by issuing a certificate. This example requests a cert for
`example.com` from the staging issuer, default namespace should be fine:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: example.com
spec:
  commonName: example.com
  dnsNames:
  - example.com
  issuerRef:
    name: letsencrypt-staging
  secretName: example-com-tls
```

After a minute or two, the `Certificate` should show as `Ready`. If not, you
can follow the resource chain from `Certificate` to `CertificateRequest` and on
down until you see a useful error message.

### Automatically creating Certificates for Ingress resources

See cert-manager
[docs](https://docs.cert-manager.io/en/latest/tasks/issuing-certificates/ingress-shim.html)
on "ingress shims".

The gist of it is adding an annotation, and a `tls` section to your Ingress
definition. A simple ingress example is below with pertinent areas bolded. We
use the `ingress-nginx` ingress controller, but it should be the same idea for
any ingress.

You do of course, need to set up an `A` Record in `Constellix` connecting the domain
to the external IP of the ingress controller's LoadBalancer service. In the
example below the domain would be `my-app.example.com`.

<pre>
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-ingress
  annotations:
    kubernetes.io/ingress.class: "nginx"
    <b>cert-manager.io/cluster-issuer: "letsencrypt-prod"</b>
spec:
  rules:
  - host: my-app.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: my-app-service
            port:
              number: 80
  <b>tls:
  - hosts:
    - my-app.example.com</b>
    <b>secretName: my-app-tls</b>
</pre>

### Troubleshooting

If things aren't working, check the logs in the main `cert-manager` pod first,
they are pretty communicative. Check logs from the other `cert-manager-*` pods
and the `cert-manager-webhook-constellix` pod.

If you've generated a `Certificate` but no `CertificateRequest` is generated,
the main `cert-manager` pod logs should show why any action was skipped.

Since this project is essentially a plugin to `cert-manager`, detailed docs
mainly live in the `cert-manager` project. Here are some specific docs that may
be helpful:

* About [ACME](https://cert-manager.io/docs/configuration/acme/)
* About [DNS01 Challenges](https://cert-manager.io/docs/configuration/acme/dns01/)
* [Troubleshooting Issuing ACME Certificates](https://cert-manager.io/docs/faq/acme/)
* [Securing Ingress Resources](https://cert-manager.io/docs/usage/ingress/)

## Development

### Running the test suite

Unit tests (no live DNS):

```bash
go test ./... -count=1
```

End-to-end **conformance** against a live zone requires [envtest](https://book.kubebuilder.io/reference/envtest.html) assets (`etcd`, `kube-apiserver`, `kubectl` on `PATH` or `TEST_ASSET_*` set—see [testdata/constellix/README.md](testdata/constellix/README.md)), Constellix credentials, and `TEST_ZONE_NAME`. There is no single `go test` entrypoint in-tree; use the cert-manager DNS webhook testing approach and the files under `testdata/constellix/` when you set up envtest locally.

### Maintaining the Docker image and Helm repository

See the `Makefile` for building and pushing the image (`build`, `buildx`, `buildx-multi`) and for packaging the Helm chart. The image uses a multi-stage build with a **distroless** runtime; build with a current Docker (BuildKit recommended).

## Contributions

Pull Requests and issues are welcome. See the
[Contellix Contribution Guidelines](https://github.com/constellix/community/blob/master/Contributing.md)
for more information.
