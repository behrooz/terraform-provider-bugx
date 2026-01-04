# Example: Create a secret
# resource "bugx_secret" "example" {
#   name        = "my-secret"
#   description = "Example secret for testing"
  
#   data = {
#     username = "admin"
#     password = "secret-password"
#     api_key  = "sk-1234567890abcdef"
#   }
# }

# Example: Create another secret with more data
resource "bugx_secret" "database" {
  name        = "database-credentials"
  description = "Database connection credentials"
  
  data = {
    host     = "db.example.com"
    port     = "5432"
    database = "mydb"
    user     = "dbuser"
    password = "dbpassword"
    DBPASS = "1234"
    REDISUSER = "2323"
  }
}

# Output the secret ID (note: values are sensitive and won't be shown)
# output "secret_id" {
#   value       = bugx_secret.example.id
#   description = "ID of the created secret"
# }

# output "secret_created_at" {
#   value       = bugx_secret.example.created_at
#   description = "Timestamp when the secret was created"
# }

