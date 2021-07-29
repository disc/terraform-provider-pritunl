variable "pritunl_url" {
  type    = string
  default = "http://localhost"
}

variable "pritunl_api_token" {
  type    = string
  default = "secret"
}

variable "pritunl_api_secret" {
  type    = string
  default = "secret"
}

variable "common_routes" {
  type    = list(map(any))
  default = []
}

variable "pritunl_insecure" {
  type    = bool
  default = false
}