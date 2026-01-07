## bugx Terraform Provider

Custom Terraform provider that talks to a bugx API and calls `/createcluster` to create clusters, manage Helm releases, and query existing resources.

### Features

- **Cluster Management**: Create, read, update, and delete bugx instances
- **Helm Release Management**: Deploy and manage Helm charts on bugx clusters
- **Secret Management**: Create, read, update, and delete secrets via REST API
- **Data Sources**: Query existing clusters without managing them
- **Retry Logic**: Automatic retry with exponential backoff for transient network errors
- **Configurable Timeouts**: Customizable HTTP client timeouts and retry settings
- **Resource Import**: Import existing clusters into Terraform state
- **Chart Version Support**: Pin specific Helm chart versions

### Build

```bash
cd /home/behrooz/Projects/vcluster_terraform
go build -o terraform-provider-bugx
```

### Install locally for Terraform

Terraform expects the provider binary in a specific directory based on
`<hostname>/<namespace>/<type>/<version>/<os>_<arch>`.

For a local provider with:

- **source**: `local/bugx/bugx`
- **version**: `0.1`

Copy the binary like this (Linux amd64 example):

```bash
mkdir -p ~/.terraform.d/plugins/local/bugx/bugx/0.1/linux_amd64
cp terraform-provider-bugx ~/.terraform.d/plugins/local/bugx/bugx/0.1/linux_amd64/
```

Adjust the OS/arch folder name if necessary.

### Example Terraform configuration

Create a new directory for using the provider, e.g. `example/` and add `main.tf`:

```hcl
terraform {
  required_providers {
    bugx = {
      source  = "local/bugx/bugx"
      version = "0.1"
    }
  }
}

provider "bugx" {
  # Credentials used for POST /login. The provider will automatically
  # call /login, get the token from {"token": "..."} and send it as
  # the Authorization header on subsequent API calls.
  username = "admin"
  password = "admin"
  
  # Optional: Configure timeout (in seconds, default: 300)
  timeout = 300
  
  # Optional: Configure max retries for failed requests (default: 3)
  max_retries = 3
}

resource "bugx_cluster" "example" {
  name             = "newtiny"
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

Then run:

```bash
cd example
terraform init
terraform apply
```

### Data Sources

Query existing clusters without managing them:

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

### Resource Import

Import existing clusters into Terraform:

```bash
terraform import bugx_cluster.example <cluster-id>
```

### Helm Release with Chart Version

Deploy a specific version of a Helm chart:

```hcl
resource "bugx_helm_release" "mysql" {
  cluster_name = bugx_cluster.devcluster.name
  namespace   = "default"
  release     = "mysql"
  chart       = "bitnami/mysql"
  repo        = "https://charts.bitnami.com/bitnami"
  chart_version = "8.0.0"  # Pin to specific version
  values_file = "${path.module}/helm-values/mysql-values.yaml"
  depends_on  = [bugx_cluster.devcluster]
}
```

### Secret Management

Create, update, and delete secrets:

```hcl
resource "bugx_secret" "example" {
  name        = "my-secret"
  description = "Example secret for testing"
  
  data = {
    username = "admin"
    password = "secret-password"
    api_key  = "sk-1234567890abcdef"
  }
}

# Output the secret metadata (data values are sensitive and won't be shown)
output "secret_id" {
  value = bugx_secret.example.id
}

output "secret_created_at" {
  value = bugx_secret.example.created_at
}
```

**Note**: The secret resource uses the `/secrets/api/v1/secrets` endpoint. Make sure your API base URL points to the correct server (e.g., `http://localhost:5173` for simple-vault API).

**Secret Resource Attributes**:
- `name` (required): Unique name for the secret
- `description` (optional): Description of the secret
- `data` (required): Map of key-value pairs (marked as sensitive)
- `created_at` (computed): Timestamp when the secret was created
- `updated_at` (computed): Timestamp when the secret was last updated

**Import existing secrets**:
```bash
terraform import bugx_secret.example <secret-id>
```

### Improvements

- **Retry Logic**: Automatic retry with exponential backoff for network errors and 5xx status codes
- **Better Error Handling**: Improved error messages for EOF and connection issues
- **Configurable Timeouts**: Set custom HTTP client timeouts per provider instance
- **Data Sources**: Query existing resources without managing them
- **Resource Import**: Import existing clusters into Terraform state
- **Chart Version Support**: Pin specific Helm chart versions for reproducible deployments


