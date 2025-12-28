terraform {
  required_providers {
    vcluster = {
      source  = "local/vcluster/vcluster"
      version = "0.1"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.4"
    }
  }
}

provider "vcluster" {
  base_url = "http://192.168.1.4"

  # Credentials used for POST /login. The provider will automatically
  # call /login, get the token from {"token": "..."} and send it as
  # the Authorization header on subsequent API calls.
  username = "admin"
  password = "admin"
}

# resource "vcluster_cluster" "devcluster" {
#   name             = "devcluster"
#   cluster_id       = "2qjqhhqr"
#   control_plane    = "k8s"  
#   cpu              = "1"
#   memory           = "1024"
#   platform_version = "v1.31.6"
#   cluster_type     = "medium"
#   coredns_cpu      = "0.5"
#   coredns_memory   = "500Mi"
#   apiserver_cpu    = "0.5"
#   apiserver_memory = "500Mi"
# }

# # Option 1: Save kubeconfig directly to a file
# resource "local_file" "kubeconfig" {
#   filename = "${path.module}/kubeconfig-${vcluster_cluster.devcluster.name}.yaml"
#   content  = try(vcluster_cluster.devcluster.kubeconfig, "")
#   file_permission = "0600"
# }

# # Option 2: Use templatefile if you want to customize the kubeconfig
# locals {
#   kubeconfig_name = vcluster_cluster.devcluster.name
#   endpoint        = vcluster_cluster.devcluster.endpoint
# }

# # Output the kubeconfig (will be marked as sensitive)
# output "kubeconfig" {
#   value     = vcluster_cluster.devcluster.kubeconfig
#   sensitive = true
# }

# # Output the kubeconfig file path
# output "kubeconfig_file" {
#   value = local_file.kubeconfig.filename
# }