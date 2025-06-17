# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

kubeguide is a Kubernetes management TUI (terminal user interface) with AI assistance that provides three core modes:

- **Explorer Mode**: Browse cluster resources with AI explanations (currently implemented)
- **Editor Mode**: Create/edit Kubernetes manifests with AI assistance (planned)
- **Apply Mode**: Preview changes, validate, and apply to cluster (planned)

## Development Commands

```bash
# Build the application
make build          # Creates bin/kubeguide

# Run directly during development
make run           # Equivalent to: go run ./cmd/main.go

# Run the built binary
./bin/kubeguide
```

## Architecture

kubeguide is a TUI application written in Go using the tview framework, inspired by k9s. The application uses Go 1.24.3 and has a modular architecture centered around three core packages:

### Key Packages

- **`internal/app`**: Main application state management and mode orchestration
- **`internal/kubernetes`**: Unified Kubernetes client that handles both core resources and Custom Resource Definitions (CRDs) with resource discovery caching
- **`internal/ui`**: Modular UI components including fuzzy selectors, resource browsers, and detail viewers

### Architecture Patterns

1. **Unified Kubernetes Client** (`internal/kubernetes/client.go`): Abstracts both typed and dynamic client operations, supports namespaced and cluster-scoped resources
2. **Component-Based UI**: Reusable UI components with consistent tview styling (VSCode dark theme)
3. **Event-Driven Navigation**: Global key bindings with mode-specific input handling

## Current Implementation Status

**✅ Implemented**: Explorer mode with resource browsing, YAML details, namespace switching, vi-style navigation (j/k keys)

**🚧 In Development**: Key binding system structure exists but incomplete

**❌ Planned**: Editor mode, Apply mode, AI assistance features, testing framework

## Development Setup

- **Go Version**: 1.24.3+ required
- **Kubernetes Access**: Requires access to a Kubernetes cluster (uses standard kubeconfig resolution)
- **Dependencies**: Uses client-go v0.33.1, tview for UI, fuzzy search library

The codebase emphasizes safe, validated operations before applying changes to clusters and follows a three-mode architecture that can be extended as new modes are implemented.
