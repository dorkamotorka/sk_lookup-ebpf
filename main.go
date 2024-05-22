package main

import (
	"os"
	"log"
	"time"
	"flag"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
	"github.com/oraoto/go-pidfd"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go sklookup sk_lookup.c

// On the input provide the PID and FD of the process (e.g. of the HTTP server) to which we want to forward the traffic
var targetPid = flag.Int("pid", 0, "Target PID")
var targetFd = flag.Int("fd", 0, "Target FD")

// Utility function to load the key:value into the echo_ports BPF map
func insertEchoPort(key uint32, value uint64, echoPorts *ebpf.Map) error {
	if err := echoPorts.Put(&key, &value); err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Parse()

	// Remove resource limits for kernels <5.11.
	if err := rlimit.RemoveMemlock(); err != nil { 
		log.Print("Removing memlock:", err)
	}

	// Load the compiled eBPF ELF and load it into the kernel 
	// Also pins the map to /sys/fs/bpf/sklookup/globals
	// NOTE: we could also pin the eBPF program
	var objs sklookupObjects
	if err := loadSklookupObjects(&objs, &ebpf.CollectionOptions{Maps: ebpf.MapOptions{PinPath: "/sys/fs/bpf/sklookup/globals"}}); err != nil {
			log.Print("Error loading eBPF objects:", err)
	}
	defer objs.Close() // This only unloads the eBPF program (if it is not attached to kernel) and map, but doesn't remove the pin

	// Get current process network namespace, because we need to attach the eBPF program to it
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
		The biggest one is that PIDs are reused and this can happen quickly, which creates a race condition 
		where code that operates on a process (most often by sending it a signal) might end up performing an action 
		on the wrong process (in case it go the same PID as the process that was just shutdown).
	*/

	// Get the file descriptor to a process 
	// Remember that each process has it's own file descriptor table
	targetPidFd, err := pidfd.Open(*targetPid, 0)
	if err != nil {
		panic(err)
	}

	// Duplicate socket FD
	sockFd, err := targetPidFd.GetFd(*targetFd, 0)
	if err != nil {
		panic(err)
	}

	// Store the socket file descriptor to the echo_socket eBPF map
	var key uint32 = 0
	var val uint64 = uint64(sockFd)
	if err := objs.EchoSocket.Put(&key, &val); err != nil {
		panic(err)
	}

	// Now add some random port on which the packets will be redirected to choosen socket
	insertEchoPort(uint32(8081), uint64(0), objs.EchoPorts)
	insertEchoPort(uint32(8082), uint64(0), objs.EchoPorts)
	insertEchoPort(uint32(8083), uint64(0), objs.EchoPorts)

	log.Printf("Running eBPF program in the current process network namespace...")
	// We could exit here, if we would be pinning the eBPF program, but we just to keep the program running
	for {
		time.Sleep(1 * time.Second)
	}
}