# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2025-01-XX

### Added
- Initial release of terraform-provider-bugx
- Cluster management: Create, read, update, and delete bugx instances
- Helm release management: Deploy and manage Helm charts on bugxs
- Secret management: Full CRUD operations for secrets via REST API
- Data sources: Query existing clusters without managing them
- Retry logic: Automatic retry with exponential backoff for transient network errors
- Configurable timeouts: Customizable HTTP client timeouts and retry settings
- Resource import: Import existing clusters into Terraform state
- Chart version support: Pin specific Helm chart versions for reproducible deployments

### Resources
- `bugx_cluster` - Manage bugx instances
- `bugx_helm_release` - Deploy Helm charts on bugxs
- `bugx_orphan_cleanup` - Clean up orphaned resources
- `bugx_secret` - Manage secrets via REST API

### Data Sources
- `bugx_cluster` - Query existing clusters

