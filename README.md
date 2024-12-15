# eBPF sk_lookup

This repository features a simple example where connections to different ports are redirected to the same listening socket.

While it can be expanded to handle multiple listening sockets, the main goal is to demonstrate how we can programmatically control the logic for mapping network connections to the desired listening socket.

## Build & Run

- Build the project:
```
go generate
go build
```

- Run the HTTP server. Use a script like `python3 raw-examples/socket-tests/blocking-accept.py`

- Find the PID and FD. Use the following commands:
```
lsof -i :8080
ss -lpt
```

- Run the eBPF program:
```
sudo ./inet-lookup-ebpf -pid <PID> -fd <FD>
```

- You can now send requests to the following ports: 

```
curl http://localhost:8081
curl http://localhost:8082
curl http://localhost:8083
```

**NOTE:** Ports `8081`, `8082`, and `8083` are hardcoded in the eBPF program, and requests will be redirected to the service with the provided PID.

