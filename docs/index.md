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
  base_url = "http://localhost:8082"
  username = "admin"
  password = "admin"
  
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

* `base_url` - (Required) Base URL of bugx API, e.g. `http://192.168.1.4` or `http://localhost:8082`
* `username` - (Required) Username for login to bugx API
* `password` - (Required) Password for login to bugx API (sensitive)
* `timeout` - (Optional) HTTP client timeout in seconds (default: `300`)
* `max_retries` - (Optional) Maximum number of retries for failed requests (default: `3`)

## Features

* **Cluster Management**: Create, read, update, and delete bugx instances
* **Helm Release Management**: Deploy and manage Helm charts on bugx clusters
* **Secret Management**: Create, read, update, and delete secrets via REST API
* **Data Sources**: Query existing clusters without managing them
* **Retry Logic**: Automatic retry with exponential backoff for transient network errors
* **Configurable Timeouts**: Customizable HTTP client timeouts and retry settings
* **Resource Import**: Import existing clusters and secrets into Terraform state
* **Chart Version Support**: Pin specific Helm chart versions for reproducible deployments

