# SSH Keys Implementation - Updated with Outputs

## Overview

The Terraform Provider Komodo now supports SSH key generation with **dual access methods**:

1. **Embedded in file contents** (original functionality)
2. **Available as Terraform outputs** (NEW!)

## New Features

### SSH Key Outputs
SSH keys are now available as computed attributes that can be accessed via Terraform outputs:

- `ssh_private_key` - The generated private key (marked as sensitive)
- `ssh_public_key` - The generated public key

### Dual Access Pattern
Users can now access SSH keys in two ways:

#### Method 1: From File Contents (Embedded)
Keys are automatically embedded in the `environment = """..."""` section:
```
SSH_PRIVATE_KEY=-----BEGIN OPENSSH PRIVATE KEY-----\n...\n-----END OPENSSH PRIVATE KEY-----
SSH_PUBLIC_KEY=ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI...
```

#### Method 2: From Terraform Outputs (NEW)
Keys are available as resource attributes:
```hcl
output "ssh_private_key" {
  description = "Generated SSH private key for repository access"
  value       = komodo-provider_user.example.ssh_private_key
  sensitive   = true
}

output "ssh_public_key" {
  description = "Generated SSH public key (also uploaded as deploy key)"
  value       = komodo-provider_user.example.ssh_public_key
}
```

## Usage Examples

### Complete Configuration with Outputs

```hcl
resource "komodo-provider_user" "client" {
  id               = "client-001"
  name             = "MyClient"
  server_ip        = "1.2.3.4"
  generate_ssh_keys = true
  file_contents    = templatefile("config.toml", {
    domain = "example.com"
    email  = "admin@example.com"
  })
}

# Access SSH keys as outputs
output "ssh_private_key" {
  description = "SSH private key for repository access"
  value       = komodo-provider_user.client.ssh_private_key
  sensitive   = true
}

output "ssh_public_key" {
  description = "SSH public key (also uploaded as deploy key)"
  value       = komodo-provider_user.client.ssh_public_key
}
```

### Accessing Keys from Command Line

```bash
# Get the private key
terraform output -raw ssh_private_key

# Get the public key  
terraform output -raw ssh_public_key

# Save private key to file for SSH usage
terraform output -raw ssh_private_key > ~/.ssh/repo_key
chmod 600 ~/.ssh/repo_key

# Use the key to clone the repository
git clone git@github.com:owner/repo.git
```

## Implementation Details

### Schema Changes
Added new computed attributes to the resource schema:

```go
"ssh_private_key": tfschema.StringAttribute{
    MarkdownDescription: "The generated SSH private key (only available when generate_ssh_keys is true)",
    Computed:            true,
    Sensitive:           true,
},
"ssh_public_key": tfschema.StringAttribute{
    MarkdownDescription: "The generated SSH public key (only available when generate_ssh_keys is true)",
    Computed:            true,
},
```

### Model Updates
Extended the `KomodoModel` struct:

```go
type KomodoModel struct {
    Id               tftypes.String `tfsdk:"id"`
    Name             tftypes.String `tfsdk:"name"`
    FileContents     tftypes.String `tfsdk:"file_contents"`
    ServerIP         tftypes.String `tfsdk:"server_ip"`
    GenerateSSHKeys  tftypes.Bool   `tfsdk:"generate_ssh_keys"`
    SSHPrivateKey    tftypes.String `tfsdk:"ssh_private_key"`  // NEW
    SSHPublicKey     tftypes.String `tfsdk:"ssh_public_key"`   // NEW
}
```

### Function Signature Updates
Modified functions to return SSH keys:

```go
// Before
func (r *komodoResource) createGitHubRepository(ctx context.Context, repoName string, fileContents string, generateSSHKeys bool) error

// After  
func (r *komodoResource) createGitHubRepository(ctx context.Context, repoName string, fileContents string, generateSSHKeys bool) (string, string, error)
```

## Testing

### Verify Outputs in Plan
```bash
terraform plan
```

Should show:
```
+ ssh_private_key   = (sensitive value)
+ ssh_public_key    = (known after apply)
```

### Test SSH Key Access
```bash
# Apply configuration
terraform apply

# Verify outputs are available
terraform output ssh_public_key
terraform output -raw ssh_private_key | head -1  # Should show "-----BEGIN OPENSSH PRIVATE KEY-----"
```

## Benefits

### For Users
1. **Direct Access**: Get SSH keys without parsing file contents
2. **Automation**: Use keys in other Terraform resources or scripts
3. **Security**: Private keys are properly marked as sensitive
4. **Flexibility**: Choose between embedded or output access methods

### For Integration
1. **CI/CD Pipelines**: Easily extract keys for automated deployments
2. **Key Management**: Programmatically access keys for rotation
3. **Multi-Resource**: Use generated keys in other Terraform resources

## Migration

### Existing Users
- No changes required - existing functionality remains unchanged
- SSH keys continue to be embedded in file contents
- New output functionality is additive

### New Users
- Can choose to use outputs, embedded keys, or both
- Outputs provide cleaner access pattern for automation
- Embedded keys still useful for application configuration

## Security Notes

1. **Sensitive Marking**: Private keys are marked as sensitive in Terraform state
2. **State Storage**: Keys are stored in Terraform state - ensure state is secured
3. **Output Handling**: Use `-raw` flag to avoid JSON escaping when extracting keys
4. **File Permissions**: Set proper permissions (600) when saving keys to files
