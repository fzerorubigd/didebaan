package main

import (
	"context"
	"log"
	"time"

	didebaanpb "github.com/fzerorubigd/didebaan"
	"github.com/ogier/pflag"
	"google.golang.org/grpc"
)

func main() {
	var (
		address string
	)
	pflag.StringVarP(&address, "server", "s", "localhost:55055", "Server to use")

	pflag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := didebaanpb.NewTriggerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r, err := c.Build(ctx, &didebaanpb.TriggerRequest{})
	if err != nil {
		log.Fatalf("could not build: %v", err)
	}

	switch r.Status {
	case didebaanpb.BuildStatus_BUILD_STATUS_INVALID:
		log.Fatal("Invalid response")
	case didebaanpb.BuildStatus_BUILD_STATUS_FAILED:
		log.Fatalf("Build failed with message: %q", r.Message)
	case didebaanpb.BuildStatus_BUILD_STATUS_SUCCESS:
		log.Print("Build successful")
	case didebaanpb.BuildStatus_BUILD_STATUS_ALREADY_STARTED:
		log.Fatal("Build already in progress")
	}
}
