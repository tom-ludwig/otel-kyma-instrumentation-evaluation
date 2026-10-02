# Demo app: OTel operator auto-instrumentation + Kyma managed collector

## Prerequisites

- **Cert-Manager**

  > [!WARNING] Cert-manager Values
  > Cert Manager values are tuned for a Gardener Shoot Cluster. Double-check if those values apply to you.

  ```bash
  helm install \
    cert-manager oci://quay.io/jetstack/charts/cert-manager \
    --namespace cert-manager \
    --create-namespace \
    --version v1.21.2 \
    --values k8s/cert-manager-values.yaml
  ```

- **OpenTelemetry Operator**

  ```bash
  helm repo add opentelemetry-helm https://open-telemetry.github.io/opentelemetry-helm-charts
  helm install my-opentelemetry-operator opentelemetry-helm/opentelemetry-operator \
    --version 0.123.1 \
    --values k8s/otel-operator-values.yaml
  ```

- **Kyma Telemetry and Istio module** enabled (provides `telemetry-otlp-traces.kyma-system` service)
  - https://kyma-project.io/02-get-started/01-quick-install.html

- **(Optional) Jaeger** (in-memory, for trace visualization):

  ```bash
  helm repo add jaegertracing https://jaegertracing.github.io/helm-charts
  helm install jaeger jaegertracing/jaeger \
    --namespace demo --create-namespace \
    --values k8s/jaeger-values.yaml
  ```

## Apps

All three apps share a single Gateway API `Gateway` (defined in `k8s/istio.yaml`) and route via `HTTPRoute` resources bundled in each deployment file. Apply the Telemetry CR and TracePipeline once:

```bash
kubectl apply -f k8s/tracepipeline.yaml
kubectl apply -f k8s/istio.yaml
```

Then deploy each app independently:

| App     | File                       | Host              | Instrumentation                  |
| ------- | -------------------------- | ----------------- | -------------------------------- |
| Go      | `k8s/deployment-go.yaml`   | `demo-go.local`   | eBPF sidecar (`inject-go`)       |
| Node.js | `k8s/deployment-node.yaml` | `demo-node.local` | init-container (`inject-nodejs`) |
| Java    | `k8s/deployment-java.yaml` | `demo-java.local` | init-container (`inject-java`)   |

```bash
kubectl apply -f k8s/deployment-go.yaml
kubectl apply -f k8s/deployment-node.yaml
kubectl apply -f k8s/deployment-java.yaml
```

Each file bundles the app Deployment, Service, Postgres (if needed), Instrumentation CR, and HTTPRoute.

### Generate traffic

```bash
INGRESS=$(kubectl -n istio-system get gateway demo-gateway -o jsonpath='{.status.addresses[0].value}')
curl -H "Host: demo-go.local"   http://$INGRESS/rolldice
curl -H "Host: demo-node.local" http://$INGRESS/rolldice
curl -H "Host: demo-java.local" http://$INGRESS/rolldice
```

Or via port-forward:

```bash
kubectl -n istio-system port-forward svc/istio-ingressgateway 8080:80
curl -H "Host: demo-go.local" http://localhost:8080/rolldice
```

## Open the Jaeger UI

```bash
kubectl -n demo port-forward svc/jaeger 16686:16686
```

Then open http://localhost:16686.

## How it works

### Go (eBPF)

> **WARNING:** The Go eBPF auto-instrumentation sidecar requires `runAsUser: 0`.
> If your cluster enforces a restrictive PodSecurityAdmission policy, you must
> allow privileged containers in the `demo` namespace before deploying.

The otel-operator's mutating webhook detects the annotation
`instrumentation.opentelemetry.io/inject-go: "demo-instrumentation"` and injects
an eBPF-based sidecar into the pod. The sidecar intercepts HTTP calls at the kernel
level, no code changes or SDK imports required. Spans are exported via OTLP/HTTP
to Kyma's managed collector, which forwards them to Jaeger via the TracePipeline.

- The sidecar runs as root (required for eBPF map access).
- `OTEL_GO_AUTO_TARGET_EXE` must point to the exact binary path inside the app container (`/demo-app`).

### Node.js / Java (init-container)

The otel-operator detects `inject-nodejs` / `inject-java` and injects an init-container
that copies the OTel SDK into the app container. For Node.js it loads via
`NODE_OPTIONS=--require`; for Java via `-javaagent`. Both instrument the app and its DB
driver without code changes. The Java Instrumentation CR disables the metrics and logs
exporters since Kyma's collector only accepts `/v1/traces`.

You get more detailed traces with the injection method than with the eBPF sidecar model; for example, detailed database query insights.

The initial injection costs substantial compute time on container startup since the bytecode has to be rewritten; with larger projects that have more dependencies, this will take even longer. Furthermore, in the past there have been high-severity CVEs related to this injection method. Ask yourself whether it is worth the risk of injecting code that you don't fully understand or trust into your deployed applications.

## Debug collector

`k8s/debug-collector.yaml` deploys a bare OTel collector that prints every span to stdout (detailed verbosity). Point an Instrumentation CR's `exporter.endpoint` at `http://otel-debug-collector.demo.svc.cluster.local:4317` to bypass Kyma and inspect raw spans:

```bash
kubectl -n demo logs -l app=otel-debug-collector -f
```
