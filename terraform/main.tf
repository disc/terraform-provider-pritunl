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

data "pritunl_route" "default" {
  network = "0.0.0.0/0"
  nat     = true
}

//resource "pritunl_server" "main" {
//  name     = "Main"
//  protocol = "tcp"
//  port     = 12444
//  network  = "192.168.218.0/24"
//  status   = "online"
//
//  organizations = [
//    pritunl_organization.default,
//    pritunl_organization.alice,
//  ]
//
//  //  route {
//  //    network = "0.0.0.0/0"
//  //    nat     = true
//  //  }
//}

resource "pritunl_server" "test" {
  name = "test"

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