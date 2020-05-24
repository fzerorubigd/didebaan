package main

import (
	"context"
	"sync"
	"time"

	didebaanpb "github.com/fzerorubigd/didebaan"
)

type server struct {
	lock sync.Mutex
	p    *process
}

func (s *server) Build(context.Context, *didebaanpb.TriggerRequest) (*didebaanpb.TriggerResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	ret := &didebaanpb.TriggerResponse{
		Status: didebaanpb.BuildStatus_BUILD_STATUS_INVALID,
	}
	if s.p.isInProgress() {
		ret.Status = didebaanpb.BuildStatus_BUILD_STATUS_ALREADY_STARTED
		return ret, nil
	}

	// The requset context is not used since it finishes with the request
	err := s.p.run(cliContext())
	if err != nil {
		ret.Status = didebaanpb.BuildStatus_BUILD_STATUS_FAILED
		ret.Message = err.Error()
		return ret, nil
	}

	ret.Status = didebaanpb.BuildStatus_BUILD_STATUS_SUCCESS
	return ret, nil
}

func newServer(ctx context.Context, command string, timeout time.Duration, args ...string) *server {
	p := newProcess(ctx, command, timeout, args...)
	return &server{
		p: p,
	}
}
