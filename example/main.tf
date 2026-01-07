terraform {
  required_providers {
    bugx = {
      source  = "local/bugx/bugx"
      version = "0.1"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.4"
    }
  }
}

provider "bugx" {
  # Credentials used for POST /login. The provider will automatically
  # call /login, get the token from {"token": "..."} and send it as
  # the Authorization header on subsequent API calls.
  username = "admin"
  password = "admin"
}

resource "bugx_cluster" "devcluster" {
  name             = "devcluster"
  control_plane    = "k8s"  
  cpu              = "1"
  memory           = "1024"
  platform_version = "v1.31.6"
  cluster_type     = "medium"
  coredns_cpu      = "0.5"
  coredns_memory   = "500Mi"
  apiserver_cpu    = "0.5"
  apiserver_memory = "500Mi"
}

# Option 1: Save kubeconfig directly to a file
resource "local_file" "kubeconfig" {
  filename = "${path.module}/kubeconfig-${bugx_cluster.devcluster.name}.yaml"
  content  = try(bugx_cluster.devcluster.kubeconfig, "")
  file_permission = "0600"
}

# Option 2: Use templatefile if you want to customize the kubeconfig
locals {
  kubeconfig_name = bugx_cluster.devcluster.name
  endpoint        = bugx_cluster.devcluster.endpoint
}

# Output the kubeconfig (will be marked as sensitive)
output "kubeconfig" {
  value     = bugx_cluster.devcluster.kubeconfig
  sensitive = true
}

# Output the kubeconfig file path
output "kubeconfig_file" {
  value = local_file.kubeconfig.filename
}