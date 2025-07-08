output "server_ip" {
  description = "IPv4 address of the server"
  value       = hcloud_server.web.ipv4_address
}

output "server_status" {
  description = "Status of the server"
  value       = hcloud_server.web.status
}
