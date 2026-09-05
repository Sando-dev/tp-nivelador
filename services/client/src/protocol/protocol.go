package protocol

import (
	"net"
	"fmt"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/bet"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

type MessageType byte

const (
	MessageBet MessageType = iota + 1
	MessageFinish
	MessageWinners
)


type FieldType byte

const (
	FieldAgencyID FieldType = iota + 1
	FieldName
	FieldLastName
	FieldDocument
	FieldBirthdate
	FieldNumber
)

func SerializeBet(b *bet.Bet) []byte {
	packet := make([]byte, 0)
	packet = append(packet, SerializeField(FieldAgencyID, SerializeInt32(b.AgencyId))...)
	packet = append(packet, SerializeField(FieldName, []byte(b.Name))...)
	packet = append(packet, SerializeField(FieldLastName, []byte(b.LastName))...)
	packet = append(packet, SerializeField(FieldDocument, SerializeInt32(b.Document))...)
	packet = append(packet, SerializeField(FieldBirthdate, []byte(b.Birthday))...)
	packet = append(packet, SerializeField(FieldNumber, SerializeInt32(b.Number))...)

	return SerializeMessage(MessageBet, packet)
}

func SerializeField(fieldType FieldType, value []byte) []byte {
	packet := make([]byte, 0)
	packet = append(packet, byte(fieldType))
	length := len(value)
	packet = append(packet,
		byte(length>>8),
		byte(length),
	)
	packet = append(packet, value...)
	return packet
}


func SerializeInt16(value int) []byte {
	return []byte{
		byte(value >> 8),
		byte(value),
	}
}

func SerializeInt32(value int) []byte {
	return []byte{
		byte(value >> 24),
		byte(value >> 16),
		byte(value >> 8),
		byte(value),
	}
}



func SerializeMessage(messageType MessageType, payload []byte) []byte {
	packet := make([]byte, 0)
	packet = append(packet, byte(messageType))
	length := len(payload)
	packet = append(packet,
		byte(length>>8),
		byte(length),
	)
	packet = append(packet, payload...)

	return packet
}



func SerializeFinish() []byte {
	return SerializeMessage(MessageFinish, nil)
}


func ReceiveMessage(conn net.Conn) (MessageType, []byte, error) {
	messageTypeBytes, err := safe_socket.RecvAll(conn, 1)
	if err != nil {
		return 0, nil, err
	}

	payloadLengthBytes, err := safe_socket.RecvAll(conn, 2)
	if err != nil {
		return 0, nil, err
	}

	payloadLength :=
		int(payloadLengthBytes[0])<<8 |
		int(payloadLengthBytes[1])

	payload, err := safe_socket.RecvAll(conn, payloadLength)
	if err != nil {
		return 0, nil, err
	}

	messageType := MessageType(messageTypeBytes[0])

	return messageType, payload, nil
}


func DeserializeBet(payload []byte) (*bet.Bet, error) {
	var agencyId int
	var name string
	var lastName string
	var document int
	var birthday string
	var number int

	for len(payload) > 0 {
		if len(payload) < 3 {
			return nil, fmt.Errorf("invalid payload length")
		}

		fieldType := FieldType(payload[0])
		payload = payload[1:]

		fieldLength := int(payload[0])<<8 | int(payload[1])
		payload = payload[2:]

		if len(payload) < fieldLength {
			return nil, fmt.Errorf("invalid field length")
		}

		value := payload[:fieldLength]
		payload = payload[fieldLength:]

		switch fieldType {
		case FieldAgencyID:
			if fieldLength != 4 {
				return nil, fmt.Errorf("invalid agency ID length")
			}

			agencyId =
				int(value[0])<<24 |
				int(value[1])<<16 |
				int(value[2])<<8 |
				int(value[3])

		case FieldName:
			name = string(value)

		case FieldLastName:
			lastName = string(value)

		case FieldDocument:
			if fieldLength != 4 {
				return nil, fmt.Errorf("invalid document length")
			}

			document =
				int(value[0])<<24 |
				int(value[1])<<16 |
				int(value[2])<<8 |
				int(value[3])

		case FieldBirthdate:
			birthday = string(value)

		case FieldNumber:
			if fieldLength != 4 {
				return nil, fmt.Errorf("invalid number length")
			}

			number =
				int(value[0])<<24 |
				int(value[1])<<16 |
				int(value[2])<<8 |
				int(value[3])

		default:
			return nil, fmt.Errorf("invalid field type")
		}
	}

	return bet.NewBet(
		agencyId,
		name,
		lastName,
		document,
		birthday,
		number,
	), nil
}