package main

// Most of the code is from https://github.com/cespare/reflex/blob/master/reflex.go
// MIT Licensed.

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	pty "github.com/creack/pty"
)

type process struct {
	command string
	args    []string

	cmd *exec.Cmd
	tty *os.File

	cnl context.CancelFunc

	lock       sync.RWMutex
	running    bool
	killed     bool
	inProgress bool

	timeout time.Duration
}

func (p *process) run(ctx context.Context) error {
	p.setInProgress(true)
	defer p.setInProgress(false)

	if p.isRunnig() {
		p.terminate(ctx)
	}

	return p.start(ctx)
}

func (p *process) start(ctx context.Context) error {
	ctx, p.cnl = context.WithCancel(ctx)

	p.cmd = exec.CommandContext(ctx, p.command, p.args...)
	fl, err := pty.Start(p.cmd)
	if err != nil {
		return fmt.Errorf("failed to open the pty: %w", err)
	}
	p.tty = fl

	// Handle pty size.
	chResize := make(chan os.Signal, 1)
	signal.Notify(chResize, syscall.SIGWINCH)
	go func() {
		for {
			select {
			case <-chResize:
				// Intentionally ignore errors in case stdout is not a tty
				pty.InheritSize(os.Stdout, p.tty)
			case <-ctx.Done():
				signal.Stop(chResize)
				break
			}
		}
	}()
	chResize <- syscall.SIGWINCH // Initial resize.

	go func() {
		scanner := bufio.NewScanner(p.tty)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
			//stdout <- OutMsg{r.id, scanner.Text()}
		}
		// Intentionally ignoring scanner.Err() for now. Unfortunately,
		// the pty returns a read error when the child dies naturally,
		// so I'm just going to ignore errors here unless I can find a
		// better way to handle it.
	}()

	p.setRunning(true)
	go func() {
		err := p.cmd.Wait()
		if !p.isKilled() && err != nil {
			log.Printf("(error exit: %s)", err)
			//stdout <- OutMsg{r.id, fmt.Sprintf("(error exit: %s)", err)}
		}
		p.cnl()
		p.setRunning(false)
	}()

	return nil
}

func (p *process) terminate(ctx context.Context) {
	p.setKilled(true)
	// Write ascii 3 (what you get from ^C) to the controlling pty.
	// (This won't do anything if the process already died as the write will
	// simply fail.)
	p.tty.Write([]byte{3})

	timer := time.NewTimer(p.timeout)
	sig := syscall.SIGINT
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			if sig == syscall.SIGINT {
				log.Print("Sending SIGINT signal...")
			} else {
				log.Print("Sending SIGKILL signal...")
			}

			// Instead of killing the process, we want to kill its
			// whole pgroup in order to clean up any children the
			// process may have created.
			if err := syscall.Kill(-1*p.cmd.Process.Pid, sig); err != nil {
				log.Printf("Error killing: %s", err)
				if err.(syscall.Errno) == syscall.ESRCH { // no such process
					return
				}
			}
			// After SIGINT doesn't do anything, try SIGKILL next.
			timer.Reset(p.timeout)
			sig = syscall.SIGKILL
		}
	}
}

func (p *process) setRunning(b bool) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.running = b
}

func (p *process) isRunnig() bool {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.running
}

func (p *process) setKilled(b bool) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.killed = b
}

func (p *process) isKilled() bool {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.killed
}

func (p *process) setInProgress(b bool) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.inProgress = b
}

func (p *process) isInProgress() bool {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.inProgress
}

func newProcess(ctx context.Context, command string, timeout time.Duration, args ...string) *process {
	p := &process{
		command: command,
		args:    args,
		timeout: timeout,
	}

	_ = p.run(ctx)
	return p
}
