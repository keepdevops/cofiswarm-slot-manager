package pressure

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Endpoint struct {
	EndpointID string   `json:"endpoint_id"`
	Engine     string   `json:"engine"`
	Host       string   `json:"host"`
	Port       int      `json:"port"`
	Names      []string `json:"names"`
	DraftMax   int      `json:"draft_max"`
}

type Entry struct {
	Port        int      `json:"port"`
	Names       []string `json:"names"`
	Backend     string   `json:"backend"`
	OK          bool     `json:"ok"`
	Usage       *float64 `json:"usage"`
	KVUsed      *int64   `json:"kv_used"`
	KVTotal     *int64   `json:"kv_total"`
	SlotsBusy   int      `json:"slots_busy"`
	SlotsTotal  int      `json:"slots_total"`
	EndpointID  string   `json:"endpoint_id,omitempty"`
}

func Snapshot(endpoints []Endpoint) []Entry {
	out := make([]Entry, 0, len(endpoints))
	for _, ep := range endpoints {
		if ep.Engine == "mlx" {
			out = append(out, Entry{
				Port: ep.Port, Names: ep.Names, Backend: "mlx",
				OK: true, EndpointID: ep.EndpointID,
			})
			u := 0.0
			out[len(out)-1].Usage = &u
			continue
		}
		out = append(out, queryLlama(ep))
	}
	return out
}

func queryLlama(ep Endpoint) Entry {
	e := Entry{
		Port: ep.Port, Names: ep.Names, Backend: "llama",
		EndpointID: ep.EndpointID,
	}
	client := &http.Client{Timeout: 3 * time.Second}
	base := fmt.Sprintf("http://%s:%d", ep.Host, ep.Port)

	var nCtx int64
	var totalSlots int
	if body, ok := get(client, base+"/props"); ok {
		var j map[string]any
		if json.Unmarshal(body, &j) == nil {
			if v, ok := j["total_slots"].(float64); ok {
				totalSlots = int(v)
			}
			if gs, ok := j["default_generation_settings"].(map[string]any); ok {
				if v, ok := gs["n_ctx"].(float64); ok {
					nCtx = int64(v)
				}
			}
			if nCtx == 0 {
				if v, ok := j["n_ctx"].(float64); ok {
					nCtx = int64(v)
				}
			}
		}
	}

	var kvUsed int64
	var busy int
	if body, ok := get(client, base+"/slots"); ok {
		var slots []map[string]any
		if json.Unmarshal(body, &slots) == nil {
			totalSlots = len(slots)
			for _, s := range slots {
				if id, _ := s["id"].(float64); id >= 0 {
					if n, _ := s["n"].(float64); n > 0 {
						busy++
						kvUsed += int64(n)
					}
				}
			}
		}
	}

	if totalSlots < 1 {
		totalSlots = 1
	}
	kvTotal := nCtx * int64(totalSlots)
	e.SlotsBusy = busy
	e.SlotsTotal = totalSlots

	if kvTotal > 0 {
		u := float64(kvUsed) / float64(kvTotal)
		if u < 0 {
			u = 0
		}
		if u > 1 {
			u = 1
		}
		e.Usage = &u
		e.KVUsed = &kvUsed
		e.KVTotal = &kvTotal
		e.OK = true
	} else if busy > 0 && totalSlots > 0 {
		u := float64(busy) / float64(totalSlots)
		e.Usage = &u
		e.OK = true
	}
	return e
}

func get(client *http.Client, url string) ([]byte, bool) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, false
	}
	b, err := io.ReadAll(resp.Body)
	return b, err == nil
}
