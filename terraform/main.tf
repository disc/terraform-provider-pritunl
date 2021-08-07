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

  insecure = var.pritunl_insecure
}

// terraform import pritunl_user.test 610e42d2a0ed366f41dfe6e8-610e42d6a0ed366f41dfe72b

// 610e42d6a0ed366f41dfe72b
resource "pritunl_user" "test" {
  name            = "test-user"
  organization_id = pritunl_organization.test.id
  email           = "test@test.com"
  groups = [
    "admins",
  ]
  network_links = [
    "1.1.1.1/32",
    "1.1.1.0/24",
  ]
  client_to_client = true
  auth_type        = "local"
  mac_addresses = [
    "00:1B:44:11:3A:B6",
    "00:1B:44:11:3A:B7",
  ]
  dns_servers = [
    "1.1.1.1",
    "8.8.4.4",
  ]
  dns_suffix       = "host.com"
  bypass_secondary = true
  port_forwarding = [
    {
      dport : 8080,
      port : 80,
      protocol : "udp",
    },
    {
      dport : 21,
      port : 21,
      protocol : "tcp",
    },
  ]
}

//  610e42d2a0ed366f41dfe6e8
resource "pritunl_organization" "test" {
  name = "Test"
}

resource "pritunl_server" "test" {
  name     = "test"
  port     = 22222
  protocol = "tcp"
  network  = "192.168.1.0/24"
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
  dh_param_bits      = 1024
  ping_interval      = 15
  ping_timeout       = 60
  link_ping_interval = 2
  link_ping_timeout  = 20
  inactive_timeout   = 600
  max_clients        = 20
  network_mode       = "tunnel" // bridge
  //  network_start = "192.168.1.200" // is required for network_mode = bridge
  //  network_end   = "192.168.1.240" // is required for network_mode = bridge
  mss_fix         = 1450
  max_devices     = 5
  pre_connect_msg = "Hello world"
  allowed_devices = "mobile"
  search_domain   = "abc.org, dot.com"
  replica_count   = 3

  multi_device      = true
  debug             = true
  restrict_routes   = true
  block_outside_dns = true
  dns_mapping       = true
  inter_client      = true
  //  vxlan             = true // requires Premium subscription

  organizations = [
    pritunl_organization.test,
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