// -*- mode: Go; fill-column: 110 -*-

package protocol

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"time"
)

type Profile struct {
	StartTime  time.Time
	FinishTime time.Time
	Commands   []ProfiledCommand
}

type ProfiledCommand struct {
	StartTime  time.Time
	FinishTime time.Time

	MakeLevel    uint
	MakeRestarts uint
	MakeDir      string

	RecipeTarget       string
	RecipeDependencies []string

	Args         []string
	ProcessState *os.ProcessState

	SubCommands []ProfiledCommand
}

type Listener interface {
	net.Listener
	SetDeadline(time.Time) error
}

type Logger interface {
	Printf(format string, a ...interface{}) (n int, err error)
}

type server struct {
	log Logger
}

func (srv server) master(listener net.Listener, cmds chan<- ProfiledCommand) error {
	var workers sync.WaitGroup
	var tempDelay time.Duration

	defer workers.Wait()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if netErr, isNetErr := err.(net.Error); isNetErr {
				if netErr.Timeout() {
					// This indicates that we used .SetDeadline(time.Now()) to
					// trigger a graceful shutdown.
					return nil
				}
				if netErr.Temporary() {
					// This is the same backoff algorithm as net/http.Server
					if tempDelay == 0 {
						tempDelay = 5 * time.Millisecond
					} else {
						tempDelay *= 2
					}
					if max := 1 * time.Second; tempDelay > max {
						tempDelay = max
					}
					srv.log.Printf("Accept error: %v; retrying in %v", err, tempDelay)
				}
			}
			return err
		}
		workers.Add(1)
		go func(conn net.Conn) {
			defer workers.Done()
			srv.worker(conn, cmds)
		}(conn)
	}
}

func (srv server) worker(conn net.Conn, cmds chan<- ProfiledCommand) {
	bs, err := ioutil.ReadAll(conn)
	if err != nil {
		srv.log.Printf("Connection i/o error: %v", err)
		return
	}
	var cmd ProfiledCommand
	err = json.Unmarshal(bs, &cmd)
	if err != nil {
		srv.log.Printf("Connection protocol error: %v", err)
		return
	}
	cmds <- cmd
}

func RunServer(ctx context.Context, listener Listener, log Logger) ([]ProfiledCommand, error) {
	cmdChan := make(chan ProfiledCommand)

	var cmdsLock sync.Mutex
	cmdsLock.Lock()
	var cmds []ProfiledCommand
	go func() {
		defer cmdsLock.Unlock()
		for cmd := range cmdChan {
			cmds = append(cmds, cmd)
		}
	}()

	errChan := make(chan error)
	go func() {
		srv := &server{log: log}
		errChan <- srv.master(listener, cmdChan)
		close(cmdChan)
	}()

	var returnErr error
	select {
	case <-ctx.Done():
		listener.SetDeadline(time.Now())
		returnErr = <-errChan
	case returnErr = <-errChan:
	}
	cmdsLock.Lock()
	return cmds, returnErr
}

func WithServer(listenerName string, log Logger, fn func()) ([]ProfiledCommand, error) {
	listener, err := net.Listen("unix", listenerName)
	if err != nil {
		return nil, err
	}
	defer listener.Close()

	// launch the server
	serverCtx, serverCancel := context.WithCancel(context.Background())
	var serverLock sync.Mutex

	serverLock.Lock()
	var serverErr error
	var serverCmds []ProfiledCommand
	go func() {
		defer serverLock.Unlock()
		serverCmds, serverErr = RunServer(serverCtx, listener.(Listener) /*TODO*/, nil)
	}()

	// run the function
	fn()

	// shut down the server
	serverCancel()
	serverLock.Lock()

	// return the results
	return serverCmds, serverErr
}
