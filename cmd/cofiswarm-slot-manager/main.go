package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/keepdevops/cofiswarm-slot-manager/internal/httpapi"
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
	log.Printf("slot-manager listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, srv.Handler()))
}
