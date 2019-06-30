package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

func gracefulStop(additional func()) {

	// Handle ^C and SIGTERM gracefully
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-gracefulStop
		Debug.Printf("Caught signal: %+v", sig)

		additional()

		time.Sleep(2 * time.Second)
		os.Exit(0)
	}()
}
