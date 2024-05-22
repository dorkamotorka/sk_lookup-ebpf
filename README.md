# Test

Run HTTP server, e.g. `python3 raw-examples/socket-tests/blocking-accept.py`

Figure out its PID using `lsof -i :8080`. Run eBPF code using `sudo ./inet-lookup-ebpf -pid <PID>`

Then you can make a request using 

```
curl http://localhost:8081
curl http://localhost:8082
curl http://localhost:8083
```

which is hardcoded in eBPF and the request will be redirected to the service with provided PID.

