# bugx_secret Resource

Manages secrets in the bugx API. This resource creates, updates, and deletes secrets via the `/secrets/api/v1/secrets` endpoint.

## Example Usage

### Basic Secret

```hcl
resource "bugx_secret" "example" {
  name        = "my-secret"
  description = "Example secret for testing"
  
  data = {
    username = "admin"
    password = "secret-password"
    api_key  = "sk-1234567890abcdef"
  }
}
```

### Secret with Output

```hcl
resource "bugx_secret" "api_credentials" {
  name        = "api-credentials"
  description = "API credentials for external service"
  
  data = {
    api_key    = var.api_key
    api_secret = var.api_secret
  }
}

output "secret_id" {
  value = bugx_secret.api_credentials.id
}

output "secret_created_at" {
  value = bugx_secret.api_credentials.created_at
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the secret (must be unique)
* `description` - (Optional) Optional description of the secret
* `data` - (Required, Sensitive) Map of key-value pairs containing the secret data. All values must be strings

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `created_at` - (Computed) Timestamp when the secret was created
* `updated_at` - (Computed) Timestamp when the secret was last updated

## Import

Secrets can be imported using the secret ID:

```bash
terraform import bugx_secret.example <secret-id>
```

## Notes

* The `data` attribute is marked as sensitive and will not be displayed in Terraform output
* Secret names must be unique within the bugx API
* The resource uses the `/secrets/api/v1/secrets` endpoint. Make sure your API base URL points to the correct server
* When importing, you can use either the secret ID or name
* The provider will automatically look up secrets by name if the ID is not available

