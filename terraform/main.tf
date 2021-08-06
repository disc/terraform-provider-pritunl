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
  groups = [
    "admins",
    "users",
  ]
  dns_servers = [
    "8.8.8.8",
    "1.1.1.1",
  ]
  network_wg = "192.168.5.0/24"
  port_wg    = 44444
  otp_auth   = true
  //  ipv6               = true
  dh_param_bits = 1024
  ping_interval = 15
  //  ping_timeout       = 60
  link_ping_interval = 2
  //  link_ping_timeout  = 20
  //  inactive_timeout   = 600
  max_clients  = 20
  network_mode = "tunnel" // bridge
  //  network_start = "192.168.1.200" // is required for network_mode = bridge
  //  network_end   = "192.168.1.240" // is required for network_mode = bridge
  //  mss_fix = 1450
  //  max_devices = 5
  //  pre_connect_msg = "Hello world"
  //  allowed_devices = "mobile"
  //  search_domain = "abc.org,dot.com"
  replica_count = 3

  //  multi_device = true
  //  debug = true
  //  restrict_routes = true
  //  block_outside_dns = true
  //  dns_mapping = true
  //  inter_client = true
  //  vxlan = true

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