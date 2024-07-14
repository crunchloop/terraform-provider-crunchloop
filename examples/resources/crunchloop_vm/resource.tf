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

data "crunchloop_host" "test-1" {
  name = "Test Host 1"
}

data "crunchloop_vmi" "ubuntu" {
  name = "Ubuntu 24.04 (noble)"
}

resource "crunchloop_vm" "default" {
  name                       = "terraform-vm"
  vmi_id                     = data.crunchloop_vmi.ubuntu.id
  host_id                    = data.crunchloop_host.test-1.id
  cores                      = 1
  memory_megabytes           = 1024
  root_volume_size_gigabytes = 10
  user_data                  = "echo 'Hello, World!'"
}
