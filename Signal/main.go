package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Process ID: ", os.Getpid())

	// create channel to receive system signal
	signalChannel := make(chan os.Signal, 1)

	// register signal we want to listen
	// SIGINT and SIGTERM signal
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	// create channel to block main goroutine until shutdown is complete
	done := make(chan bool, 1)

	go func() {
		// deque signal
		currSignal := <-signalChannel
		fmt.Printf("\n\n[HANDLER] Received signal: %v\n", currSignal)
		// notify main goroutine that handler has finished and its safe to exit
		done <- true
	}()
	fmt.Println("[MAIN] waiting for work or signal...")
	<-done
	fmt.Println("Application shut down successfully")
}
