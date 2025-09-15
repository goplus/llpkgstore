# llpkgstore - Unified Package Distribution Service

[![Go Report Card](https://goreportcard.com/badge/github.com/goplus/llpkgstore)](https://goreportcard.com/report/github.com/goplus/llpkgstore)
[![Go Version](https://img.shields.io/github/go-mod/go-version/goplus/llpkgstore)](https://github.com/goplus/llpkgstore/blob/main/go.mod)
[![License](https://img.shields.io/github/license/goplus/llpkgstore)](https://github.com/goplus/llpkgstore/blob/main/LICENSE)
[![Build Status](https://github.com/goplus/llpkgstore/workflows/CI/badge.svg)](https://github.com/goplus/llpkgstore/actions)

> **llpkgstore** is a comprehensive package distribution service designed for [**LLGo**](https://github.com/goplus/llgo), providing trustworthy and convenient language binding access for multi-language ecosystems.

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage Guide](#usage-guide)
- [Configuration Reference](#configuration-reference)
- [Development Guide](#development-guide)
- [Technical Implementation Details](#technical-implementation-details)
- [FAQ](#faq)
- [Changelog](#changelog)
- [Related Resources](#related-resources)

## Overview

llpkgstore is a unified package distribution service designed for [**LLGo**](https://github.com/goplus/llgo), providing trustworthy and convenient language binding access for multi-language ecosystems.

### What is llpkgstore?

llpkgstore is a comprehensive package management tool that can:

- **Automatically generate language bindings**: Automatically generate Go language bindings for C/C++ and Python libraries
- **Unified version management**: Provide consistent version mapping and management mechanisms
- **Smart package detection**: Prioritize packages in the system environment for improved efficiency
- **CI/CD integration**: Seamless integration with GitHub Actions for automated package generation and publishing

### Core Values

- **🚀 Simplified Integration**: Simplify complex cross-language binding generation into a few commands
- **🔒 Secure and Reliable**: Ensure package quality through automated processes and validation mechanisms
- **🌍 Multi-language Support**: Unified support for C/C++ and Python ecosystems
- **⚡ Efficient and Convenient**: Smart detection and caching mechanisms to reduce redundant work

## Features

### Multi-language Support

llpkgstore currently supports the following programming languages and platforms:

| Language/Platform | Status | Generator Tool | Package Manager | Key Features |
|-------------------|--------|----------------|-----------------|--------------|
| **C/C++** | ✅ Fully Supported | `llcppg` | Conan | Binary distribution, headers, .pc files |
| **Python** | ✅ Fully Supported | `llpyg` | pip | Module bindings, Go interfaces, type safety |
| **JavaScript** | 🚧 Planned | - | - | Planned support |
| **Rust** | 🚧 Planned | - | - | Planned support |

### Core Features

#### 🔄 Unified Version Management
- **Smart version extraction**: Automatically extract version information from commit messages
- **Version format validation**: Support for semantic versioning and custom formats
- **Dual recording mechanism**: Synchronize local and centralized version records
- **Automatic Git tags**: Automatically create and push tags based on version information

#### 🎯 Smart Package Detection
- **System package priority**: Prioritize packages already installed in the system environment
- **Temporary installation**: Perform temporary package installation only when necessary
- **Caching mechanism**: Smart caching to reduce redundant downloads and installations

#### 🚀 Automated Workflows
- **CI/CD integration**: Seamless integration with GitHub Actions
- **Automatic publishing**: Automatically create GitHub Releases
- **Test validation**: Automatically generate and run test code
- **Error handling**: Comprehensive error handling and rollback mechanisms

#### 🔧 Developer Experience
- **Unified CLI**: Consistent command-line interface
- **Configuration-driven**: Flexible settings based on configuration files
- **Detailed logging**: Complete operation logs and debugging information
- **Documentation generation**: Automatically generate usage documentation and examples

## 🏗️ System Architecture

### Overall Architecture Diagram

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   User Interface│    │  Business Logic │    │  Data Storage   │
│                 │    │                 │    │                 │
│ • CLI Commands  │◄──►│ • Package Logic │◄──►│ • Config Files  │
│ • Config Interface│    │ • Version Mgmt  │    │ • Version Records│
│ • Error Handling│    │ • Generator Calls│    │ • Metadata Store│
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Core Components

#### 1. Command Layer
- Unified CLI interface
- Automatic package type detection
- Smart routing and distribution

#### 2. Business Logic Layer
- Core package management logic
- Version management and mapping
- Generator integration

#### 3. Configuration Management
- Unified configuration file format (`llpkg.cfg`)
- Support for package type-specific configuration options
- Automatic configuration validation and error prompts

#### 4. Version Management
- **Unified version mapping**: C/C++ and Python packages use consistent version management mechanisms
- **Smart version extraction**: Automatically extract version information from commit messages
- **Automatic Git tags**: Automatically create and push Git tags based on version information
- **Version validation**: Complete version format validation and conflict detection

## Installation

### System Requirements

Before installing llpkgstore, ensure your system meets the following requirements:

| Component | Minimum Version | Description |
|-----------|-----------------|-------------|
| **Go** | 1.19+ | For building and running llpkgstore |
| **Python** | 3.7+ | For Python package support |
| **Git** | 2.0+ | For version control and tag management |
| **Operating System** | - | Linux, macOS, Windows |

### Installation Methods

#### Method 1: Using go install (Recommended)

```bash
go install github.com/goplus/llpkgstore/cmd/llpkgstore@latest
```

#### Method 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/goplus/llpkgstore.git
cd llpkgstore

# Build
go build -o llpkgstore ./cmd/llpkgstore

# Install to system path
sudo cp llpkgstore /usr/local/bin/
```

#### Method 3: Download Pre-compiled Binaries

Download pre-compiled versions suitable for your system from [GitHub Releases](https://github.com/goplus/llpkgstore/releases).

### Verify Installation

After installation, verify that llpkgstore is correctly installed:

```bash
llpkgstore --version
```

If installation is successful, you should see output similar to:

```
llpkgstore version 2.0.0
```

### Environment Configuration

#### Python Environment Configuration

llpkgstore requires a Python environment to support Python package processing:

```bash
# Check Python version
python3 --version

# Check pip version
pip3 --version

# Optional: Create a dedicated virtual environment
python3 -m venv ~/.llpkgstore-env
source ~/.llpkgstore-env/bin/activate  # Linux/macOS
# or
~/.llpkgstore-env\Scripts\activate     # Windows
```

#### C/C++ Environment Configuration (Optional)

If you need to handle C/C++ packages, you can install Conan:

```bash
# Install Conan
pip3 install conan

# Configure Conan
conan profile detect --force
```

### Troubleshooting

#### Common Installation Issues

**Issue 1**: `command not found: llpkgstore`
```bash
# Solution: Ensure Go bin directory is in PATH
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc
```

**Issue 2**: Python version incompatibility
```bash
# Solution: Use pyenv to manage Python versions
curl https://pyenv.run | bash
pyenv install 3.9.0
pyenv global 3.9.0
```

## Quick Start

This guide will help you quickly get started with llpkgstore by creating Go bindings for a Python package in a few simple steps.

### Step 1: Create Project Directory

```bash
mkdir my-python-package
cd my-python-package
```

### Step 2: Create Configuration File

Create a `llpkg.cfg` configuration file specifying the Python package to process:

```json
{
  "type": "python",
  "upstream": {
    "installer": {
      "name": "pip"
    },
    "package": {
      "name": "requests",
      "version": "2.31.0"
    }
  },
  "llpyg": {
    "output_dir": "./bindings",
    "mod_name": "github.com/your-org/requests",
    "mod_depth": 1
  }
}
```

### Step 3: Install Package

```bash
llpkgstore install llpkg.cfg
```

### Step 4: Generate Go Bindings

```bash
llpkgstore generate
```

### Step 5: Verify Results

Check the generated files:

```bash
ls -la bindings/
```

You should see generated Go files including:
- `requests.go` - Main Go binding file
- `requests_autogen_link.go` - Auto-generated link file

### Step 6: Test Bindings

```bash
cd bindings
go mod tidy
go run _demo/main.go
```

## Usage Guide

### Basic Commands

llpkgstore provides the following main commands:

| Command | Description | Example |
|---------|-------------|---------|
| `install` | Install specified package | `llpkgstore install llpkg.cfg` |
| `generate` | Generate Go bindings | `llpkgstore generate` |
| `postprocessing` | Post-processing (version management) | `llpkgstore postprocessing` |
| `release` | Create release | `llpkgstore release` |

### Getting llpkg

#### C/C++ Packages
Use `llgo get` to get C/C++ llpkg:

```bash
llgo get clib@cversion
```

*Example* `llgo get cjson@1.7.18`

- `clib`: Original C library name
- `cversion`: Original C library version

#### Python Packages
Use `llgo get` to get Python llpkg:

```bash
llgo get github.com/goplus/llpkg/numpy@v1.26.4
```

Or use simplified syntax (if supported):

```bash
llgo get numpy@1.26.4
```

#### Universal Syntax
Both package types support universal syntax:

```bash
llgo get module_path@module_version
```

*Example* `llgo get github.com/goplus/llpkg/cjson@v1.0.0`

```bash
llgo get clib[@latest]
llgo get module_path[@latest]
```

The optional `latest` identifier is supported as a valid `cversion` or `module_version`. When using `llgo get clib@latest`, `llgo get` will first convert `clib` to `module_path`, then process it as `module_path@latest`.

### Listing Package Version Mappings

```
llgo list -m [-versions] [-json] [modules/clibs]
```

- `llgo list -m` is compatible with `go list -m`
- `modules`: A set of space-separated module_path[@module_version]
- `clibs`: A set of space-separated clib[@cversion]

Each argument is processed separately.

#### Module Query Examples

**C/C++ packages**:
```bash
llgo list -m cjson
# Output: github.com/goplus/llpkg/cjson v0.1.0[conan:cjson/1.7.18]
```

**Python packages**:
```bash
llgo list -m numpy
# Output: github.com/goplus/llpkg/numpy v1.26.4[pip:numpy/1.26.4]
```

**View all versions**:
```bash
llgo list -m -versions cjson
# Output: github.com/goplus/llpkg/cjson v0.1.0[conan:cjson/1.7.18] v0.1.1[conan:cjson/1.7.18] v0.2.0[conan:cjson/1.7.19]
```

### Configuration File Format

Detailed configuration file format description can be found in the [Configuration Reference](#configuration-reference) section.

### Package Generation Workflow

#### C/C++ Package Generation

Standard method for generating valid C/C++ llpkgs:

1. **Receive binaries/headers**: Receive binary files and headers from installer, and index them into `.pc` files
2. **Detect generator**: Detect generator from configuration files. For example, if an `llcppg.cfg` file is present in the current directory, we can directly use `llcppg`
3. **Automatically generate llpkg**: Use generator to automatically generate llpkg for different platforms
4. **Combine generated results**: Combine generated results into one Go module
5. **Debug and re-generate**: Debug and re-generate llpkg by modifying configuration files

#### Python Package Generation

Standard method for generating valid Python llpkgs:

1. **Install Python package**: Use pip installer to install Python package
2. **Generate Go bindings**: Use `llpyg` tool to generate Go bindings for Python modules
3. **Configure module**: Configure output directory, module name, and extraction depth
4. **Generate Go interfaces**: Generate Go interfaces and type-safe bindings
5. **Create Go module**: Create proper Go module with dependencies
6. **Test generated bindings**: Test generated bindings with demo code

#### Generation Commands

**C/C++ packages**:
```bash
llpkgstore generate
```

**Python packages**:
```bash
llpkgstore generate
```

This automatically detects package type from `llpkg.cfg` and uses the appropriate generator.

### PR Workflow

#### Standard PR Workflow

1. **Create PR**: Trigger GitHub Action
2. **PR verification**: Verify PR content
3. **llpkg generation**: Generate llpkg
4. **Run tests**: Execute tests
5. **Review generated llpkg**: Check generation results
6. **Merge PR**: Merge to main branch
7. **Post-processing**: Run post-processing GitHub Action on main branch

#### PR Verification Workflow

1. **Ensure uniqueness**: Ensure there is only one `llpkg.cfg` file across all directories. If multiple `llpkg.cfg` instances are detected, the PR will be aborted
2. **Directory name validation**: Check if directory name is valid, directory name in PR **SHOULD** equal `Package.Name` field in `llpkg.cfg` file
3. **Commit message validation**: Check if PR commit footer contains [`{MappedVersion}`](#mappedversion-in-pr-commit)

#### Merge PR

Maintainers **SHOULD** squash commits before merging a PR. The squash commit message **MUST** include [`{MappedVersion}`](#mappedversion-in-pr-commit) to enable the Post-processing GitHub Action to parse it correctly.

**`{MappedVersion}` in PR Commit**:
`{MappedVersion}` **MUST** be included in at least one of the commits in the PR and **MUST** follow this format:

```
Release-as: {PackageName}/{MappedVersion}
```

The PR verification process will validate this format and abort the PR if it is invalid.

**Example for C/C++ package**:
```bash
git merge
# Modify the merge commit message
git commit --amend -m "feat: add cjson" -m "Release-as: cjson/v1.0.0"
```

**Example for Python package**:
```bash
git merge
# Modify the merge commit message
git commit --amend -m "feat: add numpy" -m "Release-as: numpy/v1.26.4"
```

## Configuration Reference

### Basic Structure

```json
{
  "type": "package_type",
  "upstream": {
    "installer": {
      "name": "installer_name"
    },
    "package": {
      "name": "package_name",
      "version": "version_number"
    }
  }
}
```

### Package Type Specific Configuration

#### C/C++ Package Configuration

```json
{
  "type": "cpp",
  "upstream": {
    "installer": {
      "name": "conan"
    },
    "package": {
      "name": "cjson",
      "version": "1.7.18"
    }
  }
}
```

#### Python Package Configuration

```json
{
  "type": "python",
  "upstream": {
    "installer": {
      "name": "pip"
    },
    "package": {
      "name": "numpy",
      "version": "1.26.4"
    }
  },
  "llpyg": {
    "output_dir": "./test",
    "mod_name": "github.com/PengPengPeng717/llpkg/numpy",
    "mod_depth": 1
  }
}
```

### Field Description

**Common Fields**

| Field | Type | Default Value | Optional | Description |
|-------|------|---------------|----------|-------------|
| type | `string` | "cpp" | ✅ | Package type: "cpp" or "python" |
| upstream.installer.name | `string` | "conan" | ✅ | Upstream binary provider |
| upstream.installer.config | `map[string]string` | {} | ✅ | Installer configuration |
| upstream.package.name | `string` | - | ❌ | Package name in platform |
| upstream.package.version | `string` | - | ❌ | Original package version |

**Python-specific Fields (llpyg section)**

| Field | Type | Default Value | Optional | Description |
|-------|------|---------------|----------|-------------|
| llpyg.output_dir | `string` | "./test" | ✅ | Output directory for generated files |
| llpyg.mod_name | `string` | package name | ✅ | Go module name |
| llpyg.mod_depth | `int` | 1 | ✅ | Maximum module extraction depth (0-10) |

## Development Guide

### Development Environment Setup

#### 1. Clone Repository

```bash
git clone https://github.com/goplus/llpkgstore.git
cd llpkgstore
```

#### 2. Install Dependencies

```bash
# Install Go dependencies
go mod tidy

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/goplus/llgo@latest
```

#### 3. Build Project

```bash
# Build llpkgstore
go build -o llpkgstore ./cmd/llpkgstore

# Run tests
go test ./...

# Run lint checks
golangci-lint run
```

### Code Structure

#### Core Modules

- **`cmd/llpkgstore/`**: Command-line interface implementation
- **`internal/actions/`**: Core business logic
- **`config/`**: Configuration management and validation
- **`upstream/`**: Upstream package manager integration
- **`metadata/`**: Metadata management and caching

#### Development Standards

1. **Code style**: Follow Go official code standards
2. **Test coverage**: New features must include unit tests
3. **Documentation updates**: Synchronize documentation updates when modifying features
4. **Error handling**: Provide clear error information and handling mechanisms

### Contribution Process

#### 1. Create Issue

Before submitting code, please first create an Issue describing the problem you want to solve or the feature you want to add.

#### 2. Fork and Branch

```bash
# Clone after forking repository
git clone https://github.com/your-username/llpkgstore.git
cd llpkgstore

# Create feature branch
git checkout -b feature/your-feature-name
```

#### 3. Development and Testing

```bash
# Develop features
# ... write code ...

# Run tests
go test ./...

# Run lint
golangci-lint run

# Build verification
go build ./cmd/llpkgstore
```

#### 4. Submit Code

```bash
# Add changes
git add .

# Commit changes
git commit -m "feat: add your feature description"

# Push branch
git push origin feature/your-feature-name
```

#### 5. Create Pull Request

Create a Pull Request on GitHub, including:

- Clear title and description
- Related Issue links
- Test result screenshots (if applicable)
- Documentation update description

### Release Process

#### Version Management

llpkgstore uses semantic versioning:

- **Major version**: Incompatible API changes
- **Minor version**: Backward-compatible feature additions
- **Patch version**: Backward-compatible bug fixes

#### Release Steps

1. **Update version number**: Modify version in `go.mod`
2. **Update documentation**: Update CHANGELOG.md
3. **Create tag**: `git tag v1.0.0`
4. **Push tag**: `git push origin v1.0.0`
5. **Create Release**: Create Release on GitHub

### Debugging Guide

#### Enable Debug Mode

```bash
# Set debug environment variable
export LLPKGSTORE_DEBUG=1

# Run command to view detailed logs
llpkgstore generate --verbose
```

#### Common Debug Scenarios

1. **Package installation failure**: Check network connection and package manager configuration
2. **Generator errors**: Verify configuration file format and dependencies
3. **Version conflicts**: Check version mapping and Git tags

## Technical Implementation Details

### Smart Package Detection Mechanism

llpkgstore includes smart package detection that prioritizes system-installed packages:

```go
// isPackageInstalledInSystem checks if the specified package is already installed in the system environment
func isPackageInstalledInSystem(packageName string) bool {
    // Method 1: Try to import the package directly
    if canImportPackage(packageName) {
        return true
    }
    
    // Method 2: Check pip list output
    if isPackageInPipList(packageName) {
        return true
    }
    
    return false
}
```

### Architecture Improvements

#### Unified Version Management

The latest version of llpkgstore includes unified version management for all package types:

1. **Common version extraction**: Unified logic for extracting versions from commit messages
2. **Standardized version validation**: Consistent semver validation across package types
3. **Unified version mapping**: Common version mapping strategies
4. **Consistent branch management**: Standardized branch naming and lifecycle

#### Enhanced Error Handling

- **Structured error types**: Consistent error handling across all operations
- **Detailed logging**: Comprehensive logging for debugging and monitoring
- **Error recovery**: Mechanisms for handling and recovering from errors

#### Improved CI/CD Pipeline

- **Package type detection**: Automatic detection of package types
- **Conditional processing**: Different processing logic for different package types
- **Unified workflow**: Consistent workflow across all package types
- **Enhanced testing**: Comprehensive testing for all package types

### Environment Variable Design

One usage is to store `.pc` files of the C library and allow `llgo build` to find them.

1. `LLGOCACHE` defaults to `{UserCacheDir}/llgo/`
2. `.pc` files of C libs needed by llpkg will be stored in `{LLGOCACHE}/pkg-config/{module_path}@{module_version}/`
3. If `UserCacheDir` isn't available, `llgo` will exit with an error

## FAQ

### Installation Issues

**Q: Cannot find llpkgstore command after installation**
A: Ensure `/usr/local/bin` is in your PATH environment variable, or use `go install` to install to GOPATH.

**Q: Python package installation failed**
A: Ensure Python 3.8+ and pip are installed, and check network connection.

### Configuration Issues

**Q: Configuration file format error**
A: Use JSON validator to check configuration file format, ensure all required fields exist.

**Q: Package type detection failed**
A: Ensure `type` field is set to `"python"` or empty.

### Generation Issues

**Q: Go binding generation failed**
A: Check if upstream package is correctly installed, ensure llpyg or llcppg tools are available.

## Troubleshooting

### Debug Mode

```bash
# Enable verbose output
llpkgstore --verbose generate

# Check configuration
llpkgstore config validate
```

### Log Files

- **Installation logs**: `~/.llpkgstore/logs/install.log`
- **Generation logs**: `~/.llpkgstore/logs/generate.log`
- **Error logs**: `~/.llpkgstore/logs/error.log`

### Common Errors

#### 1. Package Not Found
```
Error: package not found
```
**Solution**: Check if package name and version number are correct.

#### 2. Configuration Error
```
Error: invalid configuration
```
**Solution**: Verify configuration file format and required fields.

#### 3. Permission Issues
```
Error: permission denied
```
**Solution**: Check file permissions and directory access permissions.

## Changelog

### v2.0.0 (Latest)

#### 🎉 Major Updates
- **Unified version management**: C/C++ and Python packages use consistent version management mechanisms
- **Smart version extraction**: Automatically extract version information from commit messages
- **Automatic Git tags**: Automatically create and push Git tags based on version information

#### ✨ New Features
- **Python package support**: Complete Python package Go binding generation
- **Unified command interface**: Automatic package type detection and smart routing
- **Enhanced error handling**: Structured error handling and recovery mechanisms
- **CI/CD integration**: Optimized GitHub Actions workflows

#### 🔧 Improvements
- **Architecture unification**: Eliminated architectural inconsistencies between C++ and Python parts
- **Code reuse**: Reduced duplicate code, improved maintainability
- **User experience**: Consistent command interface and error message format

### v1.0.0

#### 🎉 Initial Version
- **C/C++ package support**: Complete support for C/C++ library Go binding generation
- **Conan integration**: Support for Conan package manager
- **Version management**: Complex semantic version mapping
- **CI/CD integration**: GitHub Actions workflow support

#### ✨ Core Features
- **Package management**: Automatic package installation and dependency management
- **Binding generation**: Generate Go bindings using llcppg
- **Test framework**: Automatically generate demo code and tests
- **Release management**: GitHub Releases integration

---

## Related Resources

- **[Architecture Documentation](./ARCHITECTURE.md)**: Detailed system architecture description
- **[GitHub Repository](https://github.com/goplus/llpkgstore)**: Source code and issue tracking
- **[LLGo Project](https://github.com/goplus/llgo)**: LLGo language extension

---

**llpkgstore** - Making cross-language development simpler and more reliable! 🚀
