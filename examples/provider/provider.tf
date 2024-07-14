terraform {
  required_providers {
    crunchloop = {
      source = "bilby91/crunchloop"
    }
  }
}

provider "crunchloop" {
  url = "http://localhost:3000"
}
