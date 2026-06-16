# cofiswarm-slot-manager

Cofiswarm component: `slot-manager`.

- Layout: [REPO-STANDARD-LAYOUT](https://github.com/keepdevops/cofiswarmdev/blob/main/docs/REPO-STANDARD-LAYOUT.md)
- Migration: [MIGRATION-SPRINTS](https://github.com/keepdevops/cofiswarmdev/blob/main/docs/MIGRATION-SPRINTS.md)

## FHS paths

| Path | Purpose |
|------|---------|
| `/etc/cofiswarm/slot-manager/` | config |
| `/var/lib/cofiswarm/slot-manager/` | state |
| `/var/log/cofiswarm/slot-manager/` | logs |

## Test

```bash
./test/scripts/assert-layout.sh slot-manager
```
