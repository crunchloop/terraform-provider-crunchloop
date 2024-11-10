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

# You can be assign a host directly to a VM
#
resource "crunchloop_vm" "with_host" {
  name                       = "terraform-with-host"
  vmi_id                     = data.crunchloop_vmi.ubuntu.id
  host_id                    = data.crunchloop_host.test-1.id
  cores                      = 1
  memory_megabytes           = 1024
  root_volume_size_gigabytes = 10
  user_data                  = base64encode("echo 'Hello, World!'")
}

# Or the can let the system allocate the host for you
#
resource "crunchloop_vm" "without_host" {
  name                       = "terraform-without-host"
  vmi_id                     = data.crunchloop_vmi.ubuntu.id
  cores                      = 1
  memory_megabytes           = 1024
  root_volume_size_gigabytes = 10
}

# Use a cloud-init configuration to configure the VM
#
data "cloudinit_config" "cloudinit" {
  gzip          = false
  base64_encode = true

  part {
    filename     = "cloud-config.yaml"
    content_type = "text/cloud-config"

    content = <<-EOF
    #cloud-config
    cloud_final_modules:
      - [scripts-user, always]

    runcmd:
      - echo "Hello, World!" > /tmp/hello.txt
    EOF
  }
}

resource "crunchloop_vm" "with_cloudinit" {
  name                       = "terraform-with-cloudinit"
  vmi_id                     = data.crunchloop_vmi.ubuntu.id
  cores                      = 1
  memory_megabytes           = 1024
  root_volume_size_gigabytes = 10
  user_data                  = data.cloudinit_config.cloudinit.rendered
}