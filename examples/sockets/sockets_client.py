import socket
import json
import time
import sys
import random
import datetime
import signal
from threading import Thread


class SocketClient:
    def __init__(self, host: str, port: str, socket_id: str):
        self.client_name = 'Sockets client'
        self.__server_host = host
        self.__server_port = port
        self.__alive = True
        self.__socket = None
        self.__socket_id = socket_id

    def run(self):
        while self.__alive:
            try:
                if self.__socket is None:
                    self.open_socket_with_server()

                response_data = self.send_message(self.build_message())

                if len(response_data) == 0:
                    print('Empty response from server')
                else:
                    clean_response = str(response_data, 'utf8').strip("'<>() ").replace('\'', '\"').replace("\\n", '')
                    json_response = json.loads(clean_response)
                    print(json.dumps(json_response, indent=2, sort_keys=True))

                time.sleep(self.delay_until_next_message())

            except Exception as e:
                print(e)
                break

        print('closing socket {}'.format(self.__socket_id))
        self.__socket.close()

    def open_socket_with_server(self):
        print('Opening socket #{0} to server {1}:{2}'.format(self.__socket_id, self.__server_host,
                                                             self.__server_port))
        self.__socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.__socket.settimeout(2)
        self.__socket.connect((self.__server_host, int(self.__server_port)))

    def build_message(self) -> dict:

        available_bodies = ['open-door', 'open-light', 'activate-all']

        return {
            'Sender': "{0}-{1}".format(self.client_name, self.__socket_id),
            'Body': random.choice(available_bodies),
            'Time': int(time.time())
        }

    def send_message(self, message: dict):
        print('  {0} sending {1}'.format(self.__socket_id, message['Body']))

        self.__socket.sendall(json.dumps(message).encode())

        return self.__socket.recv(1024 * 1024)

    def delay_until_next_message(self) -> int:
        """
            Returned delay is in seconds
        """
        return random.randint(0, 5)


if __name__ == "__main__":
    # i = 0
    #
    # aksk_for_server_status = (sys.argv + [''])[1] == 'status' or (sys.argv + [''])[1] == 'status-full'
    #
    # if aksk_for_server_status:
    #     socket_client = SocketClient(HOST, ADMIN_PORT, 98721, True)
    #
    #     thread = Thread(target=socket_client.run)
    #     thread.start()
    # else:
    #     while i < TOTAL_CONNECTIONS:
    #         socket_client = SocketClient(HOST, PORT, i, False)
    #         thread = Thread(target=socket_client.run)
    #         thread.start()
    #
    #         print('Created thread: #{0}'.format(i))
    #         i = i + 1
    #
    #         if HOST != 'localhost':
    #             time.sleep(1)
    server_host = (sys.argv + [''])[1]
    server_port = (sys.argv + [''])[2]
    amount_of_sockets = int((sys.argv + [''])[3])

    if len(sys.argv) < 4:
        print('Missing arguments, usage:\n python sockets_client.py server_host server_port amount_of_sockets')
        sys.exit(1)

    def signal_handler(sig, frame):
        print('Closing client')
        print('Client closed. Bye!')
        sys.exit(0)

    signal.signal(signal.SIGINT, signal_handler)

    i = 0

    while i < amount_of_sockets:
        print('Creating thread: socket_{0}'.format(i))
        socket_client = SocketClient(server_host, server_port, 'socket_{}'.format(i))
        print('  Client created')
        thread = Thread(target=socket_client.run)
        thread.start()

        print('Created thread: socket_{0}'.format(i))
        i = i + 1

    time.sleep(30)
