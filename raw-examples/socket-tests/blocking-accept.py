import os
import socket

if True:
    sd = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sd.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    sd.bind(('127.0.0.1', 8080))
    sd.listen(10)
    for i in range(3):
        # os.fork() == 0 means it's a child process
        if os.fork () == 0:
            x = 0
            while True:
                cd, _ = sd.accept() 
                req = cd.recv(1024).decode()
                cd.sendall(b"HTTP/1.1 200 OK\n\nHello World")
                cd.close()
                x += 1
                #print("Process: %d, request: %d", os.getpid(), x)
    os.wait()