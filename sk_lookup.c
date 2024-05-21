//go:build ignore

#include <linux/bpf.h>

#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>


// Map of open echo service ports.
struct bpf_map_def SEC("maps") echo_ports = {
	.type		= BPF_MAP_TYPE_HASH,
	.max_entries	= 1024,
	.key_size	= sizeof(__u16), // port number
	.value_size	= sizeof(__u8), // socket file descriptor
};

/* Echo server socket */
struct bpf_map_def SEC("maps") echo_socket = {
	.type		= BPF_MAP_TYPE_SOCKMAP,
	.max_entries	= 1,
	.key_size	= sizeof(__u32),
	.value_size	= sizeof(__u64),
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
	__u8 *open;
	long err;

	/* Is echo service enabled on packets destination port? */
	port = ctx->local_port;
	open = bpf_map_lookup_elem(&echo_ports, &port);
	if (!open)
		return SK_PASS;

	// Get the socket where the echo server is listening
	sk = bpf_map_lookup_elem(&echo_socket, &zero);
	if (!sk)
		return SK_DROP;

	// Select a socket to receive the packet
	err = bpf_sk_assign(ctx, sk, 0);
	// Release the reference held by sock
	bpf_sk_release(sk);
	// Selecting a socket only takes effect if the program has terminated with SK_PASS code.
	return err ? SK_DROP : SK_PASS;
}

SEC("license") const char __license[] = "Dual BSD/GPL";
