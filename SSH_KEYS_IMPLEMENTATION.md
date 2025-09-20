# SSH Keys Implementation for Terraform Provider Komodo

## Overview

This implementation adds optional SSH key generation and management to the Terraform Provider Komodo. When creating GitHub repositories with file contents and `generate_ssh_keys = true`, the provider:

1. Generates an ed25519 SSH key pair
2. Uploads the public key as a deploy key to the GitHub repository
3. Embeds both private and public keys in the file contents

The SSH key generation is **optional** and controlled by the `generate_ssh_keys` parameter.

## Implementation Details

### New Functions Added

#### `generateSSHKeyPair() (string, string, error)`
- Generates an ed25519 SSH key pair using Go's crypto libraries
- Returns the private key in OpenSSH format and public key in SSH authorized_keys format
- Uses secure random number generation from `crypto/rand`

#### `uploadDeployKey(ctx, owner, repoName, publicKey, title, readOnly) error`
- Uploads the generated public key as a deploy key to the GitHub repository
- Uses the GitHub API v3 to create the deploy key
- Provides read/write access by default (readOnly=false)

#### `addSSHKeysToFileContents(fileContents, privateKey, publicKey) string`
- Parses the file contents to find the `environment = """..."""` section
- Adds SSH_PRIVATE_KEY and SSH_PUBLIC_KEY environment variables
- Properly escapes newlines in the private key for TOML format

### Modified Functions

#### `createGitHubRepository()`
- Now accepts a `generateSSHKeys` boolean parameter
- Conditionally generates SSH keys based on the parameter
- Uploads the public key as a deploy key only when SSH keys are generated
- Embeds both keys in the file contents only when SSH keys are generated

#### `updateFileInRepository()`
- Now accepts a `generateSSHKeys` boolean parameter
- Preserves existing SSH keys when updating file contents (if SSH generation is enabled)
- Generates new SSH keys only if none exist and SSH generation is enabled
- Maintains backward compatibility with existing repositories

### Schema Changes

#### New Field: `generate_ssh_keys`
- Type: `Bool`
- Optional: `true`
- Description: "Whether to generate SSH keys and upload them as deploy keys to the GitHub repository"
- When `true`: SSH keys are generated and embedded
- When `false` or unset: No SSH key generation occurs

### Import Changes

Added new imports for SSH key generation:
```go
import (
    "crypto/ed25519"
    "crypto/rand"
    "encoding/pem"
    mathrand "math/rand"  // Aliased to avoid conflict with crypto/rand
    "golang.org/x/crypto/ssh"
)
```

## Usage Examples

### With SSH Key Generation Enabled

When you set `generate_ssh_keys = true` and provide file contents like this:

```toml
[[stack]]
name = "pangolin-setup"
[stack.config]
server = "server-xxzQw"
repo = "name/repo"
file_paths = ["docker-compose-setup.yml"]
environment = """
DOMAIN=example.com
EMAIL=me@example.com
ADMIN_USERNAME=admin@example.com
ADMIN_PASSWORD=Password!
"""
```

The provider automatically transforms it to:

```toml
[[stack]]
name = "pangolin-setup"
[stack.config]
server = "server-xxzQw"
repo = "name/repo"
file_paths = ["docker-compose-setup.yml"]
environment = """
DOMAIN=example.com
EMAIL=me@example.com
ADMIN_USERNAME=admin@example.com
ADMIN_PASSWORD=Password!
SSH_PRIVATE_KEY=-----BEGIN OPENSSH PRIVATE KEY-----\n...\n-----END OPENSSH PRIVATE KEY-----
SSH_PUBLIC_KEY=ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI...
"""
```

### With SSH Key Generation Disabled

When you set `generate_ssh_keys = false` or omit the parameter, the file contents remain unchanged:

```hcl
resource "komodo-provider_user" "example" {
  id               = "1"
  name             = "MyClient"
  server_ip        = "1.2.3.4"
  generate_ssh_keys = false  # or omit this line
  file_contents    = "..."
}
```

The file contents will be created exactly as provided, without any SSH key modifications.

## Security Features

1. **ed25519 Keys**: Uses modern, secure ed25519 cryptography
2. **Private Repository**: GitHub repositories are created as private by default
3. **Deploy Key**: Public key is uploaded as a deploy key for secure repository access
4. **Key Preservation**: Existing SSH keys are preserved during updates

## Testing

Added comprehensive tests in `internal/provider/ssh_test.go`:

- `TestGenerateSSHKeyPair`: Verifies SSH key generation functionality
- `TestAddSSHKeysToFileContents`: Verifies proper embedding of keys in file contents

## Example Usage

A complete example is provided in `examples/ssh_keys/` directory with:

- `main.tf`: Terraform configuration
- `variables.tf`: Variable definitions
- `config-template.toml`: Template file with environment section
- `terraform.tfvars.example`: Example configuration values
- `outputs.tf`: Output definitions
- `README.md`: Detailed usage instructions

## Backward Compatibility

This implementation maintains full backward compatibility:

- Existing repositories without SSH keys continue to work unchanged
- File contents without environment sections are handled gracefully
- Updates to existing repositories preserve existing SSH keys

## Benefits

1. **Automated Security**: No manual SSH key management required
2. **Secure Access**: Deploy keys provide secure repository access
3. **Easy Integration**: SSH keys are automatically available in the application environment
4. **No Manual Steps**: Fully automated process from Terraform apply

## Future Enhancements

Potential future improvements could include:

1. Key rotation functionality
2. Support for different key types (RSA, ECDSA)
3. Custom deploy key permissions
4. Key expiration management
