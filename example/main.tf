terraform {
  required_providers {
    bugx = {
      source  = "behrooz/bugx"
      version = "1.8.17"
    }
  }
}

provider "bugx" {

  # Credentials used for POST /login. The provider will automatically
  # call /login, get the token from {"token": "..."} and send it as
  # the Authorization header on subsequent API calls.
  access_key = ""
  secret_key = ""
}

resource "bugx_cluster" "devcluster" {
  name             = "devcluster"
  control_plane    = "k8s"
  cluster_type     = "large"
  cpu              = "4"
  memory           = "8Gi"
  platform_version = "v1.31.6"
  coredns_cpu      = "512m"
  coredns_memory   = "512Mi"
  apiserver_cpu    = "512m"
  apiserver_memory = "512Mi"
}

# Option 1: Save kubeconfig directly to a file
resource "local_file" "kubeconfig" {
  filename        = "${path.module}/kubeconfig-${bugx_cluster.devcluster.name}.yaml"
  content         = try(bugx_cluster.devcluster.kubeconfig, "")
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