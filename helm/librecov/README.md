# LibreCov Helm Chart

This Helm chart deploys LibreCov, an open-source code coverage history viewer, on a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- PV provisioner support in the underlying infrastructure (if using PostgreSQL persistence)

## Installing the Chart

To install the chart with the release name `librecov`:

```bash
helm install librecov ./helm/librecov
```

Or from the OCI registry:

```bash
helm install librecov oci://ghcr.io/frantche/charts/librecov --version 1.0.0
```

## Uninstalling the Chart

To uninstall/delete the `librecov` deployment:

```bash
helm uninstall librecov
```

## Configuration

The following table lists the configurable parameters of the LibreCov chart and their default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of LibreCov replicas | `1` |
| `image.repository` | LibreCov image repository | `ghcr.io/frantche/librecov` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `image.tag` | Image tag | `latest` |
| `service.type` | Kubernetes service type | `ClusterIP` |
| `service.port` | Service port | `4000` |
| `ingress.enabled` | Enable ingress | `false` |
| `ingress.className` | Ingress class name | `""` |
| `ingress.hosts` | Ingress hosts | `[{"host": "librecov.local", "paths": [{"path": "/", "pathType": "Prefix"}]}]` |
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `512Mi` |
| `resources.requests.cpu` | CPU request | `250m` |
| `resources.requests.memory` | Memory request | `256Mi` |
| `postgresql.enabled` | Enable PostgreSQL | `true` |
| `postgresql.auth.username` | PostgreSQL username | `postgres` |
| `postgresql.auth.password` | PostgreSQL password | `postgres` |
| `postgresql.auth.database` | PostgreSQL database | `librecov` |
| `config.oidc.enabled` | Enable OIDC authentication | `false` |
| `config.oidc.issuer` | OIDC issuer URL | `""` |
| `config.oidc.clientId` | OIDC client ID | `""` |
| `config.oidc.clientSecret` | OIDC client secret | `""` |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example:

```bash
helm install librecov ./helm/librecov \
  --set image.tag=1.0.0 \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=librecov.example.com
```

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart:

```bash
helm install librecov ./helm/librecov -f my-values.yaml
```

## OIDC Configuration

To enable OIDC authentication:

```yaml
config:
  oidc:
    enabled: true
    issuer: "https://your-oidc-provider.com"
    clientId: "your-client-id"
    clientSecret: "your-client-secret"
    redirectUrl: "https://librecov.example.com/auth/callback"
```

## Using External PostgreSQL

To use an external PostgreSQL instance:

```yaml
postgresql:
  enabled: false

externalDatabase:
  host: "postgres.example.com"
  port: 5432
  user: "librecov"
  password: "your-password"
  database: "librecov"
```

## Ingress

To enable ingress:

```yaml
ingress:
  enabled: true
  className: "nginx"
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
  hosts:
    - host: librecov.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: librecov-tls
      hosts:
        - librecov.example.com
```

## Persistence

The chart uses a PersistentVolumeClaim for PostgreSQL data persistence when `postgresql.enabled=true`. The volume is mounted at `/var/lib/postgresql/data`.

To use a specific storage class:

```yaml
postgresql:
  primary:
    persistence:
      enabled: true
      storageClass: "fast-ssd"
      size: 10Gi
```
