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


//output "first_organization_id" {
//  value = pritunl_organization.my-first-org.id
//}
//
//output "first_organization_name" {
//  value = pritunl_organization.my-first-org.name
//}
//
//output "second_organization_id" {
//  value = my-second-org.name
//}
