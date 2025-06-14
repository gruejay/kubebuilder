# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

kubeguide is a Kubernetes management TUI (terminal user interface) with AI assistance that provides three core modes:

- **Explorer Mode**: Browse cluster resources with AI explanations
- **Editor Mode**: Create/edit Kubernetes manifests with AI assistance
- **Apply Mode**: Preview changes, validate, and apply to cluster

## Key Features

The tool focuses on AI-assisted Kubernetes manifest editing with:
- Real-time syntax checking and cluster API validation
- Smart suggestions for resource limits, labels, and best practices
- Template library for common patterns (web apps, databases, jobs)
- Dry-run capabilities for safe change preview

## Architecture

kubeguide is a TUI (terminal user interface) written in Go using tview. The inspiration is k9s, a popular
TUI for kubernetes/kubectl. 

The application operates in three distinct modes that handle different aspects of Kubernetes management:
1. Resource browsing and exploration with AI explanations
2. Manifest creation and editing with live validation
3. Change preview and cluster application

## Development Notes

This is a Kubernetes-focused tool that integrates AI assistance throughout the workflow. When working on this codebase, consider the three-mode architecture and the emphasis on safe, validated operations before applying changes to clusters.
