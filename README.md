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

## FHS

| Path | Purpose |
|------|---------|
| `/etc/cofiswarm/slot-manager/endpoints.json` | endpoint roster |
| `/var/lib/cofiswarm/slot-manager/` | state |
| `/var/log/cofiswarm/slot-manager/` | logs |

## Test

```bash
make test
```
