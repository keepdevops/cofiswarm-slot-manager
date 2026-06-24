package bus

import (
	"encoding/json"
	"testing"

	"github.com/keepdevops/cofiswarm-slot-manager/internal/pressure"
)

func fakeDeps(evicted *int) Deps {
	eps := []pressure.Endpoint{
		{EndpointID: "e1", Host: "h1", Port: 8086},
		{EndpointID: "e2", Host: "h2", Port: 8087},
	}
	usage := 0.9
	return Deps{
		Endpoints: func() []pressure.Endpoint { return eps },
		Snapshot: func([]pressure.Endpoint) []pressure.Entry {
			return []pressure.Entry{
				{EndpointID: "e1", Port: 8086, SlotsTotal: 4, Usage: &usage},
				{EndpointID: "e2", Port: 8087, SlotsTotal: 2, Usage: &usage},
			}
		},
		Evict: func(host string, port, slots int) int { *evicted += slots; return slots },
	}
}

func TestPressureMapsReadings(t *testing.T) {
	out, err := Routes(fakeDeps(new(int)))[SubjPressure](nil)
	if err != nil {
		t.Fatal(err)
	}
	r := out.(pressureReply)
	if len(r.Readings) != 2 || r.Readings[0].Host != "h1" || r.Readings[0].Slots != 4 {
		t.Fatalf("got %+v", r.Readings)
	}
	if r.Readings[0].Usage != 0.9 {
		t.Fatalf("usage not mapped: %+v", r.Readings[0])
	}
}

func TestEvictAllEndpoints(t *testing.T) {
	evicted := 0
	out, _ := Routes(fakeDeps(&evicted))[SubjEvict]([]byte(`{}`))
	r := out.(evictReply)
	if r.Endpoints != 2 || r.Cleared != 6 { // 4 + 2 slots
		t.Fatalf("got %+v", r)
	}
}

func TestEvictSingleEndpoint(t *testing.T) {
	evicted := 0
	out, _ := Routes(fakeDeps(&evicted))[SubjEvict]([]byte(`{"endpoint_id":"e2"}`))
	r := out.(evictReply)
	if r.Endpoints != 1 || r.Cleared != 2 {
		t.Fatalf("got %+v", r)
	}
}

func TestEvictReplyFieldNames(t *testing.T) {
	out, _ := Routes(fakeDeps(new(int)))[SubjEvict]([]byte(`{}`))
	b, _ := json.Marshal(out)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	for _, k := range []string{"schema_version", "ok", "cleared", "endpoints"} {
		if _, ok := m[k]; !ok {
			t.Fatalf("missing %q in %s", k, b)
		}
	}
}
