# Debug Testing Guide for SSH Key Generation

## Problem
SSH keys are not being generated even when `generate_ssh_keys = true` is set.

## Debug Steps

### 1. Build the Provider with Debug Logging

```bash
cd /home/ivob/Projects/terraform-provider-komodo
go build -o terraform-provider-komodo
```

### 2. Set Up Local Provider for Testing

Create a local provider configuration to use your built binary:

```bash
# Create terraform.d directory in your test folder
mkdir -p ~/.terraform.d/plugins/registry.example.com/mattercoder/komodo-provider/1.0.0/linux_amd64/

# Copy your built provider
cp terraform-provider-komodo ~/.terraform.d/plugins/registry.example.com/mattercoder/komodo-provider/1.0.0/linux_amd64/

# Make it executable
chmod +x ~/.terraform.d/plugins/registry.example.com/mattercoder/komodo-provider/1.0.0/linux_amd64/terraform-provider-komodo
```

### 3. Use the Debug Test Configuration

```bash
cd examples/debug_test
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your actual values
```

### 4. Run Terraform with Debug Output

```bash
# Initialize
terraform init

# Plan with debug output
TF_LOG=DEBUG terraform plan

# Apply with debug output
TF_LOG=DEBUG terraform apply
```

### 5. Check Debug Output

Look for these debug messages in the output:
- `DEBUG: GenerateSSHKeys is: true`
- `DEBUG: Generating SSH keys...`
- `DEBUG: Generated SSH keys - private key length: X, public key length: Y`
- `DEBUG: Uploading deploy key 'terraform-deploy-key-XXXXX' to repo owner/repo`
- `DEBUG: Successfully uploaded deploy key`
- `DEBUG: Adding SSH keys to file contents`

### 6. Alternative: Use Development Override

If the above doesn't work, create a development override:

```bash
# Create dev_overrides.tfrc
cat > ~/.terraformrc << EOF
provider_installation {
  dev_overrides {
    "registry.example.com/mattercoder/komodo-provider" = "/home/ivob/Projects/terraform-provider-komodo"
  }
  direct {}
}
EOF
```

Then run terraform normally - it will use your local binary.

### 7. Check GitHub Repository

After a successful apply:
1. Go to your GitHub repository
2. Check Settings > Deploy keys - you should see a new deploy key
3. Check the resources.toml file - it should contain SSH_PRIVATE_KEY and SSH_PUBLIC_KEY

## Common Issues

### Issue 1: Provider Not Using Local Binary
**Symptom**: No debug output appears
**Solution**: Use development override method above

### Issue 2: GitHub Token Permissions
**Symptom**: Error uploading deploy key
**Solution**: Ensure GitHub token has `repo` permissions

### Issue 3: Repository Not Found
**Symptom**: Error creating repository
**Solution**: Check GitHub organization name and token permissions

### Issue 4: File Contents Not Updated
**Symptom**: SSH keys generated but not in file
**Solution**: Check if environment section exists in config template

## Manual Verification

You can also test the SSH key generation function directly:

```bash
cd /home/ivob/Projects/terraform-provider-komodo
go test ./internal/provider -v -run TestGenerateSSHKeyPair
```

This will show you if the SSH key generation itself is working.
