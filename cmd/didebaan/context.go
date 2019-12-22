package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Context returns a context that is cancelled automatically when a kill signal received
func Context() context.Context {
	var sig = make(chan os.Signal, 4)
	ctx, cancel := context.WithCancel(context.Background())
	signal.Notify(sig, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGABRT)
	go func() {
		<-sig
		cancel()
	}()

	return ctx
}
