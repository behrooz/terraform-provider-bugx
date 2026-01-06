# Publishing to HashiCorp Terraform Registry

This guide explains how to publish your `terraform-provider-bugx` to the HashiCorp Terraform Registry, making it available to users worldwide.

## Prerequisites

1. **GitHub Account** - Your provider must be in a public GitHub repository
2. **Repository Naming** - Must follow the pattern: `terraform-provider-{name}`
   - Your current name: `terraform-provider-bugx` ✅ (correct)
3. **GoReleaser** - For automated releases and binary building
4. **GPG Key** - For signing releases

## Step 1: Prepare Your Repository

### 1.1 Repository Structure

Ensure your repository has:
- ✅ `main.go` - Provider entry point
- ✅ `go.mod` - Go module file
- ✅ `README.md` - Provider documentation
- ✅ `LICENSE` - License file (MIT, Apache 2.0, etc.)
- ✅ `docs/` directory (optional but recommended)

### 1.2 Update Repository Settings

1. Go to your GitHub repository settings
2. Ensure the repository is **public**
3. Add topics: `terraform`, `terraform-provider`, `bugx`, `kubernetes`

## Step 2: Create GoReleaser Configuration

Create a `.goreleaser.yml` file in the root of your repository:

```yaml
project_name: terraform-provider-bugx

builds:
  - id: provider
    binary: terraform-provider-bugx
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64
    flags:
      - -trimpath
    ldflags:
      - -s -w
      - -X github.com/behrooz/terraform-provider-bugx/version.Version={{.Version}}
    env:
      - CGO_ENABLED=0

archives:
  - id: default
    builds:
      - provider
    format_overrides:
      - goos: windows
        format: zip
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: '{{ .ProjectName }}_{{ .Version }}_SHA256SUMS'
  algorithm: sha256

signs:
  - artifacts: checksum
    signature: "${artifact}.sig"

release:
  github:
    owner: behrooz  # Your GitHub username
    name: terraform-provider-bugx
  mode: replace
  header: |
    ## Terraform Provider bugx {{ .Tag }}
    
    See the [CHANGELOG.md](CHANGELOG.md) for details.
  footer: |
    ## Installation
    
    ```hcl
    terraform {
      required_providers {
        bugx = {
          source  = "behrooz/bugx"
          version = "{{ .Tag }}"
        }
      }
    }
    ```

terraform_registry:
  owner: behrooz  # Your GitHub username
  namespace: behrooz  # Usually same as owner
  name: bugx
  version: "{{ .Version }}"
  os: "{{ .Os }}"
  arch: "{{ .Arch }}"
  shasum: "{{ .Sha256 }}"
  filename: "{{ .ArtifactName }}"
```

## Step 3: Set Up GPG Signing

### 3.1 Generate GPG Key

```bash
# Generate a new GPG key
gpg --full-generate-key

# Follow the prompts:
# - Key type: RSA and RSA (default)
# - Key size: 4096
# - Expiration: 0 (no expiration) or set a date
# - Name: Your Name
# - Email: your-email@example.com
# - Passphrase: (choose a strong passphrase)

# List your keys
gpg --list-secret-keys --keyid-format=long

# Export your public key
gpg --armor --export YOUR_KEY_ID > public-key.gpg

# Export your private key (keep this secure!)
gpg --export-secret-keys --armor YOUR_KEY_ID > private-key.gpg
```

### 3.2 Add GPG Key to GitHub

1. Copy your public key:
   ```bash
   gpg --armor --export YOUR_KEY_ID
   ```

2. Go to GitHub → Settings → SSH and GPG keys
3. Click "New GPG key"
4. Paste your public key and save

### 3.3 Configure Git to Use GPG

```bash
# Tell Git about your GPG key
git config --global user.signingkey YOUR_KEY_ID

# Enable commit signing (optional)
git config --global commit.gpgsign true
```

## Step 4: Create GitHub Actions Workflow

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write
  id-token: write

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Import GPG key
        id: import_gpg
        uses: crazy-max/ghaction-import-gpg@v6
        with:
          gpg_private_key: ${{ secrets.GPG_PRIVATE_KEY }}
          passphrase: ${{ secrets.GPG_PASSPHRASE }}
          git_user_signingkey: true
          git_commit_gpgsign: true

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### 4.1 Add GitHub Secrets

1. Go to your repository → Settings → Secrets and variables → Actions
2. Add the following secrets:
   - `GPG_PRIVATE_KEY`: Your exported private GPG key
   - `GPG_PASSPHRASE`: Your GPG key passphrase

## Step 5: Create CHANGELOG.md

