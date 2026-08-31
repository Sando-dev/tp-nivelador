package safe_socket

import "io"

//TODO: Complete with a short-read/short-write tolerant implementation

func ShortWrite(socket io.Writer, bytes []byte) error {
	sent := 0
	for sent < len(bytes) {
		n, err := socket.Write(bytes[sent:])
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrUnexpectedEOF
		}
		sent += n
	}	
	return  nil
}

func ShortRead(socket io.Reader, size int) ([]byte, error) {
	buff := make([]byte, size)
	received := 0
	for received < size {
		n, err := socket.Read(buff[received:size])
		if err != nil {
			return nil, err
		}
		if n == 0 {
			return nil, io.ErrUnexpectedEOF
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
