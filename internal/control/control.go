// Package control is the slot-manager's pressure-driven auto-clear loop: it snapshots KV
// pressure, asks kvpool (/v1/evaluate) what to do, and triggers real eviction when the
// policy fires. This closes the loop slot-manager (pressure) -> kvpool (decision) ->
// slot-manager (evict) with no heartbeat — it acts only on observed pressure.
package control

import (
	"context"
	"log"
	"time"

	"github.com/keepdevops/cofiswarm-slot-manager/internal/pressure"
)

// Reading is the minimal pressure signal the loop needs per endpoint.
type Reading struct {
	EndpointID string
	Host       string
	Port       int
	Slots      int
	Usage      float64
}

// Deps are injected so the loop is testable without live servers/kvpool.
type Deps struct {
	Snapshot func() []Reading                          // current per-endpoint pressure
	Evaluate func(usage float64) (clear, evict bool)   // kvpool decision
	Evict    func(r Reading) int                        // perform eviction; returns slots cleared
}

// Tick runs one pass: evaluate each endpoint and evict where the policy fires.
// Returns the number of endpoints on which eviction was triggered.
func Tick(d Deps) int {
	fired := 0
	for _, r := range d.Snapshot() {
		clear, evict := d.Evaluate(r.Usage)
		if !clear && !evict {
			continue
		}
		n := d.Evict(r)
		fired++
		log.Printf("control: %s usage=%.2f -> %s, cleared %d slot(s)",
			r.EndpointID, r.Usage, decisionLabel(clear, evict), n)
	}
	return fired
}

// Run ticks on an interval until ctx is cancelled.
func Run(ctx context.Context, d Deps, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			Tick(d)
		}
	}
}

func decisionLabel(clear, evict bool) string {
	if clear {
		return "auto_clear"
	}
	if evict {
		return "proactive_evict"
	}
	return "none"
}

// ReadingsFrom converts a pressure snapshot into loop readings (drops entries with no usage).
func ReadingsFrom(entries []pressure.Entry, eps []pressure.Endpoint) []Reading {
	hostPort := map[string]pressure.Endpoint{}
	for _, ep := range eps {
		hostPort[ep.EndpointID] = ep
	}
	out := make([]Reading, 0, len(entries))
	for _, e := range entries {
		if e.Usage == nil {
			continue
		}
		ep := hostPort[e.EndpointID]
		out = append(out, Reading{
			EndpointID: e.EndpointID, Host: ep.Host, Port: e.Port,
			Slots: e.SlotsTotal, Usage: *e.Usage,
		})
	}
	return out
}
