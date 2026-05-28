package server

import (
	"log"
	"os"
	"sync"
	"sync/atomic"

	"github.com/Anhtran0208/redis-server-intro/internal/constant"
)

var serverStatus int32 = constant.ServerStatusIdle

func WaitForSignal(wg *sync.WaitGroup, signals chan os.Signal) {
	defer wg.Done()

	sig := <-signals
	log.Printf("Received signal: %s", sig)

	// wait for ongoing cmd to finish
	for {
		// compare current server status: if it's idle => shutting down
		if atomic.CompareAndSwapInt32(&serverStatus, constant.ServerStatusIdle, constant.ServerStatusShuttingDown) {
			log.Println("Shutting down gracefully")
			os.Exit(0) // shut down
		}
	}
}
