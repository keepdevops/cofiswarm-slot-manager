// Package evict performs real KV eviction on a llama.cpp server by erasing its slots
// (POST /slots/{id}?action=erase). This replaces the coordinator_kv_ops C++ path.
package evict

import (
	"fmt"
	"net/http"
	"time"
)

// EndpointKV erases up to `slots` slots on the llama server at host:port. Returns the
// number of slots successfully erased. Best-effort: errors on individual slots are skipped.
func EndpointKV(host string, port, slots int) int {
	if slots < 1 {
		slots = 1
	}
	client := &http.Client{Timeout: 5 * time.Second}
	cleared := 0
	for id := 0; id < slots; id++ {
		url := fmt.Sprintf("http://%s:%d/slots/%d?action=erase", host, port, id)
		resp, err := client.Post(url, "application/json", nil)
		if err != nil {
			continue
		}
		_ = resp.Body.Close()
		if resp.StatusCode < 300 {
			cleared++
		}
	}
	return cleared
}
