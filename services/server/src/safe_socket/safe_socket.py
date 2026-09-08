import socket


def short_read(socket: socket.socket, size):
    data = b""

    while len(data) < size:
        chunk = socket.recv(size - len(data))

        if chunk == b"":
            raise ConnectionError("socket connection closed")

        data += chunk

    return data

def short_write(socket: socket.socket, data):
    total_sent = 0

    while total_sent < len(data):
        sent = socket.send(data[total_sent:])
        total_sent += sent

    return total_sent

def recv_all(socket: socket.socket, size):
    return short_read(socket, size)

def send_all(socket: socket.socket, bytes):
    return short_write(socket, bytes)


