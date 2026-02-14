# Publishing Guide: Deploy to GitHub & Terraform Registry

This guide walks you through publishing the Directus Terraform Provider to GitHub so others can discover, download, and use it.

> **Reference**: This guide is based on the official [HashiCorp Publishing Providers](https://developer.hashicorp.com/terraform/registry/providers/publishing) documentation.

There are **two distribution paths** covered here:

| Path | Users install via | GPG signing required? |
|---|---|---|
| **A. GitHub Releases** | Manual download from Releases page | No |
| **B. Terraform Registry** | `terraform init` (automatic) | Yes |

Most users will want to start with **Path A** and optionally add **Path B** later.

---

## Prerequisites

- [Git](https://git-scm.com/) installed
- A [GitHub account](https://github.com)
- [Go 1.21+](https://go.dev/dl/) installed
- (Optional) [GoReleaser](https://goreleaser.com/install/) for local testing
- (Optional, for Terraform Registry) A GPG key pair (**RSA or DSA** -- see note in Step 6)

---

## Step 1: Prepare the Provider for Publishing

Before publishing, ensure the provider meets the Terraform Registry's requirements.

### 1a. Repository Naming

The repository **must** be named `terraform-provider-directus`:

- The `terraform-provider-` prefix is **required** for the Terraform Registry to recognize it.
- **Only lowercase** repository names are supported.
- The repository must be **public** for the Terraform Registry (private works for GitHub Releases only).

### 1b. Terraform Registry Manifest File

The Terraform Registry requires a `terraform-registry-manifest.json` file at the root of the repository. This file is included in release assets and provides metadata about the provider.

```json
{
  "version": 1,
  "metadata": {
    "protocol_versions": ["6.0"]
  }
}
```

| Field | Purpose |
|---|---|
| `version` | Numeric version of the manifest format (always `1`). |
| `metadata.protocol_versions` | Supported Terraform protocol versions. Providers built with **Terraform Plugin Framework** should use `["6.0"]`. Providers built with **Plugin SDK v2** should use `["5.0"]`. |

> **Important**: The file must contain valid JSON syntax (no trailing commas).

### 1c. Provider Documentation

The Terraform Registry displays documentation from the `docs/` directory. This project already includes the full documentation set:

| Location | Filename | Description | Status |
|---|---|---|---|
| `docs/` | `index.md` | Provider overview, example usage, argument reference | Included |
| `docs/resources/` | `policy.md` | `directus_policy` resource | Included |
| `docs/resources/` | `role.md` | `directus_role` resource | Included |
| `docs/resources/` | `role_policies_attachment.md` | `directus_role_policies_attachment` resource | Included |
| `docs/resources/` | `collection.md` | `directus_collection` resource | Included |
| `docs/guides/` | `authentication.md` | Authentication & security guide | Included |
| `docs/data-sources/` | `<name>.md` | Documentation for each data source | Future |

At a minimum, a provider must contain `docs/index.md` and at least one resource or data source document.

Each document may include **YAML frontmatter** with optional `page_title` and `subcategory` attributes. Documents have a **500KB storage limit** on the Registry.

**Document format requirements**:

- **Index page** (`docs/index.md`): Must include provider summary, example usage, and argument reference.
- **Resource pages** (`docs/resources/*.md`): Must include description, example usage, argument reference, attribute reference, and import instructions.
- **Guides** (`docs/guides/*.md`): Must include a `page_title` in the frontmatter.
- **Callouts**: Use `->` for blue notes, `~>` for yellow notes, and `!>` for red warnings.

**Auto-generating docs**: Use [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs) to regenerate documentation from the provider schema:

```bash
# Add to your Go source
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

# Then run
go generate ./...
```

**Preview docs**: Use the [Terraform Registry Doc Preview Tool](https://registry.terraform.io/tools/doc-preview) to test how your docs will render before publishing.

For detailed formatting requirements, see [Provider Documentation](https://developer.hashicorp.com/terraform/registry/providers/docs).

---

## Step 2: Create the GitHub Repository

### Option A: Create via GitHub Web UI

1. Go to https://github.com/new
2. Set the repository name to **`terraform-provider-directus`**
3. Set visibility to **Public**.
4. Do **not** initialize with README, `.gitignore`, or license (we already have them locally).
5. Click **Create repository**.

### Option B: Create via GitHub CLI

```bash
gh repo create terraform-provider-directus --public --source=. --remote=origin
```

---

## Step 3: Push Your Code

```bash
# Navigate to the project root
cd /path/to/directus-terraform-provider

# Initialize git (if not already done)
git init -b main

# Add the remote (replace kylindc with your GitHub username)
git remote add origin https://github.com/kylindc/terraform-provider-directus.git

# Stage all files
git add .

# Make the initial commit
git commit -m "feat: initial release of directus terraform provider

- Policy resource (CRUD + import)
- Role resource with parent-child hierarchy
- Role-policy attachment resource (authoritative M2M)
- Collection resource with metadata
- Full E2E test suite
- CI/CD pipeline with GitHub Actions"

# Push to GitHub
git push -u origin main
```

After pushing, visit your repository on GitHub to confirm the code is there.

---

## Step 4: Verify CI Pipeline

The push to `main` will automatically trigger the **CI workflow** (`ci.yml`):

1. Go to your repository on GitHub.
2. Click the **Actions** tab.
3. You should see the **CI** workflow running.
4. Verify that the **Lint**, **Unit Tests**, **Build**, and **Acceptance Tests** jobs pass.
5. The Acceptance Tests run against a real Directus instance spun up by GitHub Actions services.

If any job fails, check the logs and fix the issue before proceeding to the release step.

> **Note**: Ensure [GitHub Organization settings](https://docs.github.com/en/organizations/managing-organization-settings/disabling-or-limiting-github-actions-for-your-organization) and [Repository settings](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/enabling-features-for-your-repository/managing-github-actions-settings-for-a-repository) allow running the workflows and actions.

---

## Step 5: Create a Release (Path A -- GitHub Releases)

This is the simplest way to distribute the provider. No GPG key needed.

### 5a. Version Tagging Requirements

The tag **must** be a valid [Semantic Version](https://semver.org/) preceded with `v` (e.g., `v1.2.3`).

- Prerelease versions are supported with a hyphen delimiter: `v1.2.3-pre` (available if explicitly defined but not chosen automatically by `terraform init`).
- **There must not be a branch name with the same name as the tag.**

### 5b. Tag the Release

```bash
# Make sure you're on main and up to date
git checkout main
git pull origin main

# Create an annotated tag
git tag -a v0.1.0 -m "v0.1.0 - Initial release"

# Push the tag
git push origin v0.1.0
```

### 5c. What Happens Automatically

Pushing a `v*` tag triggers the **Release workflow** (`.github/workflows/release.yml`):

1. **Lint** -- `go vet` and `go mod tidy` checks
2. **Unit Tests** -- `go test ./... -v -race`
3. **Acceptance Tests** -- Terraform acceptance tests against a real Directus v11.15 instance (PostgreSQL 16 + Directus via GitHub Actions services)
4. **GoReleaser** -- builds, signs, and publishes the release:
   - Cross-compiles binaries for Linux, macOS, and Windows (amd64 + arm64)
   - Creates a GitHub Release with:
   - Zip archives for every platform: `terraform-provider-directus_{VERSION}_{OS}_{ARCH}.zip`
   - Binary naming: `terraform-provider-directus_v{VERSION}`
   - SHA256 checksum file: `terraform-provider-directus_{VERSION}_SHA256SUMS`
   - GPG signature (if configured): `terraform-provider-directus_{VERSION}_SHA256SUMS.sig`
   - Registry manifest: `terraform-provider-directus_{VERSION}_manifest.json`
   - Auto-generated release notes

> **Important**: Avoid modifying or replacing an already-released version. This will cause checksum errors for users attempting to download the provider. Instead, release a new version.

### 5d. Verify the Release

1. Go to your repository's **Releases** page: `https://github.com/kylindc/terraform-provider-directus/releases`
2. You should see `v0.1.0` with:
   - Release notes
   - Platform zip archives
   - SHA256 checksum file
   - Manifest JSON file

---

## Step 6: How Others Can Use It

### Option 1: Manual Installation (from GitHub Releases)

Users download the binary for their platform and install it locally:

```bash
# 1. Download the binary for your platform
# Replace VERSION, OS, ARCH with actual values
# OS: linux, darwin, windows
# ARCH: amd64, arm64
VERSION="0.1.0"
OS="darwin"
ARCH="arm64"

curl -Lo terraform-provider-directus.zip \
  "https://github.com/kylindc/terraform-provider-directus/releases/download/v${VERSION}/terraform-provider-directus_${VERSION}_${OS}_${ARCH}.zip"

# 2. Unzip
unzip terraform-provider-directus.zip

# 3. Create the plugin directory
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/kylindc/directus/${VERSION}/${OS}_${ARCH}

# 4. Move the binary
mv terraform-provider-directus_v${VERSION} \
  ~/.terraform.d/plugins/registry.terraform.io/kylindc/directus/${VERSION}/${OS}_${ARCH}/terraform-provider-directus

# 5. Make it executable
chmod +x ~/.terraform.d/plugins/registry.terraform.io/kylindc/directus/${VERSION}/${OS}_${ARCH}/terraform-provider-directus
```

Then in Terraform configuration:

```hcl
terraform {
  required_providers {
    directus = {
      source  = "kylindc/directus"
      version = "0.1.0"
    }
  }
}

provider "directus" {
  endpoint = "https://your-directus-instance.com"
  token    = var.directus_token
}
```

### Option 2: Install from Terraform Registry (if published -- see Step 7)

After publishing to the registry, users simply add this to their Terraform config:

```hcl
terraform {
  required_providers {
    directus = {
      source  = "kylindc/directus"
      version = "~> 0.1"
    }
  }
}
```

Then:

```bash
terraform init   # automatically downloads the provider
```

---

## Step 7: Publish to Terraform Registry (Path B -- Optional)

The [Terraform Registry](https://registry.terraform.io) is the official distribution channel. Users get automatic installation via `terraform init`, and Terraform will verify GPG signatures during download.

### 7a. Requirements

Before you begin:

- [ ] Repository is **public** on GitHub
- [ ] Repository name is `terraform-provider-directus` (lowercase, with `terraform-provider-` prefix)
- [ ] You have at least one release tag (e.g., `v0.1.0`)
- [ ] `terraform-registry-manifest.json` exists at the root of the repository
- [ ] Provider documentation exists in `docs/` directory
- [ ] You have a GPG key pair for signing releases

### 7b. Preparing and Adding a Signing Key

All provider releases are required to be signed. You must provide HashiCorp with the **public key** for the GPG keypair you will use to sign releases. The Terraform Registry validates the signature when publishing each version, and `terraform init` verifies it during download.

> **Important**: The Terraform Registry accepts **RSA and DSA** keys, but **not the default ECC type**. When generating your key, explicitly choose RSA.

#### Generate a GPG Key

```bash
# Generate a new GPG key
gpg --full-generate-key
# Choose:
#   - Key type: RSA and RSA (default) -- DO NOT use ECC
#   - Key size: 4096
#   - Expiration: 0 (no expiration) or your preference
#   - Real name: Your name
#   - Email: your-email@example.com
#   - Passphrase: choose a strong passphrase

# List your keys to get the key ID
gpg --list-secret-keys --keyid-format LONG
# Output example:
# sec   rsa4096/ABCDEF1234567890 2024-01-01 [SC]
#       1234567890ABCDEF1234567890ABCDEF12345678

# Export the ASCII-armored PRIVATE key (for GitHub Secrets)
gpg --armor --export-secret-keys your-email@example.com > gpg-private-key.asc

# Export the ASCII-armored PUBLIC key (for Terraform Registry)
gpg --armor --export your-email@example.com > gpg-public-key.asc
```

Refer to [GitHub's detailed GPG key instructions](https://docs.github.com/en/github/authenticating-to-github/generating-a-new-gpg-key) for more help. You do **not** need to add the key to GitHub -- only to the Terraform Registry.

### 7c. Add GitHub Secrets

Go to your repository: **Settings > Secrets and variables > Actions > New repository secret**

Add these two secrets (names match the [official scaffolding workflow](https://github.com/hashicorp/terraform-provider-scaffolding-framework/blob/main/.github/workflows/release.yml)):

| Secret Name | Value | How to Get |
|---|---|---|
| `GPG_PRIVATE_KEY` | Content of `gpg-private-key.asc` | `cat gpg-private-key.asc` |
| `PASSPHRASE` | Your GPG key passphrase | The passphrase you chose |

> **Note**: The GPG fingerprint is automatically derived from the imported key by the CI workflow -- you do not need to add it as a separate secret.

### 7d. Upload Public Key to Terraform Registry

1. Go to https://registry.terraform.io
2. Click **Sign In** and authenticate with GitHub
   - The GitHub account must have appropriate permission scopes on the provider repository. Verify at [GitHub Settings > Authorized OAuth Apps](https://github.com/settings/applications) under the Terraform Registry Application.
3. Navigate to [User Settings > Signing Keys](https://registry.terraform.io/settings/gpg-keys)
4. Click **Add GPG Key**
5. Paste the content of `gpg-public-key.asc`
6. Save
   - You can add keys for your personal namespace or any organization where you are an admin.

### 7e. Publish the Provider

1. Go to https://registry.terraform.io
2. Click [**Publish > Provider**](https://registry.terraform.io/publish/provider)
3. Select your GitHub repository (`terraform-provider-directus`)
4. Follow the prompts to connect it

Publishing creates a **webhook** on your GitHub repository subscribed to `release` events. Future versions released will automatically notify the Terraform Registry, which will ingest them.

> **Webhook Troubleshooting**: If the webhook is missing or not functioning, go to the provider's settings page on the Terraform Registry, remove any existing webhooks for `registry.terraform.io` from your GitHub repo settings, then click the **Resync** button on the Registry.

### 7f. Trigger a Signed Release

Once secrets are configured, push a new tag to trigger a signed release:

```bash
git tag -a v0.1.1 -m "v0.1.1 - First registry release"
git push origin v0.1.1
```

The Release workflow will:
1. Run all tests (unit + acceptance) as a quality gate
2. Build cross-platform binaries via GoReleaser
3. Sign the SHA256 checksum file with your GPG key
4. Include the registry manifest in the release
5. Create a GitHub Release
6. The Terraform Registry detects the new release automatically via the webhook

### 7g. Verify on Terraform Registry

After a few minutes, your provider should appear at:

```
https://registry.terraform.io/providers/kylindc/directus/latest
```

### 7h. Terms of Use

Anything published to the Terraform Registry is subject to the [Terms of Use](https://registry.terraform.io/terms).

### 7i. Support

If you experience issues publishing to the Terraform Registry, contact [terraform-registry@hashicorp.com](mailto:terraform-registry@hashicorp.com).

---

## Release Asset Naming Conventions

The Terraform Registry expects specific naming conventions for release assets. GoReleaser handles this automatically, but here's the reference for manual builds:

| Asset | Naming Pattern |
|---|---|
| Zip archive | `terraform-provider-directus_{VERSION}_{OS}_{ARCH}.zip` |
| Binary (inside zip) | `terraform-provider-directus_v{VERSION}` |
| Checksum file | `terraform-provider-directus_{VERSION}_SHA256SUMS` |
| Signature file | `terraform-provider-directus_{VERSION}_SHA256SUMS.sig` |
| Manifest file | `terraform-provider-directus_{VERSION}_manifest.json` |

### Recommended OS / Architecture Combinations

| OS | Architectures |
|---|---|
| `linux` | `amd64`, `arm64`, `arm`, `386` |
| `darwin` | `amd64`, `arm64` |
| `windows` | `amd64`, `386` |
| `freebsd` | `amd64`, `arm64`, `arm`, `386` |

See [HashiCorp's recommended OS/arch list](https://developer.hashicorp.com/terraform/registry/providers/os-arch) for the full list.

---

## Quick Reference: Release Checklist

Use this checklist every time you create a new release:

```
[ ] All tests pass locally: go test ./...
[ ] Acceptance tests pass: TF_ACC=1 go test ./internal/provider/ -run TestAcc
[ ] E2E tests pass: ./scripts/test-e2e-comprehensive.sh
[ ] Code is committed and pushed to main
[ ] terraform-registry-manifest.json is present and valid
[ ] docs/ directory is complete:
    [ ] docs/index.md — provider overview with example usage and argument reference
    [ ] docs/resources/policy.md
    [ ] docs/resources/role.md
    [ ] docs/resources/role_policies_attachment.md
    [ ] docs/resources/collection.md
    [ ] docs/guides/authentication.md
[ ] Documentation renders correctly in the Doc Preview Tool
[ ] CHANGELOG or release notes are prepared
[ ] No branch exists with the same name as the tag
[ ] Tag created and pushed: git tag -a vX.Y.Z -m "message" && git push origin vX.Y.Z
[ ] Release workflow passes (check GitHub Actions tab)
[ ] Release artifacts appear on the Releases page:
    [ ] Zip archives for each platform
    [ ] SHA256SUMS file
    [ ] SHA256SUMS.sig (if GPG configured)
    [ ] Manifest JSON file
[ ] (If registry) Provider version appears on Terraform Registry
```

---

## Updating an Existing Release

To publish a new version after making changes:

```bash
# 1. Make your changes
# 2. Run tests
go test ./...
TF_ACC=1 DIRECTUS_ENDPOINT="http://localhost:8055" DIRECTUS_TOKEN="your-token" \
  go test ./internal/provider/ -run TestAcc -timeout 30m
./scripts/test-e2e-comprehensive.sh

# 3. Commit and push
git add .
git commit -m "feat: add new resource support"
git push origin main

# 4. Tag the new version (follow semantic versioning)
git tag -a v0.2.0 -m "v0.2.0 - Added new features"
git push origin v0.2.0
```

> **Important**: Never modify or replace an already-released version. Always release a new version instead.

### Semantic Versioning

Follow [semver](https://semver.org/) for version numbers:

| Change Type | Version Bump | Example |
|---|---|---|
| Bug fix, no API change | Patch | `v0.1.0` -> `v0.1.1` |
| New feature, backwards compatible | Minor | `v0.1.0` -> `v0.2.0` |
| Breaking change | Major | `v0.1.0` -> `v1.0.0` |
| Prerelease | Hyphen suffix | `v0.2.0-beta1` |

---

## Alternative Release Methods

### Using GoReleaser Locally

If you prefer to build releases locally instead of via GitHub Actions:

1. Install [GoReleaser](https://goreleaser.com/install/):
   ```bash
   brew install goreleaser   # macOS
   # or: go install github.com/goreleaser/goreleaser/v2@latest
   ```

2. Copy the [`.goreleaser.yml`](https://github.com/hashicorp/terraform-provider-scaffolding-framework/blob/main/.goreleaser.yml) from the scaffolding repo (or use the one in this project).

3. Set your `GITHUB_TOKEN` to a [Personal Access Token](https://github.com/settings/tokens/new?scopes=public_repo) with `public_repo` scope.

4. Cache your GPG passphrase:
   ```bash
   gpg --armor --detach-sign  # Enter passphrase to cache it
   ```
   > **Note**: GoReleaser does not support GPG keys that require a passphrase interactively. Some systems cache the passphrase for a few minutes. If caching doesn't work, either use the `--batch` flag approach in the config or sign manually after building.

5. Tag and release:
   ```bash
   git tag -a v0.2.0 -m "v0.2.0"
   git push origin v0.2.0
   goreleaser release --clean
   ```

6. Dry run (creates artifacts locally without publishing):
   ```bash
   goreleaser release --snapshot --clean
   ls -lh dist/
   ```

### Manually Preparing a Release

If you need to create release assets without GoReleaser, ensure the following:

1. **Zip archives**: One per platform, named `terraform-provider-directus_{VERSION}_{OS}_{ARCH}.zip`, each containing the binary named `terraform-provider-directus_v{VERSION}`.

2. **SHA256 checksums file**:
   ```bash
   shasum -a 256 *.zip terraform-provider-directus_{VERSION}_manifest.json \
     > terraform-provider-directus_{VERSION}_SHA256SUMS
   ```

3. **GPG signature** (binary, not ASCII-armored):
   ```bash
   gpg --detach-sign terraform-provider-directus_{VERSION}_SHA256SUMS
   ```
   This creates `terraform-provider-directus_{VERSION}_SHA256SUMS.sig`.

4. **Manifest file**: Include `terraform-registry-manifest.json` renamed to `terraform-provider-directus_{VERSION}_manifest.json`.

5. Upload all assets to a GitHub Release tagged with the version.

---

## Troubleshooting

### CI pipeline fails on tag push

Check the **Actions** tab for detailed logs. Common issues:

- **Go version mismatch**: The CI uses Go 1.23. Ensure `go.mod` is compatible.
- **Test failures**: Fix failing tests and create a new tag. **Do not** re-tag the same version.
- **Permission errors**: Ensure the `GITHUB_TOKEN` has `contents: write` permission (already configured in the workflow).

### GoReleaser fails with GPG error

If you haven't set up GPG secrets:

1. Set up GPG secrets (see Step 7c), or
2. The workflow is designed to skip GPG signing when secrets are not configured -- the release will be created unsigned (sufficient for GitHub Releases, but not for Terraform Registry).

### Release artifacts missing

- Ensure the tag follows the `v*` format (e.g., `v0.1.0`, not `0.1.0`).
- Ensure there is **no branch** with the same name as the tag.
- Verify all CI jobs passed before the release job runs.

### Users can't find the provider in Terraform Registry

- Repository must be **public**.
- Repository name must be `terraform-provider-directus` (**lowercase**).
- GPG public key must be uploaded to the Registry.
- At least one GPG-signed release must exist.
- `terraform-registry-manifest.json` must be included in the release assets.
- The release must be **finalized** (not a private draft).

### Webhook not syncing new releases

1. Go to your GitHub repo's **Settings > Webhooks**.
2. Remove any existing webhooks for `registry.terraform.io`.
3. Go to the provider's settings page on the Terraform Registry.
4. Click the **Resync** button -- a new webhook will be created.

### Checksum errors for users

This usually means a released version was modified after publication. **Never** modify or replace an already-released version. Create a new version instead.

---

## Architecture of the CI/CD Pipeline

The pipeline uses **three separate workflows** following the
[official scaffolding pattern](https://github.com/hashicorp/terraform-provider-scaffolding-framework):

### Workflow 1: CI (`ci.yml`) -- Push / PR to main

```
  +---------+
  |  Lint   |──┐
  +---------+  │
               ├──> +---------+
  +---------+  │    |  Build  |
  |  Test   |──┘    +---------+
  +---------+  │
               └──> +------------------+
                    | Acceptance Tests |  (Directus service)
                    +------------------+
```

### Workflow 2: Release (`release.yml`) -- Push v* tag

```
  +-------------------------------------------+
  |  Test Gate                                |
  |  Unit Tests ──> Acceptance Tests          |
  |              (Directus service)           |
  +-------------------------------------------+
                      |
                      v
  +-------------------------------------------+
  |  GoReleaser                               |
  |  - Import GPG key                         |
  |  - Cross-compile binaries                 |
  |    (linux/darwin/windows/freebsd)         |
  |  - GPG sign checksums                     |
  |  - Include registry manifest              |
  |  - Create GitHub Release                  |
  +-------------------------------------------+
                      |
               Published on GitHub
          (Registry webhook picks it up)
```

### Workflow 3: E2E Tests (`e2e.yml`) -- Manual trigger

Two jobs run **in parallel**, each with its own Directus instance:

```
  +----------------------+    +----------------------+
  |  E2E Basic           |    |  E2E Comprehensive   |
  |  - test-e2e.sh       |    |  - test-e2e-         |
  |  - Core resource     |    |    comprehensive.sh  |
  |    lifecycle tests   |    |  - All resources +   |
  |                      |    |    import + state    |
  +----------------------+    +----------------------+
```

Trigger via **Actions > E2E Tests > Run workflow**.
Supports a configurable Directus version input.

---

## Key References

| Resource | URL |
|---|---|
| Publishing Providers (official) | https://developer.hashicorp.com/terraform/registry/providers/publishing |
| Provider Documentation Format | https://developer.hashicorp.com/terraform/registry/providers/docs |
| Recommended OS/Arch | https://developer.hashicorp.com/terraform/registry/providers/os-arch |
| Scaffolding Framework Repo | https://github.com/hashicorp/terraform-provider-scaffolding-framework |
| GoReleaser Config Reference | https://github.com/hashicorp/terraform-provider-scaffolding-framework/blob/main/.goreleaser.yml |
| Release Workflow Reference | https://github.com/hashicorp/terraform-provider-scaffolding-framework/blob/main/.github/workflows/release.yml |
| tfplugindocs | https://github.com/hashicorp/terraform-plugin-docs |
| Doc Preview Tool | https://registry.terraform.io/tools/doc-preview |
| Terraform Registry Terms | https://registry.terraform.io/terms |
| Registry Support Email | terraform-registry@hashicorp.com |

---

## Summary

| What | Command/URL |
|---|---|
| Run tests | `go test ./...` |
| Run acceptance tests | `TF_ACC=1 go test ./internal/provider/ -run TestAcc` |
| Run E2E tests | `./scripts/test-e2e-comprehensive.sh` |
| Create a release | `git tag -a v0.1.0 -m "msg" && git push origin v0.1.0` |
| Check CI status | `https://github.com/kylindc/terraform-provider-directus/actions` |
| View releases | `https://github.com/kylindc/terraform-provider-directus/releases` |
| Terraform Registry | `https://registry.terraform.io/providers/kylindc/directus` |
