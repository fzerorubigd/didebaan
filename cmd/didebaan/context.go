package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

var signals = []os.Signal{
	syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGABRT,
}

// cliContext returns a context that is cancelled automatically when a kill signal received
func cliContext() context.Context {
	var sig = make(chan os.Signal, 4)
	ctx, cancel := context.WithCancel(context.Background())
	signal.Notify(sig, signals...)
	go func() {
		<-sig
		cancel()
	}()

	return ctx
}
