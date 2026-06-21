package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/keepdevops/cofiswarm-observer-sdk/pkg/servicecomponent"
	"github.com/keepdevops/cofiswarm-slot-manager/internal/bus"
	"github.com/keepdevops/cofiswarm-slot-manager/internal/control"
	"github.com/keepdevops/cofiswarm-slot-manager/internal/evict"
	"github.com/keepdevops/cofiswarm-slot-manager/internal/httpapi"
	"github.com/keepdevops/cofiswarm-slot-manager/internal/kvclient"
	"github.com/keepdevops/cofiswarm-slot-manager/internal/pressure"
)

func main() {
	addr := flag.String("listen", ":8013", "listen address (HTTP mode)")
	cfg := flag.String("config", "", "endpoints json path")
	busMode := flag.Bool("bus", false, "serve .slots.* on the NATS observer bus instead of HTTP")
	natsURL := flag.String("nats", "nats://127.0.0.1:4222", "NATS URL (bus mode)")
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

	if *busMode {
		serveBus(*natsURL, srv)
		return
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

func serveBus(url string, srv *httpapi.Server) {
	nc, err := servicecomponent.Connect(url, "cofiswarm-slot-manager")
	if err != nil {
		log.Fatalf("bus connect %s: %v", url, err)
	}
	defer nc.Close()
	deps := bus.Deps{
		Endpoints: srv.Endpoints,
		Snapshot:  pressure.Snapshot,
		Evict:     evict.EndpointKV,
	}
	comp := servicecomponent.New(nc, "slot-manager", "slot-manager", bus.Routes(deps))
	if err := comp.Start(); err != nil {
		log.Fatalf("bus start: %v", err)
	}
	defer comp.Shutdown()
	log.Printf("slot-manager on bus %s (.slots.pressure/.slots.evict)", url)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Print("slot-manager bus stopping")
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
