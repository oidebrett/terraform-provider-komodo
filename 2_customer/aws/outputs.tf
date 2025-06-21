# Output the public IP address of the instance
output "instance_public_ip" {
  description = "The public IP address of the AWS EC2 instance"
  value       = aws_instance.client_instance.public_ip
}

# Output the instance ID
output "instance_id" {
  description = "The ID of the AWS EC2 instance"
  value       = aws_instance.client_instance.id
}

# Output the instance name
output "instance_name" {
  description = "The name tag of the AWS EC2 instance"
  value       = aws_instance.client_instance.tags.Name
}

# Output the availability zone
output "instance_availability_zone" {
  description = "The availability zone where the instance is deployed"
  value       = aws_instance.client_instance.availability_zone
}

# Output the security group ID
output "security_group_id" {
  description = "The ID of the created security group"
  value       = aws_security_group.client_sg.id
}

# Output the security group name
output "security_group_name" {
  description = "The name of the created security group"
  value       = aws_security_group.client_sg.name
}

# Output the key pair name (if created)
output "key_pair_name" {
  description = "The name of the AWS key pair (if created)"
  value       = var.ssh_public_key != "" ? aws_key_pair.client_key[0].key_name : "No key pair created"
  sensitive   = true
}

# Output the client configuration ID
output "client_id" {
  description = "The client ID used in the configuration"
  value       = myuserprovider_user.client_syncresources.id
}

# Output the client name
output "client_name" {
  description = "The client name used in the configuration"
  value       = myuserprovider_user.client_syncresources.name
}