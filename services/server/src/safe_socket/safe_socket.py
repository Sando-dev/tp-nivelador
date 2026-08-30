import socket

# TODO: Complete with a short-read/short-write tolerant implementation


def recv_all(socket: socket.socket, size):
    data = b""
    while len(data) < size:
        data2 = socket.recv(size - len(data))
        if data2 == b"":
            break
        data += data2
    return data


def send_all(socket: socket.socket, bytes):
    return socket.send(bytes)


