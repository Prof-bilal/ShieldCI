# ShieldCI implementation phases

This is the execution plan for turning the ShieldCI concept into a working MVP. Each phase ends with a runnable artifact and a gate before the next phase begins.

## System boundary

ShieldCI is the control plane and CI/CD product. Warden is the execution engine. ShieldCI must call Warden through a narrow adapter and must preserve Warden’s fail closed and deny by default behavior.

## Phase 0 — foundation and source baseline

Status: complete for the source baseline; Phase 1 adapter slice complete

Deliverables:

- Root project structure and reproducible build commands.
- Warden source snapshot pinned to a known commit under `engine/warden/`.
- MVP scope, service boundaries, threat assumptions, and test gates.
- Local development commands for the backend, dashboard, and Warden engine.

Gate: passed for this snapshot. Warden focused policy and sandbox tests pass (84 tests across 7 packages), and the CLI builds with `-buildvcs=false` because the snapshot has no nested Git metadata.

## Phase 1 — execution adapter and policy contract

Deliverables:

- Go `wardenadapter` package wrapping the pinned Warden CLI or library boundary.
- Versioned ShieldCI policy schema with parsing, validation, normalization, and deny by default tests.
- One local command that executes a harmless fixture inside Warden and returns structured result plus audit events.

Current progress: the control-plane module under `backend/` validates the first JSON policy contract and exposes `GET /healthz` plus `POST /api/v1/policies/validate`. The `wardenadapter` now writes Warden-compatible YAML, resolves relative grants against the configured working directory, invokes the pinned CLI explicitly, and returns structured output plus an audit event. Unit tests use a harmless fake Warden executable; a real sandbox execution fixture is the next gate because this environment does not permit opening local sockets required by some Warden backends.

Gate: allowed filesystem access succeeds, undeclared filesystem access is denied, undeclared network access is denied, and missing sandbox support returns an error without running the command unsandboxed.

## Phase 2 — MVP control plane API

Status: in progress — API surface, SQLite persistence, and OpenAPI contract are implemented; worker execution is covered and production schema migration remains.

Deliverables:

- Go API service with health, policy, job, audit, and finding endpoints.
- SQLite for local development behind a repository interface; PostgreSQL remains the production target.
- Request validation, stable error responses, structured logging, and OpenAPI documentation.

Gate: API integration tests cover policy creation, job submission, job status, audit retrieval, and invalid input.

Current progress: the API now supports policy creation and lookup, queued job submission and lookup, job listing, audit retrieval, and findings retrieval through an explicit store interface. Runtime defaults to a durable SQLite database at `.shieldci/shieldci.db` through the installed `sqlite3` CLI; the in-memory implementation remains useful for isolated handler tests. The contract is documented in `backend/openapi.yaml`, and audit rows are populated by the Phase 3 worker.

## Phase 3 — job runner and GitHub Actions integration

Status: in progress — local queued-job worker, atomic job claiming, GitHub composite Action, and API-to-runner integration coverage are implemented; hosted integration tests are next.

Deliverables:

- Durable job state machine: queued, running, passed, failed, blocked, cancelled.
- Worker that receives a job, applies the policy, invokes the Warden adapter, stores output and audit events, and reports terminal status.
- Versioned GitHub Action that submits a job and waits for its result.

Gate: a fixture workflow runs a safe command successfully and a policy violation fails safely with a recorded reason.

Current progress: `internal/runner` atomically claims one queued job, invokes the Warden adapter, records adapter audit events, and finishes it as passed, failed, or blocked. Memory and SQLite stores both exercise the claim transition; concurrent memory claims prove only one worker receives a job. `cmd/shieldci/e2e_test.go` covers policy creation → job submission → worker execution → status and audit retrieval. The control-plane process starts this worker only when `SHIELDCI_WARDEN_BIN` is explicitly configured; otherwise jobs remain queued rather than running outside the sandbox. `cicd-security/actions/shieldci/action.yml` submits a policy, queues a job, polls its status, and fails the workflow for blocked or failed jobs.

## Phase 4 — scanning and security evidence

Status: in progress — scanner adapters, finding persistence, and findings API are implemented; end-to-end scanner/job linkage is next.

Deliverables:

- Semgrep adapter for SAST findings.
- Secrets detection adapter using tested fixture patterns.
- Finding model, severity normalization, deduplication, and audit linkage.

Gate: clean and vulnerable fixtures produce the expected findings, no secret values are stored in logs, and scanner failures are visible as errors.

Current progress: `internal/scanner` scans bounded source files, skips generated/dependency directories, detects tested secret patterns, and emits only rule, severity, location, and a stable fingerprint. The Semgrep adapter parses JSON reports into the same model. Findings can be persisted in SQLite and retrieved globally or by job; `POST /api/v1/scans/secrets` stores redacted results. Secret values are intentionally never included in finding messages.

## Phase 5 — dashboard MVP

Status: in progress — dashboard shell, API job wiring, and local API auth seam are implemented; stateful policy/finding screens are next.

Deliverables:

- Dashboard app connected to the API.
- Overview, jobs, job detail, policies, findings, and audit pages.
- Loading, empty, error, blocked, and narrow viewport states.
- Authentication boundary suitable for local development, with a provider seam for GitHub OAuth.

Gate: a user can submit or inspect a job, understand why it passed or was blocked, inspect a policy, and trace a finding back to an audit event.

Current progress: `landing/pages/dashboard.js` provides the first dark control-plane surface with workspace navigation, run status, policy events, coverage, and responsive narrow-viewport behavior. Dedicated `/dashboard/jobs`, `/dashboard/policies`, `/dashboard/findings`, and `/dashboard/audit` routes now exist. When `NEXT_PUBLIC_SHIELDCI_API` is set the overview loads live jobs, with representative fixtures as a safe empty/error state. The backend has an optional bearer-token boundary via `SHIELDCI_API_TOKEN`.

## Phase 6 — reliability, security, and release

Status: in progress — local and hosted verification gates pass. Manual hosted Warden proof run `42377771069` verified allowed execution and blocked undeclared filesystem access; the local container still cannot create the required egress-proxy Unix socket.

Deliverables:

- Full unit, integration, end-to-end, and Warden proof test suite.
- Database migrations, backups, retention rules, rate limits, and secret handling review.
- Docker Compose local stack and deployment documentation.
- CI checks for Go, dashboard, policy fixtures, and security regressions.

Gate: all required checks pass on a clean checkout and the MVP threat model has evidence for each high-risk boundary.

## MVP scope

Included: one organization, GitHub Actions integration, local and self-hosted runner path, YAML policies, Warden sandbox execution, basic SAST, secrets detection, audit events, jobs, findings, and a dashboard.

Deferred: GitLab and Jenkins integrations, advanced SCA and container scanning, enterprise SSO/SCIM, billing, multi-region deployment, AI-assisted detection, and broad multi-tenant isolation.

## Working rules

- Never run a job outside Warden when the requested sandbox cannot be established.
- Keep Warden changes isolated from ShieldCI product code unless a concrete integration defect requires a patch.
- Add a fixture before adding a security-sensitive behavior.
- Do not claim a feature works until its gate has a reproducible test result.
