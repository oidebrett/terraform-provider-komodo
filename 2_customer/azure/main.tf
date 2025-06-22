# Terraform configuration
terraform {
  required_version = ">= 1.0"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "=3.0.0"
    }
#    myuserprovider = {
#      source = "example.com/me/myuserprovider"
#    }
  }
}

# Azure Provider
provider "azurerm" {
  features {
    resource_group {
      prevent_deletion_if_contains_resources = false
    }
  }
  subscription_id = var.azure_subscription_id
  client_id       = var.azure_client_id
  client_secret   = var.azure_client_secret
  tenant_id       = var.azure_tenant_id
}

# Custom User Provider
#provider "myuserprovider" {
#  endpoint     = var.myuserprovider_endpoint
#  github_token = var.github_token
#}

# Resource Group
resource "azurerm_resource_group" "client_rg" {
  name     = "${var.client_name}-resources"
  location = var.azure_location
}

# Virtual Network
resource "azurerm_virtual_network" "client_vnet" {
  name                = "${var.client_name}-network"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.client_rg.location
  resource_group_name = azurerm_resource_group.client_rg.name
  
  timeouts {
    create = "30m"
    delete = "30m"
  }
}

# Subnet
resource "azurerm_subnet" "client_subnet" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.client_rg.name
  virtual_network_name = azurerm_virtual_network.client_vnet.name
  address_prefixes     = ["10.0.2.0/24"]
  
  depends_on = [
    azurerm_virtual_network.client_vnet
  ]
  
  timeouts {
    create = "30m"
    delete = "30m"
  }
}

# Network Security Group
resource "azurerm_network_security_group" "client_nsg" {
  name                = "${var.client_name}-nsg"
  location            = azurerm_resource_group.client_rg.location
  resource_group_name = azurerm_resource_group.client_rg.name

  security_rule {
    name                       = "SSH"
    priority                   = 1001
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "22"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "HTTP"
    priority                   = 1002
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "80"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "HTTPS"
    priority                   = 1003
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "443"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "CustomHTTP"
    priority                   = 1004
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "8120"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "CustomAPI"
    priority                   = 1005
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "9120"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
}

# Associate NSG with Subnet
resource "azurerm_subnet_network_security_group_association" "client_subnet_nsg_association" {
  subnet_id                 = azurerm_subnet.client_subnet.id
  network_security_group_id = azurerm_network_security_group.client_nsg.id
}

# Network Interface
resource "azurerm_network_interface" "client_nic" {
  name                = "${var.client_name}-nic"
  location            = azurerm_resource_group.client_rg.location
  resource_group_name = azurerm_resource_group.client_rg.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.client_subnet.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.client_public_ip.id
  }
  
  # Add explicit dependency
  depends_on = [
    azurerm_subnet.client_subnet,
    azurerm_public_ip.client_public_ip
  ]
}

# Associate NSG with Network Interface
resource "azurerm_network_interface_security_group_association" "client_nic_nsg_association" {
  network_interface_id      = azurerm_network_interface.client_nic.id
  network_security_group_id = azurerm_network_security_group.client_nsg.id
}

# Public IP
resource "azurerm_public_ip" "client_public_ip" {
  name                = "${var.client_name}-public-ip"
  location            = azurerm_resource_group.client_rg.location
  resource_group_name = azurerm_resource_group.client_rg.name
  allocation_method   = "Static"
  sku                 = "Standard"  # Change from Basic to Standard
}

# Linux Virtual Machine
resource "azurerm_linux_virtual_machine" "client_vm" {
  name                = var.instance_name
  resource_group_name = azurerm_resource_group.client_rg.name
  location            = azurerm_resource_group.client_rg.location
  size                = var.vm_size
  admin_username      = var.ssh_username
  network_interface_ids = [
    azurerm_network_interface.client_nic.id,
  ]

  admin_ssh_key {
    username   = var.ssh_username
    public_key = var.ssh_public_key
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "18.04-LTS"
    version   = "latest"
  }

  custom_data = base64encode(templatefile("${path.module}/startup-script.sh", {}))

  tags = {
    environment = "production"
    client      = var.client_name
  }
}

# Configure Komodo
#resource "myuserprovider_config" "client_config" {
#  name       = "${var.client_name}_config"
#  content    = templatefile("${path.module}/config-template.toml", {
#    client_name        = var.client_name
#    client_name_lower  = lower(var.client_name)
#    domain             = var.domain
#    admin_email        = var.admin_email
#    admin_username     = var.admin_username
#    admin_password     = var.admin_password
#    admin_subdomain    = var.admin_subdomain
#    crowdsec_enrollment_key = var.crowdsec_enrollment_key
#    postgres_user      = var.postgres_user
#    postgres_password  = var.postgres_password
#    postgres_host      = var.postgres_host
#    static_page        = upper(tostring(var.static_page))
#    oauth_client_id    = var.oauth_client_id
#    oauth_client_secret = var.oauth_client_secret
#    github_repo        = var.github_repo
#  })
#  server_ip = azurerm_public_ip.client_public_ip.ip_address
#}

# Output the public IP
output "instance_public_ip" {
  value = azurerm_public_ip.client_public_ip.ip_address
}

# Output SSH command
output "ssh_command" {
  value = "ssh ${var.ssh_username}@${azurerm_public_ip.client_public_ip.ip_address}"
}
