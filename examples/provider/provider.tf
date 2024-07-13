terraform {
  required_providers {
    crunchloop = {
      source = "registry.terraform.io/crunchloop/crunchloop"
    }
  }
}

provider "crunchloop" {
  url = "http://localhost:3000"
}
