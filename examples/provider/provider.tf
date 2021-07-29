provider "pritunl" {
  url      = var.pritunl_url        # optionally use PRITUNL_URL env var
  token    = var.pritunl_api_token  # optionally use PRITUNL_TOKEN env var
  secret   = var.pritunl_api_secret # optionally use PRITUNL_SECRET env var
  insecure = var.pritunl_insecure   # optionally use PRITUNL_INSECURE env var
}