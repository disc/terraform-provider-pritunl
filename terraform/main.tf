terraform {
  required_providers {
    pritunl = {
      version = "~> 0.0.1"
      source  = "localhost/disc/pritunl"
    }
  }
}

provider "pritunl" {
  url = "https://connect.cydriver.com"
  token = "rv2xqPtDiszTLN7IUsMooDXbpYZ7AAiC"
  secret = "Oq3FeJCa7hBSVD13We39GnVEty86toTI"
}

resource "pritunl_organization" "my-first-org" {
  name = "My_First_Org"
}

resource "pritunl_organization" "my-second-org" {
  name = "My_Second_Org"
}

resource "pritunl_route" "kibana-route" {
  network = "1.1.1.3/32"
  comment = "Kibana4"
  nat = true
}

resource "pritunl_route" "grafana-route" {
  network = "1.2.3.4/32"
  comment = "Grafana2"
  nat = false
}

resource "pritunl_server" "main-server" {
  name = "My_Main_Server2"
  protocol = "tcp"
  port = 54321
  organizations = [
    pritunl_organization.my-first-org.id,
    pritunl_organization.my-second-org.id,
  ]
  routes = [
    pritunl_route.kibana-route,
    pritunl_route.grafana-route
  ]
  cipher = "aes128"
  hash = "sha1"
}