# Crunchloop Terraform Provider

[![License](https://img.shields.io/badge/license-MPL--2.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

The Crunchloop Terraform Provider enables you to manage your Crunchloop resources using Terraform.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 1.0+
- [Go](https://golang.org/doc/install) 1.16+

## Installation

To install the provider, copy and paste the code below into your Terraform configuration. Then, run `terraform init` to initialize the provider.

```hcl
terraform {
  required_providers {
    crunchloop = {
      source  = "crunchloop/crunchloop"
      version = "0.1.0"
    }
  }
}

provider "crunchloop" {
  url = "http://localhost:3000"
}
```

## Usage

Here is an example of how to use the provider to manage a Crunchloop resource:

```hcl
data "crunchloop_vmi" "ubuntu" {
  name = "ubuntu-jammy-server-amd64-20241002"
}

resource "crunchloop_vm" "vm" {
  name                       = "terraform-test"
  vmi_id                     = data.crunchloop_vmi.ubuntu.id
  cores                      = 1
  memory_megabytes           = 1024
  root_volume_size_gigabytes = 10
}
```

## Developing the Provider

If you wish to contribute to the provider, follow these steps:

1. Clone the repository
2. Build the provider using Go: `go build ./...`

## Documentation

- [Terraform Documentation](https://registry.terraform.io/providers/crunchloop/crunchloop/latest/docs)

## License

This project is licensed under the Mozilla Public License 2.0.
