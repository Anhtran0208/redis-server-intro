package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Anhtran0208/redis-server-intro/internal/config"
	"github.com/Anhtran0208/redis-server-intro/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	var wg sync.WaitGroup
	wg.Add(2)

	switch cfg.Mode {
	// single thread mode - mode=single-thread
	case config.SingleThreadMode:
		log.Printf("Starting server in single thread mode on %s", cfg.Port)
		go server.RunSingleThreadServer(&wg, cfg)

	default:
		log.Fatalf("Unsupported execution mode: %s", cfg.Mode)
	}
	//s := server.NewServer()
	// go s.Start(&wg) // single listener
	//go s.StartMultiListeners(&wg) // multi listener
	go server.WaitForSignal(&wg, signals)
	wg.Wait()
}
