# bugx_cluster Resource

Manages a bugx cluster instance. This resource creates, updates, and deletes bugx clusters through the bugx API.

## Example Usage

```hcl
resource "bugx_cluster" "example" {
  name             = "mycluster"
  control_plane    = "k8s"
  cpu              = "1"
  memory           = "1024"
  platform_version = "v1.31.6"
  cluster_type     = "tiny"
  coredns_cpu      = "0.5"
  coredns_memory   = "0.250Gi"
  apiserver_cpu    = "0.5"
  apiserver_memory = "0.250Gi"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the cluster
* `cluster_id` - (Optional) Unique identifier for the cluster. If not provided, the API will generate one
* `control_plane` - (Required) Control plane type (e.g., `k8s`)
* `cpu` - (Required) CPU allocation for the cluster
* `memory` - (Required) Memory allocation for the cluster (in MB or with unit like `1024`)
* `platform_version` - (Required) Platform version (e.g., `v1.31.6`)
* `cluster_type` - (Required) Type of cluster (e.g., `tiny`)
* `coredns_cpu` - (Required) CPU allocation for CoreDNS (e.g., `0.5`)
* `coredns_memory` - (Required) Memory allocation for CoreDNS (e.g., `0.250Gi`)
* `apiserver_cpu` - (Required) CPU allocation for API server (e.g., `0.5`)
* `apiserver_memory` - (Required) Memory allocation for API server (e.g., `0.250Gi`)
* `status` - (Optional) Initial status of the cluster (default: `Progressing`)
* `health_check` - (Optional) Health check configuration
* `alert` - (Optional) Alert configuration

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `cluster_id` - (Computed) Unique identifier for the cluster (populated after creation if not provided)
* `endpoint` - (Computed) Cluster endpoint URL
* `namespace` - (Computed) Kubernetes namespace where the cluster is deployed
* `kubeconfig` - (Computed, Sensitive) Kubeconfig content for connecting to the cluster (only available when cluster status is `Healthy`)

## Import

Clusters can be imported using the cluster ID:

```bash
terraform import bugx_cluster.example <cluster-id>
```

## Notes

* The provider will automatically poll the cluster status after creation until it becomes `Healthy`
* The `kubeconfig` attribute is only populated when the cluster status is `Healthy`
* Cluster deletion requires both the cluster name and namespace

