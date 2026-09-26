# Administrator accounts are the logins of the Pritunl web console and of its
# API. They have nothing to do with `pritunl_user`, which manages the VPN
# profiles of the end users inside an organization: an administrator is not a
# member of an organization and cannot connect to a VPN server.
#
# Every request behind this resource requires the credentials configured on the
# `pritunl` provider to belong to a super user, Pritunl refuses `/admin` to
# anybody else.

# A service account for another automation to authenticate with, the reason
# `auth_api` exists. Pritunl generates the credentials itself, there is nothing
# to configure and nothing to pick: `token` and `secret` are read-only outputs
# that only carry a value while `auth_api` is true.

resource "random_password" "ci" {
  length  = 32
  special = true
}

resource "pritunl_administrator" "ci" {
  username = "ci"
  password = random_password.ci.result

  # a service account that only needs the API, and needs to be a super user for
  # the endpoints that ask for one, `/admin` and `/settings` among them
  auth_api   = true
  super_user = true
}

# Handing the generated credentials over to whatever consumes them, here an
# Azure Key Vault, the same way the pritunl_settings examples source the TLS
# certificate from one.

data "azurerm_key_vault" "example" {
  name                = "examplekv"
  resource_group_name = "some-resource-group"
}

resource "azurerm_key_vault_secret" "ci_token" {
  name         = "pritunl-ci-token"
  value        = pritunl_administrator.ci.token
  key_vault_id = data.azurerm_key_vault.example.id
}

resource "azurerm_key_vault_secret" "ci_secret" {
  name         = "pritunl-ci-secret"
  value        = pritunl_administrator.ci.secret
  key_vault_id = data.azurerm_key_vault.example.id
}

# There is no attribute to rotate the credentials with: any value at all in the
# token or the secret field of a request makes Pritunl generate a fresh one and
# throw the value away, so an attribute would rotate them on every single apply.
# Rotating is replacing the account:
#
#   terraform apply -replace=pritunl_administrator.ci

# A human administrator, with a two-step authentication code from an
# authenticator application and a YubiKey. The password is write-only, it is
# never read back from the API, so it is only ever written when the
# configuration carries one.

resource "pritunl_administrator" "operator" {
  username = "operator"
  password = var.operator_password

  super_user = false
  otp_auth   = true

  # the public id of the key, the first twelve characters of an OTP it emits. A
  # whole OTP is accepted too and truncated to those twelve characters.
  yubikey_id = "ccccccbcdefg"
}

# Adopting the account the instance was seeded with, the one Pritunl creates on
# the first start, rather than creating another super user next to it. See
# import.sh for the import itself.
#
# Take care with this one in particular: it is usually the account the `pritunl`
# provider itself authenticates as, and `auth_api = false` or `disabled = true`
# on it revokes the credentials Terraform is holding, in the middle of the very
# apply that sets them.

resource "pritunl_administrator" "default" {
  username = "pritunl"

  super_user = true
  auth_api   = true
}
