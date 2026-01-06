# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
# Build provider with debug flags (outputs to local Terraform plugins directory)
make build

# Run acceptance tests (requires Docker - spins up Pritunl server container)
make test

# Generate documentation from provider schemas
go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
```

## Testing

Tests require Docker. The `make test` command:
1. Starts a Pritunl container with MongoDB
2. Configures API credentials via MongoDB script
3. Runs acceptance tests with `TF_ACC=1`
4. Cleans up the container

To run a single test:
```bash
TF_ACC=1 \
PRITUNL_URL="https://localhost/" \
PRITUNL_INSECURE="true" \
PRITUNL_TOKEN=tfacctest_token \
PRITUNL_SECRET=tfacctest_secret \
go test -v -run TestAccOrganization ./internal/provider
```

## Architecture

This is a Terraform provider for Pritunl VPN Server using the Terraform Plugin SDK v2.

### Package Structure
- `internal/pritunl/` - API client library with HMAC-SHA512 authentication
- `internal/provider/` - Terraform provider implementation (resources, data sources)

### Resources and Data Sources
- **Resources**: `pritunl_organization`, `pritunl_server`, `pritunl_user`
- **Data Sources**: `pritunl_host`, `pritunl_hosts`

### Key Patterns

**API Client**: Interface-based design (`pritunl.Client`) with custom HTTP transport for HMAC authentication. All provider operations cast `meta` to `pritunl.Client`.

**Resource Structure**: Standard SDK v2 CRUD pattern with `CreateContext`, `ReadContext`, `UpdateContext`, `DeleteContext` functions. All resources support import via `StateContext`.

**Server Resource Complexity**: The server resource ([resource_server.go](internal/provider/resource_server.go)) is the most complex with 40+ fields, nested route blocks, organization/host attachments, and custom route state matching logic to handle reordering.

**User Import Format**: Users are imported with composite ID: `${organization_id}-${user_id}`

### Provider Configuration
Required: `url`, `token`, `secret`, `insecure`
Optional: `connection_check` (default: true) - validates API credentials on init
