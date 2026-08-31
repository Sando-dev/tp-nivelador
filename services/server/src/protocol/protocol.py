from enum import Enum
import safe_socket
import bet

class MessageType(Enum):
    BET = 1
    FINISH = 2
    WINNER = 3


def receive_message(client_socket):
    message_type = safe_socket.recv_all(client_socket, 1)
    payload_length = safe_socket.recv_all(client_socket, 2)
    payload = safe_socket.recv_all(client_socket, int.from_bytes(payload_length, "big"))
    return message_type, payload


def deserialize_int16(payload, offset):
    value = int.from_bytes(payload[offset:offset + 2], "big")
    return value, offset + 2

def deserialize_int32(payload, offset):
    value = int.from_bytes(payload[offset:offset + 4], "big")
    return value, offset + 4

def deserialize_string(payload, offset):
    length = int.from_bytes(payload[offset:offset + 2], "big")
    offset += 2
    value = payload[offset:offset + length].decode("utf-8")
    offset += length
    return value, offset

def deserialize_bet(payload):
    agency_id, offset = deserialize_int16(payload, 0)
    first_name, offset = deserialize_string(payload, offset)
    last_name, offset = deserialize_string(payload, offset)
    document, offset = deserialize_int32(payload, offset)
    birthdate, offset = deserialize_string(payload, offset)
    number, offset = deserialize_int32(payload, offset)
    
    return bet.Bet(
        agency_id,
        first_name, 
        last_name, 
        document, 
        birthdate, 
        number)

