// Package kvclient calls the cofiswarm-kvpool policy sidecar's /v1/evaluate to turn a KV
// pressure reading into a clear/evict decision. Used by the slot-manager control loop.
package kvclient

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type Decision struct {
	AutoClear      bool   `json:"auto_clear"`
	ProactiveEvict bool   `json:"proactive_evict"`
	Reason         string `json:"reason"`
}

// Evaluate asks kvpool what to do at the given KV pressure. Returns the decision and a
// boolean that is false on transport error (caller treats that as "do nothing").
func Evaluate(base string, kvPressure float64) (Decision, bool) {
	body, _ := json.Marshal(map[string]any{"kv_pressure": kvPressure})
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Post(base+"/v1/evaluate", "application/json", bytes.NewReader(body))
	if err != nil {
		return Decision{}, false
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return Decision{}, false
	}
	var d Decision
	if json.NewDecoder(resp.Body).Decode(&d) != nil {
		return Decision{}, false
	}
	return d, true
}
