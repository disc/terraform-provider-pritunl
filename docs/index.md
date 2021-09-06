---
page_title: "Provider: Pritunl"
subcategory: ""
description: |- 
  Terraform provider for interacting with Pritunl API.
---

# Pritunl Provider

## Example Usage

```terraform
provider "pritunl" {
  url      = "https://vpn.server.com"
  token    = "api-token"
  secret   = "api-secret-key"
  insecure = false
}
```

## Schema

### Optional

- **insecure** (Boolean)
- **secret** (String)
- **token** (String)
- **url** (String)
