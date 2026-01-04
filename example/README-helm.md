# Deploying Helm Charts on bugx with Terraform

This guide shows how to deploy Helm charts (like MySQL, PostgreSQL, etc.) on your bugx using the `bugx_helm_release` resource, which uses your `/helm_install` API endpoint.

## Resource Type

The resource type is `bugx_helm_release` (following Terraform naming: `provider_resource`). You can use it like this:

```hcl
resource "bugx_helm_release" "mysql" {
  cluster_name = "myttiny"
  namespace   = "default"
  release     = "mysql"
  chart       = "bitnami/mysql"
  repo        = "https://charts.bitnami.com/bitnami"
  values_file = "${path.module}/helm-values/mysql-values.yaml"
}
```

## Basic Example

```hcl
# Deploy MySQL on your cluster
resource "bugx_helm_release" "mysql" {
  cluster_name = bugx_cluster.myttiny.name
  namespace   = "default"
  release     = "mysql"
  chart       = "bitnami/mysql"
  repo        = "https://charts.bitnami.com/bitnami"
  
  # Wait for cluster to be ready
  depends_on = [bugx_cluster.myttiny]
}
```

## Using Values File

**Option 1: Reference a values file**

```hcl
resource "bugx_helm_release" "mysql" {
  cluster_name = bugx_cluster.myttiny.name
  namespace   = "default"
  release     = "mysql"
  chart       = "bitnami/mysql"
  repo        = "https://charts.bitnami.com/bitnami"
  values_file = "${path.module}/helm-values/mysql-values.yaml"
  
  depends_on = [bugx_cluster.myttiny]
}
```

**Option 2: Inline values**

```hcl
resource "bugx_helm_release" "mysql" {
  cluster_name = bugx_cluster.myttiny.name
  namespace   = "default"
  release     = "mysql"
  chart       = "bitnami/mysql"
  repo        = "https://charts.bitnami.com/bitnami"
  
  values = <<-EOT
    auth:
      rootPassword: "myrootpassword"
      database: "mydatabase"
    primary:
      persistence:
        size: "8Gi"
      resources:
        requests:
          memory: "512Mi"
          cpu: "250m"
  EOT
  
  depends_on = [bugx_cluster.myttiny]
}
```

**Option 3: Using templatefile() for dynamic values**

```hcl
locals {
  mysql_values = templatefile("${path.module}/helm-values/mysql-values.tpl", {
    root_password = "myrootpassword"
    database_name  = "mydatabase"
    storage_size   = "8Gi"
    memory_request = "512Mi"
    cpu_request    = "250m"
  })
}

resource "bugx_helm_release" "mysql" {
  cluster_name = bugx_cluster.myttiny.name
  namespace   = "default"
  release     = "mysql"
  chart       = "bitnami/mysql"
  repo        = "https://charts.bitnami.com/bitnami"
  values      = local.mysql_values
  
  depends_on = [bugx_cluster.myttiny]
}
```

## More Examples

### PostgreSQL

```hcl
resource "bugx_helm_release" "postgresql" {
  cluster_name = bugx_cluster.myttiny.name
  namespace   = "default"
  release     = "postgresql"
  chart       = "bitnami/postgresql"
  repo        = "https://charts.bitnami.com/bitnami"
  values_file = "${path.module}/helm-values/postgresql-values.yaml"
  
  depends_on = [bugx_cluster.myttiny]
}
```

### Redis

```hcl
resource "bugx_helm_release" "redis" {
  cluster_name = bugx_cluster.myttiny.name
  namespace   = "default"
  release     = "redis"
  chart       = "bitnami/redis"
  repo        = "https://charts.bitnami.com/bitnami"
  values_file = "${path.module}/helm-values/redis-values.yaml"
  
  depends_on = [bugx_cluster.myttiny]
}
```

### Nginx Ingress Controller

```hcl
resource "bugx_helm_release" "nginx_ingress" {
  cluster_name = bugx_cluster.myttiny.name
  namespace   = "ingress-nginx"
  release     = "nginx-ingress"
  chart       = "ingress-nginx"
  repo        = "https://kubernetes.github.io/ingress-nginx"
  values_file = "${path.module}/helm-values/nginx-ingress-values.yaml"
  
  depends_on = [bugx_cluster.myttiny]
}
```

## Resource Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `cluster_name` | string | Yes | Name of the bugx where to deploy |
| `namespace` | string | Yes | Kubernetes namespace |
| `release` | string | Yes | Helm release name |
| `chart` | string | Yes | Chart name (e.g., `bitnami/mysql`) |
| `repo` | string | Yes | Helm repository URL |
| `values` | string | No | Helm values as YAML string |
| `values_file` | string | No | Path to Helm values YAML file |

**Note:** Use either `values` or `values_file`, not both. If both are provided, `values_file` takes precedence.

## Running

```bash
# Rebuild the provider with the new resource
cd /home/behrooz/Projects/bugx_terraform
go build -o terraform-provider-bugx
cp terraform-provider-bugx ~/.terraform.d/plugins/local/bugx/bugx/0.1/linux_amd64/

# In your example directory
cd example
terraform init
terraform plan
terraform apply
```

## Deleting

When you run `terraform destroy`, the provider will attempt to call `/helm_uninstall` (or similar endpoint) to uninstall the Helm release. If that endpoint doesn't exist, it will just clear the Terraform state.

## API Endpoints Used

- **Create/Update**: `POST /helm_install` with JSON body:
  ```json
  {
    "Clustername": "myttiny",
    "Namespace": "default",
    "Release": "mysql",
    "Chart": "bitnami/mysql",
    "Repo": "https://charts.bitnami.com/bitnami",
    "Values": "..."
  }
  ```

- **Delete**: `DELETE /helm_uninstall` (if available)

## Tips

1. **Always use `depends_on`**: Make sure your Helm releases depend on the cluster being ready
2. **Version pinning**: Consider adding chart version support if your API supports it
3. **Values files**: Keep values files organized in a `helm-values/` directory
4. **Sensitive data**: For passwords, consider using Terraform variables or secrets management

