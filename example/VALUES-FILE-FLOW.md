# How Values File is Sent to API

## Flow Diagram

```
┌─────────────────────────────────┐
│  mysql-values.yaml              │
│  (YAML file on disk)            │
└──────────────┬──────────────────┘
               │
               │ os.ReadFile()
               │ (reads entire file)
               ▼
┌─────────────────────────────────┐
│  fileContent ([]byte)          │
│  → string(fileContent)          │
└──────────────┬──────────────────┘
               │
               │ payload.Values = string(fileContent)
               ▼
┌─────────────────────────────────┐
│  HelmInstallPayload struct      │
│  {                              │
│    Clustername: "myttiny"      │
│    Namespace: "default"        │
│    Release: "mysql"            │
│    Chart: "bitnami/mysql"       │
│    Repo: "https://..."          │
│    Values: "auth:\n  root..."   │ ← YAML as string
│  }                              │
└──────────────┬──────────────────┘
               │
               │ json.Marshal()
               ▼
┌─────────────────────────────────┐
│  JSON Request Body               │
│  {                              │
│    "Clustername": "myttiny",    │
│    "Namespace": "default",      │
│    "Release": "mysql",          │
│    "Chart": "bitnami/mysql",    │
│    "Repo": "https://...",       │
│    "Values": "auth:\\n  root..." │ ← Escaped YAML string
│  }                              │
└──────────────┬──────────────────┘
               │
               │ POST /helm_install
               │ Content-Type: application/json
               │ Authorization: <token>
               ▼
┌─────────────────────────────────┐
│  Your API Endpoint               │
│  http://localhost:8082/helm_install
└─────────────────────────────────┘
```

## Example

### Input: `helm-values/mysql-values.yaml`
```yaml
auth:
  rootPassword: "myrootpassword"
  database: "mydatabase"
primary:
  persistence:
    size: "8Gi"
```

### Output: JSON sent to API
```json
{
  "Clustername": "myttiny",
  "Namespace": "default",
  "Release": "mysql",
  "Chart": "bitnami/mysql",
  "Repo": "https://charts.bitnami.com/bitnami",
  "Values": "auth:\n  rootPassword: \"myrootpassword\"\n  database: \"mydatabase\"\nprimary:\n  persistence:\n    size: \"8Gi\"\n"
}
```

## Important Notes

1. **The entire YAML file is read as text** - no parsing or validation happens in Terraform
2. **Newlines are escaped** - `\n` in the JSON string
3. **The YAML is sent as a string** - your API receives it in the `Values` field
4. **Your API should parse the YAML** - The provider doesn't validate or parse the YAML, it just sends it as-is

## Code Location

The logic is in `resource_helm_release.go`:

- **File reading**: Lines 89-95
- **JSON marshaling**: Line 116
- **HTTP request**: Lines 121-126

