import socket
import logger
import safe_socket
from protocol import protocol
from lottery import Bet, Lottery


class Server:
    def __init__(self, server_host: str, server_port: int) -> None:
        self.server_host = server_host
        self.server_port = server_port

    def _handle_client(self, client_socket):
        action = "handle-client"
        list_of_bets = []
        lottery_instance = Lottery("bets.csv")
        message_amount = 0
        client_agency_id = None

        try:
            logger.info(action, logger.LogResult.in_progress)

            while True:
                message_type, payload = protocol.receive_message(client_socket)

                if message_type == protocol.MessageType.BATCH:
                    message_amount += 1

                    try:
                        current_bets = protocol.deserialize_batch(payload)

                        if len(current_bets) > 0 and client_agency_id is None:
                            client_agency_id = current_bets[0].agency_id

                        list_of_bets.extend(current_bets)

                        response = protocol.serialize_message(
                            protocol.MessageType.BATCH_OK,
                            b"",
                        )
                        safe_socket.send_all(client_socket, response)

                    except Exception:
                        response = protocol.serialize_message(
                            protocol.MessageType.BATCH_ERROR,
                            b"",
                        )
                        safe_socket.send_all(client_socket, response)

                elif message_type == protocol.MessageType.FINISH:
                    lottery_instance.store_bets(list_of_bets)

                    for bet in lottery_instance.load_bets():
                        if (
                            bet.agency_id == client_agency_id
                            and lottery_instance.has_won(bet)
                        ):
                            payload = protocol.serialize_bet(bet)

                            message = protocol.serialize_message(
                                protocol.MessageType.WINNER,
                                payload,
                            )

                            safe_socket.send_all(client_socket, message)

                    finish_message = protocol.serialize_message(
                        protocol.MessageType.FINISH,
                        b"",
                    )

                    safe_socket.send_all(client_socket, finish_message)

                    logger.info(
                        action,
                        logger.LogResult.success,
                        "messages-amount",
                        message_amount,
                    )

                    return

        except Exception as e:
            logger.error(
                action,
                logger.LogResult.fail,
                "messages-amount",
                message_amount,
            )
            raise e
        

    def run(self):
        action = "accept-connection"
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            while True:
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    raise e
                logger.info(action, logger.LogResult.success)

                self._handle_client(client_socket)
