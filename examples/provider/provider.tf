terraform {
  required_providers {
    pritunl = {
      version = "~> 0.0.1"
      source  = "disc/pritunl"
    }
  }
}

provider "pritunl" {
  url    = "https://localhost"
  token  = "api-token"
  secret = "api-secret"

  insecure         = false
  connection_check = true
}

resource "pritunl_organization" "developers" {
  name = "Developers"
}

resource "pritunl_organization" "admins" {
  name = "Admins"
}

resource "pritunl_user" "test" {
  name            = "test-user"
  organization_id = pritunl_organization.developers.id
  email           = "test@test.com"
  groups = [
    "admins",
  ]
}

resource "pritunl_server" "test" {
  name = "test"

  organization_ids = [
    pritunl_organization.developers.id,
    pritunl_organization.admins.id,
  ]

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

  route {
    network = "10.3.0.0/32"
    comment = "Private network #3"
    nat     = false
    net_gateway = true
  }

}
