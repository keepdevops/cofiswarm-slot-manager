package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/keepdevops/cofiswarm-slot-manager/internal/control"
	"github.com/keepdevops/cofiswarm-slot-manager/internal/evict"
	"github.com/keepdevops/cofiswarm-slot-manager/internal/httpapi"
	"github.com/keepdevops/cofiswarm-slot-manager/internal/kvclient"
	"github.com/keepdevops/cofiswarm-slot-manager/internal/pressure"
)

func main() {
	addr := flag.String("listen", ":8013", "listen address")
	cfg := flag.String("config", "", "endpoints json path")
	flag.Parse()
	if *cfg == "" {
		if v := os.Getenv("COFISWARM_SLOT_MANAGER_CONFIG"); v != "" {
			*cfg = v
		} else {
			*cfg = "/etc/cofiswarm/slot-manager/endpoints.json"
		}
	}
	srv, err := httpapi.New(*cfg)
	if err != nil {
		log.Printf("warn: endpoints config: %v (serving empty pressure)", err)
		srv, _ = httpapi.New("")
	}

	// Pressure-driven auto-clear loop (default-off). When COFISWARM_KVPOOL_URL is set, the
	// slot-manager snapshots pressure, asks kvpool what to do, and evicts when it fires.
	if kvURL := os.Getenv("COFISWARM_KVPOOL_URL"); kvURL != "" {
		go runControl(srv, kvURL)
		log.Printf("slot-manager control loop on (kvpool=%s)", kvURL)
	}

	log.Printf("slot-manager listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, srv.Handler()))
}

func runControl(srv *httpapi.Server, kvURL string) {
	interval := 10 * time.Second
	if v := os.Getenv("COFISWARM_CONTROL_INTERVAL_SECS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			interval = time.Duration(n) * time.Second
		}
	}
	deps := control.Deps{
		Snapshot: func() []control.Reading {
			eps := srv.Endpoints()
			return control.ReadingsFrom(pressure.Snapshot(eps), eps)
		},
		Evaluate: func(usage float64) (bool, bool) {
			d, ok := kvclient.Evaluate(kvURL, usage)
			if !ok {
				return false, false
			}
			return d.AutoClear, d.ProactiveEvict
		},
		Evict: func(r control.Reading) int {
			return evict.EndpointKV(r.Host, r.Port, r.Slots)
		},
	}
	control.Run(context.Background(), deps, interval)
}
