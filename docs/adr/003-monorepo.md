# ADR-003: Monorepo over three separate repositories

**Status:** Accepted  
**Date:** 2026-04-14

## Context

The platform has three distinct concerns with different tech stacks and release cadences:
- `monitoring-app` — Go backend services
- `monitoring-console` — React TypeScript UI
- `infrastructure` — Terraform + Helm

We initially planned three separate repos, then merged into one.

## Decision

Single monorepo under `monitoring-platform/` with three top-level subdirectories.

## Reasoning

**Atomic cross-cutting changes.** Adding a new service requires touching all three areas: Go
code, a Helm chart, and a Terraform `helm_release`. In a monorepo this is one PR, one review,
one merge. Across three repos it requires coordinating three PRs, three CI runs, and three
merges in the right order.

**Simplified CI for `terraform.tfvars` bumps.** App CI updates image tags in
`infrastructure/terraform/envs/dev/20-app/terraform.tfvars` as part of the same repo. No
cross-repo tokens (`INFRA_REPO_TOKEN`), no cross-repo API calls, no external GitHub App.

**Path-scoped triggers prevent wasted CI.** GitHub Actions `paths:` filters mean only the
relevant workflow runs on any given PR:
- `monitoring-app/**` → `app-ci.yaml` only
- `monitoring-console/**` → `console-ci.yaml` only
- `infrastructure/**` → `terraform.yaml` only

This matches the key benefit of polyrepo (independent CI) without the coordination overhead.

**Single source of truth for architecture.** `docs/` sits at the repo root and can reference
any part of the system without cross-repo links that rot.

## Tradeoffs accepted

- **Permissions are coarser.** Everyone with repo access can see all three areas. Mitigated by
  CODEOWNERS rules (separate owners for `infrastructure/` vs `monitoring-app/`).
- **Larger clone.** `node_modules` and Go module cache are separate per subdirectory; `git clone`
  is still fast because neither is committed.
- **If team grows significantly** (>10 engineers, separate infra and product teams) a polyrepo
  split may become worthwhile. The code boundaries are clean enough that splitting is
  straightforward if that time comes.

## Consequences

- Each subdirectory maintains its own build toolchain: `go.mod` in `monitoring-app/`,
  `package.json` in `monitoring-console/`, `terraform` in `infrastructure/`.
- Root `Makefile` delegates to each subdirectory — `make app-build`, `make console-build`,
  `make infra-plan`.
- `.gitignore` at the root covers all three build artifact patterns.
