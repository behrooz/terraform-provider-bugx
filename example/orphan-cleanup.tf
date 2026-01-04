# # Example: Clean up orphaned applications
# # This resource can be used to delete applications that exist on the server
# # but are not managed by Terraform (e.g., after commenting out a resource)

# # Option 1: Explicitly specify apps to delete by their full app names
# # App names follow the pattern: {cluster_namespace}-{release_name}
# # For example, if cluster namespace is "ns-977i" and release is "rabbitmq",
# # the app name would be "ns-977i-rabbitmq"
# resource "bugx_orphan_cleanup" "cleanup" {
#   cluster_name = bugx_cluster.debugx.name
  
#   # Explicitly list apps to delete
#   apps_to_delete = [
#     "ns-977i-rabbitmq",  # Example: delete rabbitmq if it was commented out
#     # Add more app names here as needed
#   ]
  
#   depends_on = [bugx_cluster.debugx]
# }

# # Option 2: Use keep_releases to specify which releases should exist
# # This is useful when you know which releases should be kept, and want to
# # delete any others. However, without a list API, you still need to know
# # which apps might exist.
# # resource "bugx_orphan_cleanup" "cleanup" {
# #   cluster_name = bugx_cluster.debugx.name
# #   
# #   # Specify which releases to keep
# #   keep_releases = [
# #     "mysql",
# #     "redis",
# #     # rabbitmq is not in the list, so if it exists, it should be deleted
# #   ]
# #   
# #   depends_on = [bugx_cluster.debugx]
# # }

# # Output the list of deleted apps
# output "deleted_orphan_apps" {
#   value       = bugx_orphan_cleanup.cleanup.deleted_apps
#   description = "List of orphaned applications that were deleted"
# }

