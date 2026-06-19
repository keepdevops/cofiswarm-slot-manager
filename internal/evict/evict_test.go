package evict

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestEndpointKVErasesSlots(t *testing.T) {
	var mu sync.Mutex
	erased := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/slots/") || r.URL.Query().Get("action") != "erase" {
			t.Errorf("unexpected request %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		mu.Lock()
		erased++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	host, portStr, _ := net.SplitHostPort(strings.TrimPrefix(ts.URL, "http://"))
	port, _ := strconv.Atoi(portStr)

	if n := EndpointKV(host, port, 3); n != 3 {
		t.Fatalf("cleared = %d, want 3", n)
	}
	if erased != 3 {
		t.Fatalf("erased = %d, want 3", erased)
	}
}
