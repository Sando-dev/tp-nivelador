from enum import Enum
import safe_socket
from lottery import Bet


class MessageType(Enum):
    BET = 1
    FINISH = 2
    WINNER = 3
    BATCH = 4
    BATCH_OK = 5
    BATCH_ERROR = 6

class FieldType(Enum):
    AGENCY_ID = 1
    NAME = 2
    LAST_NAME = 3
    DOCUMENT = 4
    BIRTHDATE = 5
    NUMBER = 6


def receive_message(client_socket) -> (MessageType, bytes):
    message_type = safe_socket.recv_all(client_socket, 1)

    payload_length_bytes = safe_socket.recv_all(client_socket, 2)
    payload_length = int.from_bytes(payload_length_bytes, "big")

    payload = safe_socket.recv_all(client_socket, payload_length)

    return MessageType(message_type[0]), payload



def deserialize_bet(payload) -> bet.Bet:
    offset = 0
    fields = {}
    while offset < len(payload):
        field_type = FieldType(payload[offset])
        offset += 1

        field_length = int.from_bytes(payload[offset:offset + 2], "big")
        offset += 2

        field_value = payload[offset:offset + field_length]
        offset += field_length

        fields[field_type] = field_value
    
    agency_id = int.from_bytes(fields[FieldType.AGENCY_ID], "big")
    first_name = fields[FieldType.NAME].decode("utf-8")
    last_name = fields[FieldType.LAST_NAME].decode("utf-8")
    document = int.from_bytes(fields[FieldType.DOCUMENT], "big")
    birthdate = fields[FieldType.BIRTHDATE].decode("utf-8")
    number = int.from_bytes(fields[FieldType.NUMBER], "big")
    
    return Bet(
        agency_id,
        first_name, 
        last_name, 
        document, 
        birthdate, 
        number)

def serialize_field(field_type: FieldType, value: bytes) -> bytes:
    packet = bytearray()
    packet.append(field_type.value)
    packet.extend(len(value).to_bytes(2, "big"))
    packet.extend(value)
    return packet


def serialize_bet(bet: bet.Bet) -> bytes:
    payload = bytearray()

    payload.extend(serialize_field(FieldType.AGENCY_ID, bet.agency_id.to_bytes(4, "big")))
    payload.extend(serialize_field(FieldType.NAME, bet.first_name.encode("utf-8")))
    payload.extend(serialize_field(FieldType.LAST_NAME, bet.last_name.encode("utf-8")))
    payload.extend(serialize_field(FieldType.DOCUMENT, bet.document.to_bytes(4, "big")))
    payload.extend(serialize_field(FieldType.BIRTHDATE, bet.birthdate.encode("utf-8")))
    payload.extend(serialize_field(FieldType.NUMBER, bet.number.to_bytes(4, "big")))
    return payload

def serialize_message(message_type: MessageType, payload: bytes) -> bytes:
    packet = bytearray()
    packet.append(message_type.value)
    packet.extend(len(payload).to_bytes(2, "big"))
    packet.extend(payload)
    return packet


def deserialize_batch(payload) -> list[Bet]:
    bets = []
    offset = 0

    while offset < len(payload):
        if offset + 2 > len(payload):
            raise ValueError("invalid batch payload")

        bet_length = int.from_bytes(
            payload[offset:offset + 2],
            "big",
        )
        offset += 2

        if offset + bet_length > len(payload):
            raise ValueError("invalid bet length in batch")

        bet_payload = payload[offset:offset + bet_length]
        offset += bet_length

        bets.append(deserialize_bet(bet_payload))

    return bets


def serialize_batch_ok() -> bytes:
    return serialize_message(
        MessageType.BATCH_OK,
        b"",
    )


def serialize_batch_error() -> bytes:
    return serialize_message(
        MessageType.BATCH_ERROR,
        b"",
    )