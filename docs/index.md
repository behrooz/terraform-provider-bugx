# bugx Provider

The bugx provider manages bugx clusters, Helm releases, secrets, and provides data sources for querying existing resources through the bugx API.

## Example Usage

```hcl
terraform {
  required_providers {
    bugx = {
      source  = "behrooz/bugx"
      version = "~> 1.8"
    }
  }
}

provider "bugx" {
  # Authentication: Use either username/password OR access_key/secret_key
  # Option 1: Username and Password
  username = "admin"
  password = "admin"
  
  # Option 2: Access Key and Secret Key (alternative to username/password)
  # access_key = "your-access-key"
  # secret_key = "your-secret-key"
  
  # Optional: Configure timeout (in seconds, default: 300)
  timeout = 300
  
  # Optional: Configure max retries for failed requests (default: 3)
  max_retries = 3
}

resource "bugx_cluster" "example" {
  name             = "mycluster"
  cluster_id       = "2qjqhhqr"
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

### Authentication

The provider supports two authentication methods. You must provide either:

**Option 1: Username and Password**
* `username` - (Optional) Username for login to bugx API (required if `access_key` is not provided)
* `password` - (Optional, Sensitive) Password for login to bugx API (required if `secret_key` is not provided)

**Option 2: Access Key and Secret Key**
* `access_key` - (Optional, Sensitive) Access key for login to bugx API (required if `username` is not provided)
* `secret_key` - (Optional, Sensitive) Secret key for login to bugx API (required if `password` is not provided)

**Note:** You must provide either `username`/`password` OR `access_key`/`secret_key`. Both authentication methods will return a JWT token that is used for subsequent API calls.

### Configuration Options

* `timeout` - (Optional) HTTP client timeout in seconds (default: `300`)
* `max_retries` - (Optional) Maximum number of retries for failed requests (default: `3`)

**Note:** The base URL is hardcoded to `https://api.bugx.ir` and cannot be configured.

## Features

* **Cluster Management**: Create, read, update, and delete bugx instances
* **Helm Release Management**: Deploy and manage Helm charts on bugx clusters
* **Secret Management**: Create, read, update, and delete secrets via REST API
* **Data Sources**: Query existing clusters without managing them
* **Retry Logic**: Automatic retry with exponential backoff for transient network errors
* **Configurable Timeouts**: Customizable HTTP client timeouts and retry settings
* **Resource Import**: Import existing clusters and secrets into Terraform state
* **Chart Version Support**: Pin specific Helm chart versions for reproducible deployments

