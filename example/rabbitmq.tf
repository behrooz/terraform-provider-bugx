# # Example: Deploy rabbitmq on vcluster using the helm_install API
# # This uses your custom API endpoint instead of local kubeconfig

# # Deploy rabbitmq on the devcluster cluster
# resource "vcluster_helm_release" "rabbitmq" {
#   cluster_name = vcluster_cluster.devcluster.name
#   namespace   = "default"
#   release     = "rabbitmq"
#   chart       = "bitnami/rabbitmq"
#   repo        = "https://charts.bitnami.com/bitnami"
  
#   # Option 1: Use a values file
#   values_file = "${path.module}/helm-values/rabbitmq-values.yaml"
    
#   # Wait for cluster to be ready before deploying
#   depends_on = [vcluster_cluster.devcluster]
# }

# # Output rabbitmq connection info
# output "rabbitmq_release" {
#   value = vcluster_helm_release.rabbitmq.release
# }

