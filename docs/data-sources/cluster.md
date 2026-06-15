# bugx_cluster Data Source

Queries an existing bugx cluster without managing it. This data source allows you to read cluster information for use in other resources or outputs.

## Example Usage

### Basic Query

```hcl
data "bugx_cluster" "existing" {
  name = "mycluster"
}

output "cluster_status" {
  value = data.bugx_cluster.existing.status
}

output "cluster_endpoint" {
  value = data.bugx_cluster.existing.endpoint
}
```

### Using Cluster Data in Other Resources

```hcl
data "bugx_cluster" "production" {
  name = "prod-cluster"
}

resource "bugx_helm_release" "app" {
  cluster_name = data.bugx_cluster.production.name
  namespace   = "default"
  release     = "myapp"
  chart       = "myapp/chart"
  repo        = "https://charts.example.com"
}
```

### Accessing Kubeconfig

```hcl
data "bugx_cluster" "existing" {
  name = "mycluster"
}

# Note: kubeconfig is sensitive and won't be displayed in outputs
# but can be used in other resources or written to files
resource "local_file" "kubeconfig" {
  content  = data.bugx_cluster.existing.kubeconfig
  filename = "${path.module}/kubeconfig.yaml"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the bugx cluster to query

## Attribute Reference

The following attributes are exported:

* `cluster_id` - Cluster ID
* `status` - Current status of the cluster
* `endpoint` - Cluster endpoint URL
* `namespace` - Kubernetes namespace where the cluster is deployed
* `version` - Platform version of the cluster
* `kubeconfig` - (Sensitive) Kubeconfig content for connecting to the cluster (only available when cluster status is `Healthy`)

## Notes

* The `kubeconfig` attribute is only populated when the cluster status is `Healthy`
* If the cluster is not found, Terraform will return an error
* The `kubeconfig` attribute is marked as sensitive and will not be displayed in Terraform output or logs

