<a href="https://pritunl.com">
    <img src="https://pritunl.com/img/logo.png" alt="Pritunl logo" title="Pritunl" align="right" height="100" />
</a>
<a href="https://terraform.io">
    <img src="https://dashboard.snapcraft.io/site_media/appmedia/2019/11/terraform.png" alt="Terraform logo" title="Terraform" align="right" height="100" />
</a>

# Terraform Provider for Pritunl VPN Server

[![Release](https://img.shields.io/github/v/release/disc/terraform-provider-pritunl)](https://github.com/disc/terraform-provider-pritunl/releases)
[![Registry](https://img.shields.io/badge/registry-doc%40latest-lightgrey?logo=terraform)](https://registry.terraform.io/providers/disc/pritunl/latest/docs)
[![License](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://github.com/disc/terraform-provider-pritunl/blob/master/LICENSE)  
[![Go Report Card](https://goreportcard.com/badge/github.com/disc/terraform-provider-pritunl)](https://goreportcard.com/report/github.com/disc/terraform-provider-pritunl)

- Website: https://www.terraform.io
- Pritunl VPN Server: https://pritunl.com/
- Provider: [disc/pritunl](https://registry.terraform.io/providers/disc/pritunl/latest)

## Requirements
-	[Terraform](https://www.terraform.io/downloads.html) >=0.13.x
-	[Go](https://golang.org/doc/install) 1.16.x (to build the provider plugin)

## Building The Provider

```sh
$ git clone git@github.com:disc/terraform-provider-pritunl
$ make build
```

## Example usage

Take a look at the examples in the [documentation](https://registry.terraform.io/providers/disc/pritunl/0.0.4/docs) of the registry
or use the following example:


```hcl
# Set the required provider and versions
terraform {
  required_providers {
    pritunl = {
      source  = "disc/pritunl"
      version = "0.0.4"
    }
  }
}

# Configure the pritunl provider
provider "pritunl" {
  url    = "https://vpn.server.com"
  token  = "api-token"
  secret = "api-secret"
  insecure = false
}

# Create a pritunl organization resource
resource "pritunl_organization" "developers" {
  name = "Developers"
}

# Create a pritunl user resource 
resource "pritunl_user" "steve" {
  name            = "steve"
  organization_id = pritunl_organization.developers.id
  email           = "steve@developers.com"
  groups = [
    "developers",
  ]
}

# Create a pritunl server resource
resource "pritunl_server" "example" {
  name      = "example"
  port      = 15500
  protocol  = "udp"
  network   = "192.168.1.0/24"
  groups    = [
    "admins",
    "developers",
  ]
  
  # Attach the organization to the server
  organization_ids = [
    pritunl_organization.developers.id,
  ]

  # Describe all the routes manually
  # Default route 0.0.0.0/0 will be deleted on the server creation
  route {
    network = "10.0.0.0/24"
    comment = "Private network #1"
    nat     = true
  }
  
  route {
    network = "10.2.0.0/24"
    comment = "Private network #2"
    nat     = false
  }
  
  # Or create dynamic routes from variables
  dynamic "route" {
    for_each = var.common_routes
    content {
        network = route.value["network"]
        comment = route.value["comment"]
        nat     = route.value["nat"]
      }
  }
}
```

## Importing exist resources

Describe exist resource in the terraform file first and then import them:

Import an organization:
```hcl
# Describe a pritunl organization resource
resource "pritunl_organization" "developers" {
  name = "Developers"
}
```

Execute the shell command:
```sh
terraform import pritunl_organization.developers ${ORGANIZATION_ID}
terraform import pritunl_organization.developers 610e42d2a0ed366f41dfe6e8
```
The organization ID (as well as other resource IDs) can be found in the Pritunl API responses or in the HTML document response.

Import a user:
```hcl
# Describe a pritunl user resource
resource "pritunl_user" "steve" {
  name            = "steve"
  organization_id = pritunl_organization.developers.id
  email           = "steve@developers.com"
}
```

Execute the shell command:
```sh
terraform import pritunl_user.steve ${ORGANIZATION_ID}-${USER_ID}
terraform import pritunl_user.steve 610e42d2a0ed366f41dfe6e8-610e42d6a0ed366f41dfe72b
```

Import a server:

```hcl
# Describe a pritunl server resource
resource "pritunl_server" "example" {
  name      = "example"
  port      = 15500
  protocol  = "udp"
  network   = "192.168.1.0/24"
  groups    = [
    "developers",
  ]

  # Attach the organization to the server
  organization_ids = [
    pritunl_organization.developers.id,
  ]

  # Describe all the routes manually
  # Default route 0.0.0.0/0 will be deleted on the server creation
  route {
    network = "10.0.0.0/24"
    comment = "Private network #1"
    nat     = true
  }
}
```

Execute the shell command:
```sh
terraform import pritunl_server.example ${SERVER_ID}
terraform import pritunl_server.example 60cd0bfa7723cf3c911468a8
```

## License

The Terraform Pritunl Provider is available to everyone under the terms of the Mozilla Public License Version 2.0. [Take a look the LICENSE file](LICENSE).