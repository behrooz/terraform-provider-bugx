# Example: Deploy MySQL on vcluster using the helm_install API
# This uses your custom API endpoint instead of local kubeconfig

# Deploy MySQL on the devcluster cluster
resource "vcluster_helm_release" "mysql" {
  cluster_name = vcluster_cluster.devcluster.name
  namespace   = "default"
  release     = "mysql"
  chart       = "bitnami/mysql"
  repo        = "https://charts.bitnami.com/bitnami"
  
  # Option 1: Use a values file
  values_file = "${path.module}/helm-values/mysql-values.yaml"
  
  # Option 2: Use inline values (alternative to values_file)
  # values = <<-EOT
  #   auth:
  #     rootPassword: "myrootpassword"
  #     database: "mydatabase"
  #   primary:
  #     persistence:
  #       size: "8Gi"
  #     resources:
  #       requests:
  #         memory: "512Mi"
  #         cpu: "250m"
  # EOT
  
  # Wait for cluster to be ready before deploying
  depends_on = [vcluster_cluster.devcluster]
}

# You can also use templatefile() to generate values dynamically
# locals {
#   mysql_values = templatefile("${path.module}/helm-values/mysql-values.tpl", {
#     root_password = "myrootpassword"
#     database_name  = "mydatabase"
#     storage_size   = "8Gi"
#   })
# }
# 
# resource "vcluster_helm_release" "mysql_template" {
#   cluster_name = vcluster_cluster.devcluster.name
#   namespace   = "default"
#   release     = "mysql"
#   chart       = "bitnami/mysql"
#   repo        = "https://charts.bitnami.com/bitnami"
#   values      = local.mysql_values
#   depends_on  = [vcluster_cluster.devcluster]
# }

# Output MySQL connection info
output "mysql_release" {
  value = vcluster_helm_release.mysql.release
}

