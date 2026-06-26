# cofiswarm-slot-manager

Physical concurrency + KV pressure for inference endpoints (`endpoint_id` / `server_group`).

- Layout: [REPO-STANDARD-LAYOUT](https://github.com/keepdevops/cofiswarm-docs/blob/main/REPO-STANDARD-LAYOUT.md)
- Migration: Sprint 6 in [MIGRATION-SPRINTS](https://github.com/keepdevops/cofiswarm-docs/blob/main/MIGRATION-SPRINTS.md)
- Legacy C++ reference: `legacy/cpp/` (coordinator bridge until cutover)

## HTTP

| Route | Description |
|-------|-------------|
| `GET /healthz` | Liveness |
| `GET /api/pressure` | Per-endpoint KV pressure (parity with coordinator) |
| `POST /api/pressure/evict` | Targeted eviction (stub → port `coordinator_kv_ops`) |
| `GET /v1/endpoints` | Registered endpoints |

Default listen: `:8013` (`cofiswarm-common/ports/well-known.yaml`).

## Build & run

```bash
make build
./bin/cofiswarm-slot-manager -config configs/endpoints.json
```

## Configuration

Standalone — no FHS coupling. The endpoint roster is resolved in this order:

| Source | Notes |
|--------|-------|
| `-config <path>` flag | highest precedence |
| `COFISWARM_SLOT_MANAGER_CONFIG` env | for containers/host runners |
| `configs/endpoints.json` (repo-relative) | default; ships with the binary |

Nothing is read from or written to `/etc/cofiswarm`, `/var/lib/cofiswarm`, or
`/var/log/cofiswarm`.

`COFISWARM_ENDPOINT_HOST` (optional) rewrites every endpoint host after load — a
containerized deploy fronting host inference sets it to `host.docker.internal` so the
roster resolves to the host without editing the repo config (which stays `127.0.0.1` for
host/native runs). Mirrors the `COFISWARM_AGENT_HOST` override dispatch/modes use.

## Test

```bash
make test
```
