package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/juandes/cool-new-project/internal/stats"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	service := stats.NewStatsService()
	defer service.Shutdown(ctx)

	// Listen for shutdown signals
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	for sig := range signals {
		cancel()
		log.Printf("Received %s, exiting.\n", sig.String())
		os.Exit(0)
	}
}
