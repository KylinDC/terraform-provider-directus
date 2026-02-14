# GitHub Actions CI/CD Setup Guide

This guide explains how to configure GitHub Actions for the Directus Terraform Provider.

## Workflow Overview

The CI/CD pipeline includes:

1. **Lint**: Code quality checks with golangci-lint
2. **Unit Tests**: Run all unit tests with race detection and coverage
3. **E2E Tests**: Full integration tests with real Directus instance
4. **Build**: Cross-platform builds (Linux, macOS, Windows)
5. **Release**: Automated releases on version tags
6. **Publish**: Publish to Terraform Registry via GoReleaser

## Required GitHub Secrets

To enable full CI/CD functionality, configure these secrets in your repository settings:

### For Terraform Registry Publishing

1. **GPG_PRIVATE_KEY**
   - Your GPG private key for signing releases
   - Required for Terraform Registry
   - Generate: `gpg --armor --export-secret-keys your@email.com`

2. **GPG_PASSPHRASE**
   - Passphrase for your GPG key
   - Required to unlock the private key during signing

3. **GPG_FINGERPRINT**
   - Your GPG key fingerprint
   - Get it: `gpg --list-secret-keys --keyid-format LONG`
   - Example: `ABCD1234EFGH5678IJKL90MNOP123456`

### Optional Secrets

4. **CODECOV_TOKEN** (Optional)
   - Token for uploading coverage to Codecov
   - Get from: https://codecov.io
   - The workflow continues even if this fails

## Setting Up Secrets

### Navigate to Repository Settings
```
GitHub Repository → Settings → Secrets and variables → Actions → New repository secret
```

### Generate GPG Key (First Time Setup)

If you don't have a GPG key:

```bash
# Generate a new GPG key
gpg --full-generate-key

# Use these settings:
# - Key type: RSA and RSA
# - Key size: 4096 bits
# - Expiration: 0 (no expiration) or your preference
# - Name: Your name
# - Email: your@email.com

# List your keys to get the fingerprint
gpg --list-secret-keys --keyid-format LONG

# Export private key (use the email from above)
gpg --armor --export-secret-keys your@email.com > private-key.asc

# Copy the content of private-key.asc to GPG_PRIVATE_KEY secret
cat private-key.asc

# Get fingerprint for GPG_FINGERPRINT secret
gpg --list-secret-keys --keyid-format LONG
# Look for the 40-character hex string after "sec"
```

### Export Public Key for Terraform Registry

```bash
# Export public key
gpg --armor --export your@email.com > public-key.asc

# You'll need to upload this to:
# https://registry.terraform.io/settings/gpg-keys
```

## Workflow Triggers

### Automatic Triggers

**On Push to Main/Develop:**
- Runs lint, unit tests, E2E tests, and builds
- Does NOT create releases

**On Pull Request:**
- Runs lint and unit tests
- Provides quick feedback on code quality

**On Version Tag (v*):**
- Runs full CI/CD pipeline
- Creates GitHub release
- Publishes to Terraform Registry
- Example: `git tag v0.1.0 && git push origin v0.1.0`

### Manual Trigger

You can also manually trigger workflows from the Actions tab.

## Workflow Jobs Explained

### 1. Lint Job
- Runs `golangci-lint` for code quality
- Checks formatting, unused code, potential bugs
- Fast feedback (< 1 minute)

### 2. Unit Test Job
- Runs all Go unit tests
- Enables race detector (`-race`)
- Generates coverage report
- Uploads to Codecov (if configured)
- Artifacts: `coverage.out`, `coverage.html`

### 3. E2E Test Job
- Starts PostgreSQL and Directus services
- Waits for services to be healthy
- Creates authentication token
- Runs comprehensive E2E tests
- Tests full CRUD lifecycle
- Validates relationships

### 4. Build Job
- Builds binaries for multiple platforms:
  - Linux (amd64, arm64)
  - macOS (amd64, arm64)
  - Windows (amd64)
- Creates artifacts for each platform
- Requires lint and test to pass first

### 5. Release Job
- Only runs on version tags (`v*`)
- Downloads all build artifacts
- Generates SHA256 checksums
- Creates GitHub release with:
  - Release notes (auto-generated)
  - All platform binaries
  - Checksum file

### 6. Publish Registry Job
- Only runs on version tags
- Uses GoReleaser to:
  - Sign artifacts with GPG
  - Create Terraform Registry compatible release
  - Generate proper manifest

## Creating a Release

