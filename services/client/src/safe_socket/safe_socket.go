package safe_socket

import "io"

//TODO: Complete with a short-read/short-write tolerant implementation

func ShortWrite(socket io.Writer, bytes []byte) (int, error) {
	sent := 0
	for sent < len(bytes) {
		n, err = socket.Write(bytes[sent:])
		if err != nil {
			return sent, err
		}
		sent += n
	}	
	return sent, nil
}

func ShortRead(socket io.Reader, size int) ([]byte, error) {
	buff := make([]byte, size)
	received := 0
	for received < size {
		n, err := socket.Read(buff[received:size])
		if err != nil {
			return nil, err
		}
		received += n
	}
	return buff, nil
}

func SendAll(socket io.Writer, bytes []byte) error {
	return ShortWrite(socket, bytes)
}

func RecvAll(socket io.Reader, size int) ([]byte, error) {
	return ShortRead(socket, size)
}
