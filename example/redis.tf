# Example: Deploy redis on vcluster using the helm_install API
# This uses your custom API endpoint instead of local kubeconfig

# Deploy redis on the devcluster cluster
resource "vcluster_helm_release" "redis" {
  cluster_name = vcluster_cluster.devcluster.name
  namespace   = "default"
  release     = "redis"
  chart       = "bitnami/redis"
  repo        = "https://charts.bitnami.com/bitnami"
  
  # Option 1: Use a values file
  values_file = "${path.module}/helm-values/redis-values.yaml"
  # Wait for cluster to be ready before deploying
  depends_on = [vcluster_cluster.devcluster]
}


# Output redis connection info
output "redis_release" {
  value = vcluster_helm_release.redis.release
}