### Step 1: Prepare Release

```bash
# Ensure you're on main branch
git checkout main
git pull origin main

# Run tests locally
make test
./scripts/test-e2e-comprehensive.sh

# Update version in relevant files if needed
# Update CHANGELOG.md with changes
```

### Step 2: Create and Push Tag

```bash
# Create annotated tag
git tag -a v0.1.0 -m "Release v0.1.0"

# Push tag to trigger release
git push origin v0.1.0
```

### Step 3: Monitor Workflow

1. Go to GitHub Actions tab
2. Watch the workflow run
3. Check for any failures
4. Once complete, verify release in Releases tab

### Step 4: Verify Release

```bash
# Check release page
https://github.com/kylindc/terraform-provider-directus/releases

# Verify artifacts:
# - Binaries for each platform
# - SHA256SUMS file
# - SHA256SUMS.sig (GPG signature)
```

## Terraform Registry Setup

### Prerequisites

1. **GPG Key**: Must be added to Terraform Registry
2. **GitHub Repository**: Must be public
3. **Repository Name**: Must match `terraform-provider-*` format

### Steps

1. Go to https://registry.terraform.io
2. Sign in with GitHub
3. Navigate to "Publish" → "Provider"
4. Connect your repository
5. Add GPG public key in settings
6. Trigger a release by pushing a version tag

### Registry Requirements

The provider must have:
- ✅ Repository name: `terraform-provider-directus`
- ✅ Valid `go.mod` and Go code
- ✅ At least one version tag (v0.1.0+)
- ✅ Signed release with GPG key
- ✅ GoReleaser configuration (`.goreleaser.yml`)
- ✅ Documentation in `docs/` directory (optional but recommended)

## Troubleshooting

### GPG Signing Fails

**Error**: `gpg: signing failed: No secret key`

**Solution**:
- Verify GPG_PRIVATE_KEY secret is correct
- Verify GPG_PASSPHRASE is correct
- Check GPG_FINGERPRINT matches your key

```bash
# Test locally
gpg --list-secret-keys --keyid-format LONG
```

### E2E Tests Fail

**Error**: Services not healthy

**Solution**:
- GitHub Actions uses service containers
- Wait time might need adjustment
- Check Directus logs in workflow output

**Error**: Token creation fails

**Solution**:
- Verify Directus admin credentials
- Check Directus version compatibility
- Review API endpoint in workflow

### Build Fails

**Error**: Import cycle or compilation errors

**Solution**:
```bash
# Test locally
go mod tidy
go build ./...
```

### Release Not Created

**Error**: No release appears after pushing tag

**Solution**:
- Verify tag format: `v*` (e.g., v0.1.0)
- Check all jobs passed
- Review workflow logs
- Ensure GITHUB_TOKEN has proper permissions

## Monitoring and Badges

### Add Status Badges

Add to your README.md:

```markdown
[![CI](https://github.com/kylindc/terraform-provider-directus/actions/workflows/ci.yml/badge.svg)](https://github.com/kylindc/terraform-provider-directus/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/kylindc/terraform-provider-directus/branch/main/graph/badge.svg)](https://codecov.io/gh/kylindc/terraform-provider-directus)
[![Go Report Card](https://goreportcard.com/badge/github.com/kylindc/terraform-provider-directus)](https://goreportcard.com/report/github.com/kylindc/terraform-provider-directus)
```

## Local Testing

### Test GoReleaser Locally

```bash
# Install GoReleaser
brew install goreleaser

# Test release process (no push)
goreleaser release --snapshot --clean

# Check dist/ directory for artifacts
ls -lh dist/
```

### Test GitHub Actions Locally

```bash
# Install act
brew install act

# Run specific job
act -j test

# Run full workflow
act push
```

## Best Practices

1. **Always test locally** before pushing tags
2. **Use semantic versioning** (v0.1.0, v1.0.0, etc.)
3. **Update CHANGELOG.md** before releases
4. **Review workflow logs** for each release
5. **Test provider** after each release
6. **Keep secrets secure** - never commit them
7. **Rotate GPG keys** periodically for security

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GoReleaser Documentation](https://goreleaser.com)
- [Terraform Registry Publishing](https://www.terraform.io/docs/registry/providers/publishing.html)
- [GPG Documentation](https://gnupg.org/documentation/)

## Support

If you encounter issues:
1. Check workflow logs in GitHub Actions tab
2. Review this setup guide
3. Test locally with GoReleaser
4. Open an issue with workflow logs attached
