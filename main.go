package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/ulrichheringer/service-discovery-lab/registry"
)

func main() {
	addr := flag.String("addr", ":8081", "HTTP listen address")
	interval := flag.Duration("check-interval", 5*time.Second, "health check interval")
	flag.Parse()
	r := registry.New(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go r.RunChecks(ctx, *interval)
	log.Printf("registry listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, r.Handler()))
}
