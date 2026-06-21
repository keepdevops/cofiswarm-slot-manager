package bus

import (
	"encoding/json"

	"github.com/keepdevops/cofiswarm-observer-sdk/pkg/servicecomponent"
	"github.com/keepdevops/cofiswarm-slot-manager/internal/pressure"
)

// Capability subjects (must match observer's bus/subjects.py).
const (
	SubjPressure = servicecomponent.Prefix + ".slots.pressure"
	SubjEvict    = servicecomponent.Prefix + ".slots.evict"
)

// Deps inject the slot-manager's logic so the routes are testable without live servers
// (mirrors internal/control.Deps).
type Deps struct {
	Endpoints func() []pressure.Endpoint
	Snapshot  func([]pressure.Endpoint) []pressure.Entry
	Evict     func(host string, port, slots int) int
}

// Routes wires the slot-manager to the .slots.* subjects. Reply field names mirror observer's
// bus/contracts/resource.py (PressureReply / EvictReply).
func Routes(d Deps) map[string]servicecomponent.Handler {
	return map[string]servicecomponent.Handler{
		SubjPressure: pressureHandler(d),
		SubjEvict:    evictHandler(d),
	}
}

func pressureHandler(d Deps) servicecomponent.Handler {
	return func([]byte) (any, error) {
		eps := d.Endpoints()
		hostByPort := make(map[int]string, len(eps))
		for _, ep := range eps {
			hostByPort[ep.Port] = ep.Host
		}
		entries := d.Snapshot(eps)
		out := make([]reading, 0, len(entries))
		for _, e := range entries {
			usage := 0.0
			if e.Usage != nil {
				usage = *e.Usage
			}
			out = append(out, reading{
				EndpointID: e.EndpointID, Host: hostByPort[e.Port],
				Port: e.Port, Slots: e.SlotsTotal, Usage: usage,
			})
		}
		return pressureReply{SchemaVersion: servicecomponent.SchemaVersion, OK: true, Readings: out}, nil
	}
}

func evictHandler(d Deps) servicecomponent.Handler {
	return func(data []byte) (any, error) {
		var req struct {
			EndpointID string `json:"endpoint_id"`
		}
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		eps := d.Endpoints()
		slotsBy := make(map[string]int)
		for _, e := range d.Snapshot(eps) {
			if e.EndpointID != "" {
				slotsBy[e.EndpointID] = e.SlotsTotal
			}
		}
		cleared, acted := 0, 0
		for _, ep := range eps {
			if req.EndpointID != "" && ep.EndpointID != req.EndpointID {
				continue
			}
			slots := slotsBy[ep.EndpointID]
			if slots < 1 {
				slots = 1
			}
			cleared += d.Evict(ep.Host, ep.Port, slots)
			acted++
		}
		return evictReply{SchemaVersion: servicecomponent.SchemaVersion, OK: true, Cleared: cleared, Endpoints: acted}, nil
	}
}

type reading struct {
	EndpointID string  `json:"endpoint_id"`
	Host       string  `json:"host"`
	Port       int     `json:"port"`
	Slots      int     `json:"slots"`
	Usage      float64 `json:"usage"`
}

type pressureReply struct {
	SchemaVersion string    `json:"schema_version"`
	OK            bool      `json:"ok"`
	Error         string    `json:"error,omitempty"`
	Readings      []reading `json:"readings"`
}

type evictReply struct {
	SchemaVersion string `json:"schema_version"`
	OK            bool   `json:"ok"`
	Error         string `json:"error,omitempty"`
	Cleared       int    `json:"cleared"`
	Endpoints     int    `json:"endpoints"`
}
