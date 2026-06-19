package pressure

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestSnapshotComputesLlamaUsage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/props":
			_, _ = w.Write([]byte(`{"total_slots":2,"n_ctx":1000}`))
		case "/slots":
			_, _ = w.Write([]byte(`[{"id":0,"n":500},{"id":1,"n":0}]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()
	host, portStr, _ := net.SplitHostPort(strings.TrimPrefix(ts.URL, "http://"))
	port, _ := strconv.Atoi(portStr)

	got := Snapshot([]Endpoint{{EndpointID: "t", Engine: "llama", Host: host, Port: port}})
	if len(got) != 1 {
		t.Fatalf("entries = %d", len(got))
	}
	e := got[0]
	if !e.OK || e.SlotsTotal != 2 || e.SlotsBusy != 1 {
		t.Fatalf("entry = %+v", e)
	}
	if e.Usage == nil || *e.Usage < 0.24 || *e.Usage > 0.26 { // 500 / (1000*2)
		t.Fatalf("usage = %v, want ~0.25", e.Usage)
	}
}

func TestSnapshotMLXIsZeroPressure(t *testing.T) {
	got := Snapshot([]Endpoint{{EndpointID: "m", Engine: "mlx", Port: 8083}})
	if len(got) != 1 || got[0].Usage == nil || *got[0].Usage != 0 {
		t.Fatalf("mlx entry = %+v", got)
	}
}
