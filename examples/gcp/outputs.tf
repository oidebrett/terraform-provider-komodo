# Output the public IP address of the instance
output "instance_public_ip" {
  description = "The public IP address of the GCP compute instance"
  value       = google_compute_instance.gcp_vm.network_interface[0].access_config[0].nat_ip
}

# Output the instance name
output "instance_name" {
  description = "The name of the GCP compute instance"
  value       = google_compute_instance.gcp_vm.name
}

# Output the instance zone
output "instance_zone" {
  description = "The zone where the instance is deployed"
  value       = google_compute_instance.gcp_vm.zone
}

# Output the firewall rule name
output "firewall_rule_name" {
  description = "The name of the created firewall rule"
  value       = google_compute_firewall.allow_ports.name
}

# Output the client configuration ID
output "client_id" {
  description = "The client ID used in the configuration"
  value       = komodo-provider_user.client_syncresources.id
}

# Output the client name
output "client_name" {
  description = "The client name used in the configuration"
  value       = komodo-provider_user.client_syncresources.name
}