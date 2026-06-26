package httpapi

import (
	"os"
	"path/filepath"
	"testing"
)

const endpointsJSON = `{"endpoints":[
  {"endpoint_id":"coder7b","engine":"llama","host":"127.0.0.1","port":8086},
  {"endpoint_id":"llama8b","engine":"llama","host":"127.0.0.1","port":8085}
]}`

func writeEndpoints(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "endpoints.json")
	if err := os.WriteFile(p, []byte(endpointsJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// Without COFISWARM_ENDPOINT_HOST the repo config's hosts are preserved verbatim.
func TestLoadEndpointsNoOverride(t *testing.T) {
	t.Setenv("COFISWARM_ENDPOINT_HOST", "")
	s, err := New(writeEndpoints(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range s.Endpoints() {
		if e.Host != "127.0.0.1" {
			t.Fatalf("host = %q, want 127.0.0.1 (no override)", e.Host)
		}
	}
}

// COFISWARM_ENDPOINT_HOST rewrites every endpoint host (containerized deploy -> host inference),
// leaving ports/ids intact.
func TestLoadEndpointsHostOverride(t *testing.T) {
	t.Setenv("COFISWARM_ENDPOINT_HOST", "host.docker.internal")
	s, err := New(writeEndpoints(t))
	if err != nil {
		t.Fatal(err)
	}
	eps := s.Endpoints()
	if len(eps) != 2 {
		t.Fatalf("endpoints = %d, want 2", len(eps))
	}
	for _, e := range eps {
		if e.Host != "host.docker.internal" {
			t.Fatalf("host = %q, want host.docker.internal", e.Host)
		}
	}
	if eps[0].Port != 8086 || eps[0].EndpointID != "coder7b" {
		t.Fatalf("override clobbered other fields: %+v", eps[0])
	}
}
