package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Anhtran0208/redis-server-intro/internal/server"
)

func main() {
	var signals = make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	var wg sync.WaitGroup
	wg.Add(2)

	go server.RunIOMultiplexingServer(&wg)
	go server.WaitForSignal(&wg, signals)
	wg.Wait()
}
