import socket

# TODO: Complete with a short-read/short-write tolerant implementation




def short_read(socket: socket.socket, size):
    data = b""
    while len(data) < size:
        data2 = socket.recv(size - len(data))
        if data2 == b"":
            break
        data += data2
    return data

def short_write(socket: socket.socket, bytes):
    total_sent = 0
    while total_sent < len(bytes):
        sent = socket.send(bytes[total_sent:])
        if sent == 0:
            raise RuntimeError("socket connection broken")
        total_sent += sent
    return total_sent

def recv_all(socket: socket.socket, size):
    return short_read(socket, size)

def send_all(socket: socket.socket, bytes):
    return short_write(socket, bytes)


