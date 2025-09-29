package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/you/swipebot/internal/api"
	"github.com/you/swipebot/internal/obs"
	"github.com/you/swipebot/internal/worker"
)

func main() {
	base := envOr("API_BASE", "http://localhost:8080")
	token := os.Getenv("API_TOKEN")
	workers := 4
	maxBatch := 30

	client, err := api.NewClient(base, token)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", obs.Handler())
		addr := ":9090"
		log.Printf("Metrics available on %s/metrics", addr)
		_ = http.ListenAndServe(addr, mux)
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	pool := worker.NewPool(client, workers)
	pool.Start(ctx)

	go func() {
		for res := range pool.Results() {
			if res.Err != nil {
				log.Printf("swipe %s → %s ERROR: %v", res.ID, res.Action, res.Err)
				continue
			}
			log.Printf("swipe %s → %s | matched:%v", res.ID, res.Action, res.Matched)
		}
	}()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			pool.Stop()
			return
		case <-ticker.C:
			var cands []api.Candidate
			err := obs.Instrument(func() error {
				var e error
				cands, e = client.GetCandidates(ctx, maxBatch)
				return e
			})
			if err != nil {
				log.Printf("get candidates error: %v", err)
				continue
			}
			if len(cands) == 0 {
				log.Printf("no candidates; waiting…")
				continue
			}
			pool.EnqueueMany(cands)
		}
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