Create a `CHANGELOG.md` file:

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2025-01-XX

### Added
- Initial release
- Cluster management (create, read, update, delete)
- Helm release management
- Secret management (CRUD operations)
- Data sources for querying clusters
- Retry logic with exponential backoff
- Configurable timeouts and retry settings
```

## Step 6: Update README.md

Ensure your README includes:

1. **Provider description**
2. **Installation instructions** (will be auto-generated by registry)
3. **Usage examples**
4. **Requirements**
5. **Resources and data sources documentation**

## Step 7: Create Your First Release

### 7.1 Update Version

Create a `version/version.go` file:

```go
package version

var Version = "dev"
```

Update your `main.go` to include version info (optional):

```go
package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/behrooz/terraform-provider-bugx/version"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: Provider,
		ProviderAddr: "registry.terraform.io/behrooz/bugx",
		Version:      version.Version,
	})
}
```

### 7.2 Create and Push Tag

```bash
# Make sure all changes are committed
git add .
git commit -m "Prepare for v1.0.0 release"

# Create and push the tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0 
```

### 7.3 GitHub Actions Will Automatically:

1. Build binaries for all platforms
2. Create checksums
3. Sign the release
4. Create a GitHub release
5. Generate `terraform-registry-manifest.json`

## Step 8: Publish to Terraform Registry

### 8.1 Submit Your Provider

1. Go to [Terraform Registry](https://registry.terraform.io/)
2. Click "Publish" → "Provider"
3. Enter your GitHub repository URL: `https://github.com/behrooz/terraform-provider-bugx`
4. Click "Publish Provider"

### 8.2 Registry Requirements

The registry will verify:
- ✅ Repository is public
- ✅ Repository name follows `terraform-provider-{name}` pattern
- ✅ Has a valid `go.mod` file
- ✅ Has at least one release with proper structure
- ✅ Release includes `terraform-registry-manifest.json`
- ✅ Binaries are signed and checksummed

### 8.3 After Publishing

Once published, users can use your provider like this:

```hcl
terraform {
  required_providers {
    bugx = {
      source  = "behrooz/bugx"
      version = "~> 1.0"
    }
  }
}

provider "bugx" {
  base_url = "http://localhost:5173"
  username = "admin"
  password = "admin"
}
```

## Step 9: Maintain Your Provider

### 9.1 Versioning

Follow [Semantic Versioning](https://semver.org/):
- **MAJOR** (1.0.0): Breaking changes
- **MINOR** (0.1.0): New features, backward compatible
- **PATCH** (0.0.1): Bug fixes, backward compatible

### 9.2 Creating New Releases

1. Update `CHANGELOG.md`
2. Update version in `version/version.go`
3. Commit changes
4. Create and push new tag: `git tag -a v1.1.0 -m "Release v1.1.0"`
5. Push tag: `git push origin v1.1.0`
6. GitHub Actions will automatically create the release

## Troubleshooting

### Common Issues

1. **Registry not detecting release**
   - Ensure `terraform-registry-manifest.json` is in the release
   - Check that the release tag follows semantic versioning (v1.0.0, not 1.0.0)

2. **GPG signing fails**
   - Verify GPG key is added to GitHub
   - Check that secrets are correctly set in GitHub Actions

3. **Binary not found for platform**
   - Verify GoReleaser config includes all needed platforms
   - Check build logs in GitHub Actions

4. **Provider not appearing in registry**
   - Wait a few minutes after release (registry syncs periodically)
   - Verify repository is public
   - Check that you've submitted the provider to the registry

## Additional Resources

- [Terraform Registry Provider Publishing](https://developer.hashicorp.com/terraform/registry/providers/publishing)
- [GoReleaser Documentation](https://goreleaser.com/)
- [Terraform Provider Development](https://developer.hashicorp.com/terraform/plugin)
- [Semantic Versioning](https://semver.org/)

## Quick Start Checklist

- [ ] Repository is public on GitHub
- [ ] Repository name is `terraform-provider-bugx`
- [ ] Has `LICENSE` file
- [ ] Has `README.md` with documentation
- [ ] Has `.goreleaser.yml` configuration
- [ ] Has `.github/workflows/release.yml` workflow
- [ ] GPG key generated and added to GitHub
- [ ] GitHub secrets configured (GPG_PRIVATE_KEY, GPG_PASSPHRASE)
- [ ] `CHANGELOG.md` created
- [ ] `version/version.go` created
- [ ] First release tag created (v1.0.0)
- [ ] Provider submitted to Terraform Registry

Good luck with publishing your provider! 🚀

