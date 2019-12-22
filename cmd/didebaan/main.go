package main

import (
	"log"
	"net"
	"time"

	didebaanpb "github.com/fzerorubigd/didebaan"
	"github.com/ogier/pflag"
	"google.golang.org/grpc"
)

func main() {
	var (
		port    string
		timeout time.Duration
		command string
		args    []string
	)

	pflag.StringVarP(&port, "port", "p", ":55055", "Port for GRPC listener")
	pflag.DurationVarP(&timeout, "timeout", "t", time.Second, "Timeout to wait for binary to die")

	pflag.Parse()

	args = pflag.Args()
	if len(args) < 1 {
		log.Fatal("You should provide the command to execute")
	}
	command, args = args[0], args[1:]

	var lc net.ListenConfig
	lis, err := lc.Listen(cliContext(), "tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	didebaanpb.RegisterTriggerServer(s, newServer(cliContext(), command, timeout, args...))

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	<-cliContext().Done()
	s.Stop()
}
