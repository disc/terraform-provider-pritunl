terraform {
  required_providers {
    pritunl = {
      version = "~> 0.0.1"
      source  = "localhost/disc/pritunl"
    }
  }
}

provider "pritunl" {
  url    = var.pritunl_url
  token  = var.pritunl_api_token
  secret = var.pritunl_api_secret
}

resource "pritunl_organization" "alice" {
  name = "AliceOrg"
}

resource "pritunl_organization" "default" {
  name = "Default"
}

resource "pritunl_organization" "my-first-org" {
  name = "My_First_Org"
}

resource "pritunl_organization" "my-second-org" {
  name = "My_Second_Org"
}

data "pritunl_route" "kibana-route" {
  network = "1.1.1.3/32"
  comment = "Kibana"
  nat     = true
}

data "pritunl_route" "grafana-route" {
  network = "1.2.3.4/32"
  comment = "Grafana"
  nat     = false
}

data "pritunl_route" "global-route" {
  network = "0.0.0.0"
  comment = "internet"
  nat     = true
}

resource "pritunl_server" "main-server" {
  name     = "My_Main_Server4"
  protocol = "udp"
  port     = 55555
  cipher   = "aes128"
  hash     = "sha1"

  organizations = [
    pritunl_organization.my-second-org,
    pritunl_organization.my-first-org,
  ]

  routes = [
    data.pritunl_route.kibana-route,
    data.pritunl_route.grafana-route,
  ]
}

resource "pritunl_server" "default" {
  name     = "Main"
  protocol = "tcp"
  port     = 12444
  cipher   = "aes128"
  hash     = "sha1"

  organizations = [
    pritunl_organization.default,
  ]

  //  routes = [
  //    pritunl_route.global-route,
  //  ]

}