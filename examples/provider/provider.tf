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

resource "pritunl_user" "test_pin" {
  name            = "test-user-pin"
  organization_id = pritunl_organization.developers.id
  email           = "test@test.com"
  pin             = "123456"
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

# Override global Pritunl settings
resource "pritunl_settings_override" "main" {
  theme            = "dark"
  pin_mode         = "optional"
  client_reconnect = true
  restrict_client  = false
  reverse_proxy    = true
  server_port      = 443
  username         = "admin"
}

# Settings override with email configuration
resource "pritunl_settings_override" "email" {
  email_from     = "vpn@example.com"
  email_server   = "smtp.example.com"
  email_username = "vpn@example.com"
  email_password = "secret"
}

# Settings override with InfluxDB monitoring
resource "pritunl_settings_override" "monitoring" {
  monitoring      = "influxdb"
  influxdb_url    = "https://influxdb.example.com"
  influxdb_token  = "my-token"
  influxdb_org    = "my-org"
  influxdb_bucket = "pritunl"
}

# Network settings
resource "pritunl_settings_override" "network" {
  public_address    = "vpn.example.com"
  public_address6   = "2001:db8::1"
  routed_subnet6    = "2001:db8:1::/48"
  routed_subnet6_wg = "2001:db8:2::/48"
  ipv6              = true
}

# SSL/TLS settings
resource "pritunl_settings_override" "ssl" {
  server_cert = file("${path.module}/certs/server.crt")
  server_key = file("${path.module}/certs/server.key")
  acme_domain = "vpn.example.com"
}

# User restrictions
resource "pritunl_settings_override" "restrictions" {
  restrict_import  = true
  restrict_client  = true
  drop_permissions = true
}

# SSO with Okta + Duo
resource "pritunl_settings_override" "sso_okta_duo" {
  sso = "saml_okta_duo"

  sso_settings {
    default_organization_id = pritunl_organization.developers.id
    cache                   = true
    client_cache            = true

    saml {
      url        = "https://example.okta.com/app/pritunl/sso/saml"
      issuer_url = "http://www.okta.com/exk1234567890"
      cert = file("${path.module}/certs/okta-saml.crt")
    }

    okta {
      token  = "okta-api-token"
      app_id = "0oa1234567890"
      mode   = "push"
    }

    duo {
      token  = "DINTEGRATION_KEY"
      secret = "DUO_SECRET_KEY"
      host   = "api-12345678.duosecurity.com"
      mode   = "push"
    }
  }
}

# SSO with Azure AD
resource "pritunl_settings_override" "sso_azure" {
  sso = "azure"

  sso_settings {
    default_organization_id = pritunl_organization.developers.id

    azure {
      app_id       = "00000000-0000-0000-0000-000000000000"
      app_secret   = "azure-app-secret"
      directory_id = "00000000-0000-0000-0000-000000000001"
      region       = "us"
      version      = 2
    }
  }
}

# SSO with Google
resource "pritunl_settings_override" "sso_google" {
  sso = "google"

  sso_settings {
    default_organization_id = pritunl_organization.developers.id

    google {
      domain = "example.com"
      email  = "admin@example.com"
      private_key = file("${path.module}/certs/google-service-account.json")
    }
  }
}

# SSO with SAML
resource "pritunl_settings_override" "sso_saml" {
  sso = "saml"

  sso_settings {
    default_organization_id = pritunl_organization.developers.id

    saml {
      url        = "https://idp.example.com/sso/saml"
      issuer_url = "https://idp.example.com/metadata"
      cert = file("${path.module}/certs/saml.crt")
    }
  }
}

# SSO with Radius
resource "pritunl_settings_override" "sso_radius" {
  sso = "radius"

  sso_settings {
    default_organization_id = pritunl_organization.developers.id

    radius {
      host   = "radius.example.com:1812"
      secret = "radius-shared-secret"
    }
  }
}

# SSO with JumpCloud
resource "pritunl_settings_override" "sso_jumpcloud" {
  sso = "jumpcloud"

  sso_settings {
    default_organization_id = pritunl_organization.developers.id

    jumpcloud {
      app_id = "jumpcloud-app-id"
      secret = "jumpcloud-secret"
    }
  }
}

# Cloud Provider — AWS
resource "pritunl_settings_override" "cloud_aws" {
  cloud_provider = "aws"

  cloud_provider_aws_settings {
    route53_region = "us-east-1"
    route53_zone   = "aws"

    us_east_1_access_key = "AKIAIOSFODNN7EXAMPLE"
    us_east_1_secret_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  }
}

# Cloud Provider — Oracle
resource "pritunl_settings_override" "cloud_oracle" {
  cloud_provider = "oracle"

  cloud_provider_oracle_settings {
    oracle_user_ocid = "ocid1.user.oc1..aaaaaaaaexample"
  }
}

# Cloud Provider — Pritunl Cloud
resource "pritunl_settings_override" "cloud_pritunl" {
  cloud_provider = "pritunl"

  cloud_provider_pritunl_settings {
    host   = "https://cloud.pritunl.com"
    token  = "pritunl-cloud-token"
    secret = "pritunl-cloud-secret"
  }
}
