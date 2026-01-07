# bugx_orphan_cleanup Resource

Manages cleanup of orphaned applications (applications that exist on the server but are not in Terraform state) from a bugx cluster. This resource helps maintain consistency by removing applications that are no longer managed by Terraform.

## Example Usage

### Explicit App Cleanup

```hcl
resource "bugx_orphan_cleanup" "example" {
  cluster_name = bugx_cluster.example.name
  
  apps_to_delete = [
    "ns-977i-rabbitmq",
    "ns-977i-redis",
    "ns-977i-mysql"
  ]
  
  depends_on = [bugx_cluster.example]
}
```

### Cleanup with Keep List

```hcl
resource "bugx_orphan_cleanup" "example" {
  cluster_name = bugx_cluster.example.name
  
  keep_releases = [
    "nginx",
    "cert-manager"
  ]
  
  depends_on = [bugx_cluster.example]
}
```

## Argument Reference

The following arguments are supported:

* `cluster_name` - (Required, ForceNew) Name of the bugx cluster to clean up orphaned applications from
* `apps_to_delete` - (Optional) Set of application names to delete explicitly. These should be the full app names (e.g., `ns-977i-rabbitmq` for cluster namespace `ns-977i` and release `rabbitmq`)
* `keep_releases` - (Optional) Set of Helm release names to keep. If provided along with cluster namespace, apps matching `{namespace}-{release}` pattern that are NOT in this list will be deleted. Use this for automatic cleanup based on release names

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `deleted_apps` - (Computed) List of application names that were successfully deleted

## Notes

* This resource does not create any server-side state. It performs cleanup operations and tracks what was deleted
* The resource ID is set to `{cluster_name}-orphan-cleanup`
* If no apps are specified for deletion, the resource will be created but no cleanup will occur
* App names should be the full application names as they appear in the bugx API (typically `{cluster_namespace}-{release_name}`)
* The `keep_releases` option requires knowledge of which releases might exist, as the API may not provide a list endpoint
* Changes to `apps_to_delete` or `keep_releases` will trigger a re-run of the cleanup operation
* Deleting this resource does not restore any deleted applications

