---
name: Full Coverage Test Plan
overview: Build a near-total unit test suite (95%+) across all packages by adding deterministic test fixtures, package-level harnesses, and small testability refactors where dependencies are currently hardwired. Prioritize behavior coverage over line-chasing and ensure every core branch/path is asserted.
todos:
  - id: execute-on-test-branch
    content: Execute all test and coverage work on the already-created test branch; do not create a new branch.
    status: completed
  - id: add-test-harness
    content: Add shared test helpers for sqlite fixtures, httptest utilities, and dependency stubs.
    status: completed
  - id: cover-domain-packages
    content: Implement table-driven unit tests for models, repositories, and services including all error branches.
    status: completed
  - id: cover-web-layer
    content: Add handler/middleware tests for routing, auth gating, moderation, forum interactions, and fallback branches.
    status: completed
  - id: cover-infra-packages
    content: Add complete tests for oauth clients, render transport, and validator helpers.
    status: completed
  - id: add-coverage-gate
    content: Add coverage profile command, threshold enforcement (95%), and README/Makefile test docs.
    status: completed
isProject: false
---

# Full Coverage Unit Test Plan

## Goal
Reach near-total unit test coverage (target `>=95%`) across all packages in this repository.

## Branch Requirement
- Perform all implementation on the already-created test branch.
- Do not create a new branch for this work.
- Keep commits scoped and reviewable on this same test branch.

## Scope
- Cover all logic in:
  - `/cmd/web` handlers and middleware
  - `/internal/models`
  - `/internal/repository/sqlite`
  - `/internal/service`
  - `/internal/oauth`
  - `/internal/transport/http`
  - `/internal/validator`

## Implementation Plan
1. Add shared test harnesses
   - DB fixture helpers (SQLite test DB + schema seed)
   - HTTP test helpers (`httptest` request/response/session fixtures)
   - Mock/stub helpers for outbound HTTP and dependency seams
2. Expand domain tests
   - Table-driven tests for models/repositories/services
   - Success, validation, and error-path assertions
3. Expand web-layer tests
   - Route parsing and method switching
   - Authentication/authorization gates
   - Moderation/report/approval branches
   - Forum create/edit/remove/comment/like flows
4. Expand infra/helper tests
   - OAuth client parsing and non-200/error paths
   - Render/middleware helper behavior
   - Validator full branch checks
5. Add coverage enforcement + docs
   - Update `Makefile` with explicit coverage targets:
     - `test`
     - `test-cover`
     - `test-cover-enforce` (fails when total coverage < 95%)
   - Update `README.md` test section with new commands and expected usage

## Key File Updates
- `Makefile` (mandatory): add/refresh test + coverage targets and threshold enforcement
- `README.md`: document coverage commands and branch workflow expectations
- New `*_test.go` files across packages above

## Verification
- Run `go test ./...`
- Run coverage profile commands from `Makefile`
- Ensure threshold gate passes at `>=95%`