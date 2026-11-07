# ShieldCI

ShieldCI is a concept for zero-trust security in CI/CD pipelines. The project includes product research and a landing page prototype.

## Project contents

- [`cicd-security/`](cicd-security/README.md) — product research, technical architecture, design, and MVP planning.
- [`landing/`](landing/README.md) — Next.js landing page prototype plus the dashboard MVP at `/dashboard`.
- [`backend/`](backend/go.mod) — ShieldCI control-plane foundation and policy contract.
- [`engine/warden/`](engine/warden/SHIELDCI_BASE.md) — pinned Warden execution base.
- [`PHASES.md`](PHASES.md) — ordered implementation phases and acceptance gates.

## Run the landing page

```bash
cd landing
npm ci
npm run dev
```

The development server runs at [http://localhost:3000](http://localhost:3000).

## Run the control plane

The backend requires Go 1.24 and the system `sqlite3` executable.

```bash
cd backend
go run -buildvcs=false ./cmd/shieldci
```

Set `SHIELDCI_ADDR` or `SHIELDCI_DB` to override the HTTP address or SQLite path. Set `SHIELDCI_API_TOKEN` to require `Authorization: Bearer ...` on API requests.

To enable the local worker, also set `SHIELDCI_WARDEN_BIN` to the pinned Warden binary. Jobs stay queued when it is unset.

The local API exposes `POST /api/v1/scans/secrets`, `GET /api/v1/findings`, and `GET /api/v1/jobs/{id}/findings` for the MVP security evidence path.

## Verification

Run the complete local gate from the repository root:

```bash
make verify
```

This runs backend unit, race, vet, and build checks; focused Warden policy/sandbox tests and build; and the dashboard production build.

The hosted workflow in `.github/workflows/verify.yml` also exposes a manual `warden-proof` job for real allow/deny sandbox verification. The local container may reject that proof because Warden needs to create an egress-proxy Unix socket.
