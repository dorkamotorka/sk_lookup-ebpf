package main

import (
	"os"
	"log"
	"time"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
	"github.com/oraoto/go-pidfd"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go sklookup sk_lookup.c

func main() {
	// Remove resource limits for kernels <5.11.
	if err := rlimit.RemoveMemlock(); err != nil { 
		log.Print("Removing memlock:", err)
	}

	// Load the compiled eBPF ELF and load it into the kernel 
	// Also pins the map to /sys/fs/bpf/sklookup/globals
	var objs sklookupObjects
	if err := loadSklookupObjects(&objs, &ebpf.CollectionOptions{Maps: ebpf.MapOptions{PinPath: "/sys/fs/bpf/sklookup/globals"}}); err != nil {
			log.Print("Error loading eBPF objects:", err)
	}
	defer objs.Close() // This only unloads the eBPF program (if it is not attached to kernel) and map, but doesn't remove the pin

	// Get self net-namespace
	netns, err := os.Open("/proc/self/ns/net")
	if err != nil {
		log.Fatal("Failed to read netns:", err)
	}
	defer netns.Close()

	// Attach the network namespace to the link
	// NOTE: Multiple programs can be attached to one network namespace. Programs will be invoked in the same order as they were attached.
	// This program finds a listening (TCP) or an unconnected (UDP) socket for an incoming packet.
	// Incoming traffic to established (TCP) and connected (UDP) sockets is delivered as usual without triggering the BPF sk_lookup hook.
	l, err := link.AttachNetNs(int(netns.Fd()), objs.EchoDispatch)
	if err != nil {
		log.Fatal("Failed to attach eBPF program to the network namespace:", err)
	}
	defer l.Close()

	/*
		Unix-like systems traditionally represent objects as files, but processes have always been an exception. 
		They are, instead, represented by process IDs (integer PID). There are a few problems with this representation.
		The biggest one is that PIDs are reused and this can happen quickly, 
		which creates a race condition where code that operates on a process 
		(most often by sending it a signal) might end up performing an action on the wrong process.
	*/

	// Get the file descriptor to a process 
	// Duplicate socket FD and store the socket to sockmap
	// Remember that each process has it's own file descriptor table
	// TODO: sharing file descriptor between processes?
	targetPid := 123602
	targetPidFd, err := pidfd.Open(targetPid, 0)
	if err != nil {
		panic(err)
	}

	targetFd := 3
	sockFd, err := targetPidFd.GetFd(targetFd, 0)
	if err != nil {
		panic(err)
	}

	var key uint32 = 0
	var val uint64 = uint64(sockFd)
	if err := objs.EchoSocket.Put(&key, &val); err != nil {
		panic(err)
	}

	log.Printf("Running eBPF program in the current process network namespace...")
	for {
		time.Sleep(1 * time.Second)
	}
}