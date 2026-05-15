# storage-metrics-apiserver Helm chart

Installs an aggregated `custom.metrics.k8s.io` API server that exposes
PersistentVolumeClaim storage metrics (capacity, used, available, inodes)
scraped from each kubelet's `/stats/summary` endpoint.

## TL;DR

```bash
kubectl create namespace storage-metrics
helm install storage-metrics-apiserver \
  ./charts/storage-metrics-apiserver \
  --namespace storage-metrics
```

## What this exposes

Under `custom.metrics.k8s.io/v1beta2`, on the `persistentvolumeclaims`
resource (group `""`, version `v1`):

| Metric                            | Unit         | Notes                                  |
|-----------------------------------|--------------|----------------------------------------|
| `volume_capacity_bytes`           | bytes        | Reported by kubelet FS stats           |
| `volume_available_bytes`          | bytes        |                                        |
| `volume_used_bytes`               | bytes        |                                        |
| `volume_used_percentage`          | milli (1000m=100%) | `used / capacity`                |
| `volume_inodes`                   | count        |                                        |
| `volume_inodes_free`              | count        |                                        |
| `volume_inodes_used`              | count        |                                        |
| `volume_inodes_used_percentage`   | milli        | `inodesUsed / inodes`                  |

Works for any volume kubelet mounts as a filesystem PVC — in-tree drivers,
external CSI drivers, and migrated drivers — without depending on
`NodeGetVolumeStats`.

## Values

See `values.yaml`. Common knobs:

- `image.repository`, `image.tag`
- `extraArgs` — append args like `--metric-resolution=30s`,
  `--kubelet-insecure-tls`, `--kubelet-certificate-authority=/path/ca.crt`
- `apiService.create=false` — skip APIService registration if you manage it
  out of band (e.g. with cert-manager + apiservice-registrar)
- `apiService.insecureSkipTLSVerify=true` — set to `false` once you have
  proper serving certs
- `nodeSelector`, `tolerations`, `affinity`
- `podDisruptionBudget.enabled`
