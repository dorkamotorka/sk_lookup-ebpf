Output of `ss -tpan | grep :8080` is:

```
LISTEN     0      10                    127.0.0.1:8080                 0.0.0.0:*     users:(("python3",pid=2834325,fd=3),("python3",pid=2834324,fd=3),("python3",pid=2834323,fd=3),("python3",pid=2834322,fd=3))
```
So multiple processes sharing the same socket. Don't get confused that each of these processes has a `fd=3`. Each process has its own process table although just a copy of the parents' one. But the main point is that all of these file descriptors (`fd=3`) point to the same file DESCRIPTION (a.k.a. entry in the open file table).

## Testing

```
ab -n 500 http://localhost:8080/
```