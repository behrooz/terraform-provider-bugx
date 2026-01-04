# # Example: Deploy rabbitmq on bugx using the helm_install API
# # This uses your custom API endpoint instead of local kubeconfig

# # Deploy rabbitmq on the debugx cluster
# resource "bugx_helm_release" "rabbitmq" {
#   cluster_name = bugx_cluster.debugx.name
#   namespace   = "default"
#   release     = "rabbitmq"
#   chart       = "bitnami/rabbitmq"
#   repo        = "https://charts.bitnami.com/bitnami"
  
#   # Option 1: Use a values file
#   values_file = "${path.module}/helm-values/rabbitmq-values.yaml"
    
#   # Wait for cluster to be ready before deploying
#   depends_on = [bugx_cluster.debugx]
# }

# # Output rabbitmq connection info
# output "rabbitmq_release" {
#   value = bugx_helm_release.rabbitmq.release
# }

