//go:build ignore

#include <linux/bpf.h>

#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>


// Map of open service ports that we listen on for requests
struct bpf_map_def SEC("maps") echo_ports = {
	.type		= BPF_MAP_TYPE_HASH,
	.max_entries	= 1024,
	.key_size	= sizeof(__u32), // port number
	.value_size	= sizeof(__u64), // key for echo_socket
};

// Map of socket to which the listening ports forward traffic (in our case only one)
struct bpf_map_def SEC("maps") echo_socket = {
	.type		= BPF_MAP_TYPE_SOCKMAP,
	.max_entries	= 1,
	.key_size	= sizeof(__u32), // key for the socket
	.value_size	= sizeof(__u64), // socket file descriptor
};

// When invoked BPF sk_lookup program can select a socket that will receive the incoming packet 
// by calling the bpf_sk_assign() BPF helper function.
// Hooks for a common attach point (BPF_SK_LOOKUP) exist for both TCP and UDP.
SEC("sk_lookup/echo_dispatch")
int echo_dispatch(struct bpf_sk_lookup *ctx)
{
	const __u32 zero = 0;
	struct bpf_sock *sk;
	__u16 port;
	__u32 *open;
	long err;

	// Extract metadata from the request which can then be used to decide 
	// to which socket we want to forward traffic
	port = ctx->local_port;
	open = bpf_map_lookup_elem(&echo_ports, &port);
	if (!open)
		return SK_PASS;

	// Get the corresponding socket 
	// In our case we have only one socket in the map, but in general,
	// we could have multiple sockets and we would need to select the right one
	sk = bpf_map_lookup_elem(&echo_socket, &zero);
	if (!sk)
		return SK_DROP;

	// Assign the received packet/request to the socket
	err = bpf_sk_assign(ctx, sk, 0);
	// Release the reference held by sock
	bpf_sk_release(sk);
	// Selecting a socket only takes effect if the program has terminated with SK_PASS code.
	return err ? SK_DROP : SK_PASS;
}

SEC("license") const char __license[] = "Dual BSD/GPL";
