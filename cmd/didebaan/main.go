package main

import (
	"context"
	"log"
	"net"
	"strings"
	"syscall"
	"time"

	"github.com/fzerorubigd/clictx"
	didebaanpb "github.com/fzerorubigd/didebaan"
	"github.com/ogier/pflag"
	"google.golang.org/grpc"
)

func cliContext() context.Context {
	return clictx.Context(
		syscall.SIGKILL,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGABRT,
	)
}

func main() {
	var (
		port    string
		timeout time.Duration
	)

	pflag.StringVarP(&port, "port", "p", ":55055", "Port for GRPC listener")
	pflag.DurationVarP(&timeout, "timeout", "t", time.Second, "Timeout to wait for binary to die")

	pflag.Parse()

	if len(pflag.Args()) < 1 {
		log.Fatal("You should provide the command to execute")
	}

	ctx := cliContext()
	all := strings.Join(pflag.Args(), " ")
	args := strings.Split(all, " ")

	var lc net.ListenConfig
	lis, err := lc.Listen(ctx, "tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	didebaanpb.RegisterTriggerServer(s, newServer(ctx, args[0], timeout, args[1:]...))

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	<-ctx.Done()
	s.Stop()
}
