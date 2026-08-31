package protocol

import (
	"encoding/binary"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/bet"
)

type MessageType byte

const (
	MessageBet MessageType = iota + 1
	MessageFinish
	MessageWinners
)



func SerializeBet(b *bet.Bet) []byte {
	packet := make([]byte, 0)
	packet = append(packet, SerializeInt16(b.AgencyId)...)
	packet = append(packet, SerializeString(b.Name)...)
	packet = append(packet, SerializeString(b.LastName)...)
	packet = append(packet, SerializeInt32(b.DNI)...)
	packet = append(packet, SerializeString(b.Birthday)...)
	packet = append(packet, SerializeInt32(b.Amount)...)
	length := make([]byte, 2)
	binary.BigEndian.PutUint16(length, uint16(len(packet)))
	packet = append(length, packet...)
	return SerializeMessage(MessageBet, packet)
}

func SerializeString(value string) []byte {
    data := []byte(value)

    length := make([]byte, 2)
    binary.BigEndian.PutUint16(length, uint16(len(data)))

    return append(length, data...)
}

func SerializeInt16(value int) []byte {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, uint16(value))
	return data
}

func SerializeInt32(value int) []byte {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, uint32(value))
	return data
}


func SerializeMessage(messageType MessageType, payload []byte) []byte {
	packet := make([]byte, 0)
	packet = append(packet, byte(messageType))
	length := make([]byte, 2)
	binary.BigEndian.PutUint16(length, uint16(len(payload)))
	packet = append(packet, length...)
	packet = append(packet, payload...)

	return packet
}

func SerializeFinish() []byte {
	return SerializeMessage(MessageFinish, nil)
}