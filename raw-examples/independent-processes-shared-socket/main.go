package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"

	"github.com/oraoto/go-pidfd"
)

var pid = flag.Int("pid", 0, "Target PID")
var fd = flag.Int("fd", 0, "Target fd")

// NOTES
// This example is anything in particular special, just that file descriptor is shared/copied between processes 
// that arent in the parent-child relantionship.
// What you end up is, is two processes that are binded to the same listen socket 
// This socket obviously present a bottleneck since multiple processes need to be server but there's only one accept queue
// You end in the same situation if you spawn multiple child processes from a parent, where automatically
// each child gets the copy of the parent's file descriptor table, and thus the same socket.
func main() {
	flag.Parse()

	x := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		x += 1
		fmt.Printf("Request from process %d, x=%d\n", os.Getpid(), x)
	})

	listenAddr := ":8080"
	if *pid != 0 && *fd != 0 {
		// Start a server, listen & serve on a socket that already exists
		dupFdAndServe(*pid, *fd, handler)
	} else {
		// Start a server, open a new socket on the given `listenAddr`
		listenAndServe(listenAddr, handler)
	}
}

// Start a normal http server
func listenAndServe(listenAddr string, handler http.HandlerFunc) {
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			// Print listen FD and PID for later tests
			c.Control(func(fd uintptr) { fmt.Printf("Listening on %s, fd=%d, pid=%d\n", listenAddr, fd, os.Getpid()) })
			return nil
		},
	}
	ln, err := lc.Listen(context.Background(), "tcp", listenAddr)
	panicOnError(err)

	panicOnError(http.Serve(ln, http.HandlerFunc(handler)))
}

// Start a http server by duplicating the given FD in the given process
func dupFdAndServe(targetPid int, targetFd int, handler http.HandlerFunc) {
	p, err := pidfd.Open(targetPid, 0)
	panicOnError(err)

	// This can be checked in /proc/<pid>/fd
	fmt.Printf("File descriptor %d in the current process to refer to the other process %d\n", p, targetPid)

	listenFd, err := p.GetFd(targetFd, 0)
	panicOnError(err)

	ln, err := net.FileListener(os.NewFile(uintptr(listenFd), ""))
	panicOnError(err)

	// This spawns a new FD (probably need to go into the File Listener or the http.Serve to print it - that's why there are three processes)
	fmt.Printf("Duplicated the given socket FD and listening on it, pid=%d\n", os.Getpid())
	panicOnError(http.Serve(ln, http.HandlerFunc(handler)))
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}