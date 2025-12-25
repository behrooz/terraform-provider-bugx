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
  
  depends_on = [vcluster_cluster.devcluster]
}

# Output MySQL connection info
output "mysql_release" {
  value = vcluster_helm_release.mysql.release
}

