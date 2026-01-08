# Example: Deploy MySQL on bugx using the helm_install API
# This uses your custom API endpoint instead of local kubeconfig

# Deploy MySQL on the devcluster cluster
resource "bugx_helm_release" "mysql" {
  cluster_name = bugx_cluster.devcluster.name
  namespace   = "default"
  release     = "mysql"
  chart       = "bitnami/mysql"
  
  # Option 1: Use a values file
  values_file = "${path.module}/helm-values/mysql-values.yaml"
  
  depends_on = [bugx_cluster.devcluster]
}

# Output MySQL connection info
output "mysql_release" {
  value = bugx_helm_release.mysql.release
}

