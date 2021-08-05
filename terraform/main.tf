terraform {
  required_providers {
    pritunl = {
      version = "~> 0.0.1"
      source  = "disc/pritunl"
    }
  }
}

provider "pritunl" {
  url    = var.pritunl_url
  token  = var.pritunl_api_token
  secret = var.pritunl_api_secret
}

resource "pritunl_organization" "my-first-org" {
  name = "My_First_Org"
}

resource "pritunl_organization" "my-second-org" {
  name = "My_Second_Org"
}

resource "pritunl_server" "test" {
  name    = "test"
  port    = 65500
  network = "192.168.1.0/24"

  organizations = [
    pritunl_organization.my-first-org,
  ]

  dynamic "route" {
    for_each = var.common_routes
    content {
      network = route.value["network"]
      comment = route.value["comment"]
      nat     = route.value["nat"]
    }
  }
}