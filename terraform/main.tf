terraform {
  required_providers {
    hashicups = {
      version = "~> 0.3.1"
      source  = "hashicorp.com/edu/hashicups"
    }
  }
}

provider "hashicups" {
  username = "disc"
  password = "disc"
}

resource "hashicups_order" "edu" {
  items {
    coffee {
      id = 3
    }
    quantity = 2
  }
  items {
    coffee {
      id = 2
    }
    quantity = 2
  }
}

output "edu_order" {
  value = hashicups_order.edu
}


provider "pritunl" {
  baseUrl = "https://connect.cydriver.com"
  token = "rv2xqPtDiszTLN7IUsMooDXbpYZ7AAiC"
  secret = "Oq3FeJCa7hBSVD13We39GnVEty86toTI"
}

resource "pritunl_organization" "demo_org" {
  name = "organization name"
  description = "organization desc"
}

//
//resource "pritunl_organization" "stripchat" {
//  items {
//    coffee {
//      id = 3
//    }
//    quantity = 2
//  }
//  items {
//    coffee {
//      id = 2
//    }
//    quantity = 2
//  }
//}
//
//output "stripchat_org" {
//  value = pritunl_organization.stripchat
//}
