# kubeguide

Enhanced "K8s Mentor" with Manifest Editor and AI Analysis

## Usage

Uses whatever context is selected in your kubeconfig. There is currently no support
for changing contexts inside kubeguide.

```bash
make run    # or go run cmd/main.go
make build  # or go build -o bin/kubeguide cmd/main.go
```

Once launched, use `?` for help.

Navigation is similar to Vim:
- `j/k` or `Tab/Shift-tab` to move forward/backward in lists
- In fuzzy search, use `Ctrl-j/k` or `Tab/Shift-tab` to change selection

## AI-Powered Pod Analysis

kubeguide includes AI assistance to help troubleshoot failed pods. When viewing a pod in Explorer mode, press `a` to get AI analysis of potential issues.

### Setup

1. **Configuration**: Copy the example config and customize:
   ```bash
   cp config.example.yaml ~/.config/kubeguide/config.yaml
   ```

2. **API Key**: Set your AI provider API key via environment variable:
   ```bash
   # For OpenAI
   export OPENAI_API_KEY="your-openai-key"
   
   # For Anthropic
   export ANTHROPIC_API_KEY="your-anthropic-key"
   
   # Or use the generic variable
   export KUBEGUIDE_AI_API_KEY="your-api-key"
   ```

### Supported AI Providers

- **OpenAI**: GPT-4o, GPT-4o-mini, GPT-3.5-turbo
- **Anthropic**: Claude-3-haiku, Claude-3-sonnet  
- **Ollama**: Any local model (llama2, codellama, etc.)
- **Any OpenAI-compatible API**

The provider is auto-detected based on your API key format or can be manually configured.

### Features

- **Smart Detection**: Focuses on failed/problematic pods
- **Root Cause Analysis**: Identifies resource constraints, image pull issues, config problems
- **Actionable Recommendations**: Provides specific fixes and best practices
- **Multiple Provider Support**: Works with OpenAI, Anthropic, Ollama, and compatible APIs

## Plan
Core Modes:

    Explorer Mode: Browse cluster resources + AI explanations
    Editor Mode: Create/edit manifests with AI assistance
    Apply Mode: Preview changes, validate, and apply to cluster

Manifest Editor Features:

    AI-Assisted Writing: Start with "I want to create a deployment for nginx" → AI generates base manifest
    Live Validation: Real-time syntax checking + cluster API validation
    Smart Suggestions: AI suggests resource limits, labels, best practices as you type
    Explain-as-you-go: Hover/select any field → AI explains what it does
    Template Library: AI can suggest common patterns (web app, database, job, etc.)

```
┌─ Mode: Editor ──────┐┌─ AI Assistant ──────┐
│ □ New Manifest      ││ 💡 I notice you're   │
│ □ Edit Selected     ││ creating a web app.  │
│ □ From Template     ││ Should I add a       │
└─────────────────────┘│ Service for it?      │
┌─ Manifest Editor ───┤└─────────────────────┘
│ apiVersion: apps/v1 │┌─ Validation ────────┐
│ kind: Deployment    ││ ✅ Syntax valid      │
│ metadata:           ││ ⚠️  No resource      │
│   name: my-app      ││    limits set        │
│ spec:               │└─────────────────────┘
│   replicas: 3       │
└─────────────────────┘

```
Workflows:

    "Edit this pod": Select running pod → AI helps you create a proper Deployment for it
    "Fix this issue": Failing resource → AI suggests manifest changes to fix it
    "Dry run first": Always preview what changes will do before applying

## Current Implementation Status

### ✅ Implemented
- **Explorer Mode**: Browse pods, services, deployments, configmaps, secrets
- **Resource Details**: View YAML details of any resource
- **Namespace Switching**: Fuzzy search namespace selector (`n` key)
- **Resource Filtering**: Filter by resource type (`r` key)
- **AI Pod Analysis**: Analyze failed pods with AI assistance (`a` key)
- **Multi-Provider AI**: Support for OpenAI, Anthropic, Ollama, and compatible APIs
- **Vi-style Navigation**: `j/k` keys, Esc to go back, `?` for help

### 🚧 Planned
- **Editor Mode**: Create/edit Kubernetes manifests with AI assistance
- **Apply Mode**: Preview changes, validate, and apply to cluster
- **Live Validation**: Real-time syntax checking + cluster API validation
- **Template Library**: AI-suggested common patterns
