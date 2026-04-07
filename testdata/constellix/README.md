# Constellix solver testdata (conformance / integration)

Use these files when running **cert-manager DNS01 webhook conformance**–style tests against a **live** Constellix zone (TXT create/delete). This is optional for normal development; **unit tests** live in the repository root (`go test ./...`).

## Prerequisites

- A DNS zone hosted in Constellix and resolvable on the public internet.
- `TEST_ZONE_NAME` set to that zone’s apex (for example `example.com`), without a trailing dot.
- [envtest](https://book.kubebuilder.io/reference/envtest.html) binaries available: `etcd`, `kube-apiserver`, and `kubectl` either on your `PATH` or via `TEST_ASSET_ETCD`, `TEST_ASSET_KUBE_APISERVER`, and `TEST_ASSET_KUBECTL`. The script [`scripts/fetch-test-binaries.sh`](../../scripts/fetch-test-binaries.sh) may help download compatible versions into `_out/kubebuilder/bin`—add that directory to `PATH` before running tests that embed cert-manager’s test control plane.

## Setup

1. **Webhook config**

   Copy one of the samples to `config.json` and edit values:

   - [`config.json.sample`](config.json.sample) — single Constellix domain: set **`zoneId`** to the numeric domain ID in Constellix (Sonar/API “domain” id).
   - [`config-zones.json.sample`](config-zones.json.sample) — multiple domains: set **`zones`** with `dnsName` (zone apex) and `zoneId`. The entry whose `dnsName` matches `TEST_ZONE_NAME` (longest suffix match) is used.

   `apiKeySecretRef` / `apiSecretSecretRef` must match the Secret you apply in the next step (`name` and `key` fields).

2. **Credentials Secret**

   Copy [`api-key.yaml.sample`](api-key.yaml.sample) to `api-key.yaml`, replace the placeholders with base64-encoded Constellix **API key** and **secret key**, then create the Secret in the cluster the tests use (same namespace as the test expects for the webhook, per cert-manager’s DNS test harness).

## Running conformance

Conformance testing follows cert-manager’s [webhook example](https://github.com/cert-manager/webhook-example) pattern: you need envtest assets, `config.json`, secrets, and `TEST_ZONE_NAME`. Importing cert-manager’s `test/acme` package **requires** those binaries to be discoverable **before** the test binary starts, or the process will exit during package initialization.

Ensure `config.json` and `api-key.yaml` are not committed if they contain real credentials (see `.gitignore`).
