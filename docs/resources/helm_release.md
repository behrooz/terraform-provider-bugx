# bugx_helm_release Resource

Manages a Helm release deployed on a bugx cluster. This resource installs, updates, and deletes Helm charts on specified clusters.

## Example Usage

### Basic Helm Release

```hcl
resource "bugx_helm_release" "mysql" {
  cluster_name = bugx_cluster.example.name
  namespace   = "default"
  release     = "mysql"
  chart       = "bitnami/mysql"
  repo        = "https://charts.bitnami.com/bitnami"
  depends_on  = [bugx_cluster.example]
}
```

### Helm Release with Chart Version

```hcl
resource "bugx_helm_release" "mysql" {
  cluster_name  = bugx_cluster.example.name
  namespace     = "default"
  release       = "mysql"
  chart         = "bitnami/mysql"
  repo          = "https://charts.bitnami.com/bitnami"
  chart_version = "8.0.0"
  values_file   = "${path.module}/helm-values/mysql-values.yaml"
  depends_on    = [bugx_cluster.example]
}
```

### Helm Release with Inline Values

```hcl
resource "bugx_helm_release" "redis" {
  cluster_name = bugx_cluster.example.name
  namespace    = "default"
  release      = "redis"
  chart        = "bitnami/redis"
  repo         = "https://charts.bitnami.com/bitnami"
  
  values = <<-EOT
    auth:
      enabled: true
      password: "mypassword"
    master:
      persistence:
        enabled: true
  EOT
  
  depends_on = [bugx_cluster.example]
}
```

## Argument Reference

The following arguments are supported:

* `cluster_name` - (Required) Name of the bugx cluster where to deploy the Helm release
* `namespace` - (Required) Kubernetes namespace where to deploy the release
* `release` - (Required) Name of the Helm release
* `chart` - (Required) Chart name (e.g., `bitnami/mysql` or `mysql`)
* `repo` - (Required) Helm repository URL (e.g., `https://charts.bitnami.com/bitnami`)
* `chart_version` - (Optional) Version of the Helm chart to install (e.g., `8.0.0`). If not specified, the latest version is used
* `values` - (Optional) Helm values as YAML string. You can use `file()` or `templatefile()` to load from a file
* `values_file` - (Optional) Path to a Helm values YAML file. Alternative to `values` attribute. If both are provided, `values_file` takes precedence

## Attribute Reference

This resource has no exported attributes.

## Notes

* The resource ID is a composite of `cluster_name:namespace:release`
* Changes to `cluster_name`, `namespace`, `release`, `chart`, `repo`, or `chart_version` require resource recreation
* Changes to `values` or `values_file` will trigger a reinstall of the Helm release
* The resource depends on the cluster being in a `Healthy` state before deployment
* When deleting, the provider constructs the app name as `{cluster_namespace}-{release}` for the delete API call

