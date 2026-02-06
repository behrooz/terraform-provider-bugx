# Provider Configuration

The `bugx` provider is used to interact with the bugx API to manage clusters, Helm releases, and secrets.

## Authentication

The provider supports two authentication methods. You must provide credentials using one of these methods:

### Option 1: Username and Password

Authenticate using a username and password:

```hcl
provider "bugx" {
  username = "admin"
  password = "admin"
}
```

**Arguments:**
* `username` - (Optional) Username for login to bugx API. Required if `access_key` is not provided.
* `password` - (Optional, Sensitive) Password for login to bugx API. Required if `secret_key` is not provided.

### Option 2: Access Key and Secret Key

Authenticate using an access key and secret key:

```hcl
provider "bugx" {
  access_key = "your-access-key"
  secret_key = "your-secret-key"
}
```

**Arguments:**
* `access_key` - (Optional, Sensitive) Access key for login to bugx API. Required if `username` is not provided.
* `secret_key` - (Optional, Sensitive) Secret key for login to bugx API. Required if `password` is not provided.

**Note:** Both authentication methods will send credentials to the `/login` endpoint, which returns a JWT token. This token is then used in the `Authorization` header for all subsequent API calls.

## Configuration Options

### timeout

(Optional) HTTP client timeout in seconds. Default: `300`

```hcl
provider "bugx" {
  username = "admin"
  password = "admin"
  timeout  = 600  # 10 minutes
}
```

### max_retries

(Optional) Maximum number of retries for failed requests. Default: `3`

```hcl
provider "bugx" {
  username    = "admin"
  password    = "admin"
  max_retries = 5
}
```

## Complete Example

```hcl
terraform {
  required_providers {
    bugx = {
      source  = "behrooz/bugx"
      version = "~> 1.0"
    }
  }
}

# Using username/password authentication
provider "bugx" {
  username = "admin"
  password = "admin"
  timeout  = 300
  max_retries = 3
}

# Or using access_key/secret_key authentication
# provider "bugx" {
#   access_key = "your-access-key"
#   secret_key = "your-secret-key"
#   timeout     = 300
#   max_retries = 3
# }
```

## Base URL

The base URL is hardcoded to `https://bugx.ir` and cannot be configured through the provider. All API requests will be made to this base URL.

## Authentication Flow

1. Provider sends authentication credentials (username/password or access_key/secret_key) to `POST /login`
2. API responds with a JWT token: `{"token": "..."}`
3. Provider stores the token and includes it in the `Authorization` header for all subsequent requests
4. Token is automatically refreshed if it expires (handled by the API)

