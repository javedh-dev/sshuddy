# Deployment Guide

This guide explains how to set up CI/CD for sshbuddy with GitHub Actions and Homebrew publishing.

## Prerequisites

1. **GitHub Repository**: Your code must be in a GitHub repository
2. **GitHub Account**: You need access to create releases and secrets
3. **Homebrew Tap Repository**: A separate repo for your Homebrew formula

## Setup Steps

### 1. Create Homebrew Tap Repository

Create a new GitHub repository named `homebrew-tap` under your account/organization:

```bash
# Example: if your username is "johndoe", create:
# https://github.com/johndoe/homebrew-tap
```

This repository will store your Homebrew formula.

### 2. Generate GitHub Personal Access Token

1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Give it a name like "Homebrew Tap Token"
4. Select scopes:
   - `repo` (all)
   - `workflow`
5. Generate and copy the token

### 3. Add Secret to Your Repository

1. Go to your sshbuddy repository
2. Settings → Secrets and variables → Actions
3. Click "New repository secret"
4. Name: `HOMEBREW_TAP_GITHUB_TOKEN`
5. Value: Paste the token from step 2
6. Click "Add secret"

### 4. Push Your Code

```bash
git add .
git commit -m "Add CI/CD configuration"
git push origin main
```

### 5. Create Your First Release

```bash
# Tag your release
git tag -a v0.1.0 -m "Initial release"
git push origin v0.1.0
```

This will trigger the release workflow which will:
- Build binaries for Linux, macOS, and Windows (amd64 and arm64)
- Create a GitHub release with all artifacts
- Automatically publish to your Homebrew tap

## Installing via Homebrew

Once published, users can install with:

```bash
# Add your tap (one time only)
brew tap YOUR_USERNAME/tap

# Install sshbuddy
brew install sshbuddy
```

Or in one command:

```bash
brew install YOUR_USERNAME/tap/sshbuddy
```

## Updating the Formula

When you create a new release (new tag), GoReleaser automatically:
1. Builds new binaries
2. Updates the Homebrew formula in your tap repository
3. Users can update with `brew upgrade sshbuddy`

## Manual Release Process

If you prefer manual releases:

```bash
# Install GoReleaser
brew install goreleaser

# Create a release locally (dry run)
goreleaser release --snapshot --clean

# Create actual release (requires tag)
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0
```

## Workflow Overview

### Test Workflow (`.github/workflows/test.yml`)
- Runs on every push and PR
- Tests on Ubuntu and macOS
- Runs build, test, and vet

### Release Workflow (`.github/workflows/release.yml`)
- Triggers on version tags (v*)
- Uses GoReleaser to build and publish
- Creates GitHub release
- Updates Homebrew formula

## Configuration Files

- `.goreleaser.yml`: GoReleaser configuration
- `.github/workflows/test.yml`: CI testing
- `.github/workflows/release.yml`: Release automation

## Troubleshooting

### Release fails with "token not found"
- Verify `HOMEBREW_TAP_GITHUB_TOKEN` secret is set correctly
- Check token has correct permissions

### Homebrew formula not updating
- Ensure your tap repository exists and is named `homebrew-tap`
- Check the token has write access to the tap repository

### Build fails
- Verify Go version in workflows matches your go.mod
- Check all dependencies are properly vendored

## Version Numbering

Follow semantic versioning:
- `v0.1.0` - Initial release
- `v0.2.0` - Minor updates, new features
- `v1.0.0` - Stable release
- `v1.0.1` - Patch/bugfix

## Next Steps

1. Update README.md with installation instructions
2. Add badges for build status
3. Consider adding more platforms if needed
4. Set up automated testing for UI components
