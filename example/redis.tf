# # Example: Deploy redis on bugx using the helm_install API
# # This uses your custom API endpoint instead of local kubeconfig

# # Deploy redis on the debugx cluster
# resource "bugx_helm_release" "redis" {
#   cluster_name = bugx_cluster.debugx.name
#   namespace   = "default"
#   release     = "redis"
#   chart       = "bitnami/redis"
#   repo        = "https://charts.bitnami.com/bitnami"
  
#   # Option 1: Use a values file
#   values_file = "${path.module}/helm-values/redis-values.yaml"
#   # Wait for cluster to be ready before deploying
#   depends_on = [bugx_cluster.debugx]
# }


# # Output redis connection info
# output "redis_release" {
#   value = bugx_helm_release.redis.release
# }

