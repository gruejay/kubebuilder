# kubeguidn

Enhanced "K8s Mentor" with Manifest Editor

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

## Implemented

- Explorer mode for Pods and Services
