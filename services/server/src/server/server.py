import socket
import logger
import safe_socket
from protocol import protocol
from lottery import Bet, Lottery
import threading
import signal


class Server:
    def __init__(self, server_host: str, server_port: int, server_agency_quorum_min: int) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self.server_agency_quorum_min = server_agency_quorum_min
        self.finished_agencies = set()
        self.condition = threading.Condition()
        self.bets_lock = threading.Lock()
        self.shutdown_event = threading.Event()
        self.client_threads = []
        self.server_socket = None


    def _handle_client(self, client_socket):
        action = "handle-client"
        lottery_instance = Lottery("bets.csv")
        message_amount = 0
        client_agency_id = None

        try:
            logger.info(action, logger.LogResult.in_progress)

            while not self.shutdown_event.is_set():
                message_type, payload = protocol.receive_message(client_socket)

                if message_type == protocol.MessageType.BATCH:
                    message_amount += 1

                    try:
                        current_bets = protocol.deserialize_batch(payload)

                        if current_bets:
                            batch_agency_id = current_bets[0].agency_id
                            
                            if client_agency_id is None:
                                client_agency_id = batch_agency_id

                            elif client_agency_id != batch_agency_id:
                                raise ValueError(
                                    "agency changed during connection"
                                )

                        with self.bets_lock:
                            lottery_instance.store_bets(current_bets)
                    

                        response = protocol.serialize_batch_ok()
                        safe_socket.send_all(client_socket, response)

                    except Exception as e:
                        print(
                            f"BATCH_ERROR agency={client_agency_id} "
                            f"error={repr(e)}",
                            flush=True,
                        )

                        response = protocol.serialize_batch_error()
                        safe_socket.send_all(client_socket, response)

                elif message_type == protocol.MessageType.FINISH:
                    with self.condition:
                        self.finished_agencies.add(client_agency_id)

                        self.condition.notify_all()

                        self.condition.wait_for(
                            lambda:
                                len(self.finished_agencies)
                                >= self.server_agency_quorum_min
                                or self.shutdown_event.is_set()
                        )

                        if self.shutdown_event.is_set():
                            return

                    with self.bets_lock:
                        bets = list(lottery_instance.load_bets())

                    agency_bets = [
                        bet for bet in bets
                        if bet.agency_id == client_agency_id
                    ]

                    winners = [
                        bet for bet in agency_bets
                        if lottery_instance.has_won(bet)
                    ]

                    print(
                        f"AGENCY={client_agency_id} "
                        f"BETS={len(agency_bets)} "
                        f"WINNERS={len(winners)}"
                    )

                    for bet in bets:
                        if (
                            bet.agency_id == client_agency_id
                            and lottery_instance.has_won(bet)
                        ):
                            payload = protocol.serialize_bet(bet)

                            message = protocol.serialize_message(
                                protocol.MessageType.WINNER,
                                payload,
                            )

                            safe_socket.send_all(
                                client_socket,
                                message,
                            )

                    finish_message = protocol.serialize_message(
                        protocol.MessageType.FINISH,
                        b"",
                    )

                    safe_socket.send_all(
                        client_socket,
                        finish_message,
                    )

                    logger.info(
                        action,
                        logger.LogResult.success,
                        "messages-amount",
                        message_amount,
                    )

                    return

        except Exception as e:
            if not self.shutdown_event.is_set():
                logger.error(
                    action,
                    logger.LogResult.fail,
                    "messages-amount",
                    message_amount,
                )
                raise e

        finally:
            client_socket.close()
        

    def run(self):
        action = "accept-connection"
        signal.signal(signal.SIGTERM, self._signal_handler)

        with open("bets.csv", "w"):
            pass

        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:

            self.server_socket = server_socket

            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()

            while not self.shutdown_event.is_set():
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                except OSError as e:
                    if self.shutdown_event.is_set():
                        break
                    raise e
                
                logger.info(action, logger.LogResult.success)

                client_thread = threading.Thread(
                    target=self._handle_client, args=(client_socket,)
                )

                self.client_threads.append(client_thread)
                client_thread.start()
        for thread in self.client_threads:
            thread.join(timeout=2)

    def _signal_handler(self, signum, frame):
        self.shutdown_event.set()

        with self.condition:
            self.condition.notify_all()

        if self.server_socket is not None:
            self.server_socket.close()