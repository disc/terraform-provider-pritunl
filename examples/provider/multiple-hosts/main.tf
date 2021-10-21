terraform {
  required_providers {
    pritunl = {
      version = "0.1.0"
      source  = "disc/pritunl"
    }
  }
}

provider "pritunl" {
  url    = var.pritunl_url
  token  = var.pritunl_api_token
  secret = var.pritunl_api_secret

  insecure = var.pritunl_insecure
}

data "pritunl_host" "main" {
  hostname = "nyc1.vpn.host"
}

data "pritunl_host" "reserve" {
  hostname = "nyc3.vpn.host"
}

resource "pritunl_server" "test" {
  name    = "some-server"
  network = "192.168.250.0/24"
  port    = 15500

  host_ids = [
    data.pritunl_host.main.id,
    data.pritunl_host.reserve.id,
  ]
}
