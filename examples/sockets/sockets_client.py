import socket
import json
import time
import sys
import random
import datetime
import signal
from threading import Thread


class SocketClient:
    def __init__(self, host: str, port: str, socket_id: str, max_delay: int):
        self.client_name = 'Sockets client'
        self.__server_host = host
        self.__server_port = port
        self.__alive = True
        self.__socket = None
        self.__socket_id = socket_id
        self.__max_delay = max_delay

    def run(self):
        while self.__alive:
            try:
                if self.__socket is None:
                    self.__open_socket_with_server()

                response_data = self.__send_message(self.__build_message())

                if len(response_data) == 0:
                    print('Empty response from server')
                else:
                    clean_response = str(response_data, 'utf8').strip("'<>() ").replace('\'', '\"').replace("\\n", '')
                    json_response = json.loads(clean_response)
                    print(json.dumps(json_response, indent=2, sort_keys=True))

                time.sleep(self.__delay_until_next_message())

            except Exception as e:
                print(e)
                break

        print('closing socket {}'.format(self.__socket_id))
        self.__socket.close()

    def stop(self):
        print('Stopping client')
        self.__alive = False

    def __open_socket_with_server(self):
        print('Opening socket #{0} to server {1}:{2}'.format(self.__socket_id, self.__server_host,
                                                             self.__server_port))
        self.__socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.__socket.settimeout(2)
        self.__socket.connect((self.__server_host, int(self.__server_port)))

    def __build_message(self) -> dict:

        available_bodies = ['open-door', 'open-light', 'activate-all']

        return {
            'Sender': "{0}-{1}".format(self.client_name, self.__socket_id),
            'Body': random.choice(available_bodies),
            'Time': int(time.time())
        }

    def __send_message(self, message: dict):
        print('  {0} sending {1}'.format(self.__socket_id, message['Body']))

        self.__socket.sendall(json.dumps(message).encode())

        return self.__socket.recv(1024 * 1024)

    def __delay_until_next_message(self) -> int:
        """
            Returned delay is in seconds
        """
        return random.randint(0, self.__max_delay)


if __name__ == "__main__":
    server_host = (sys.argv + [''])[1]
    server_port = (sys.argv + [''])[2]
    amount_of_sockets = int((sys.argv + [''])[3])
    max_delay = int((sys.argv + [0])[4])
    socket_clients = []

    if len(sys.argv) < 4:
        print('Missing arguments, usage:\n python sockets_client.py server_host server_port amount_of_sockets max_delay (optional)')
        sys.exit(1)

    def signal_handler(sig, frame):
        print('Closing clients')

        for i, socket_client in enumerate(socket_clients):
            print('  Stoping client {}'.format(i))
            socket_client.stop()

        print('All clients closed. Bye!')
        sys.exit(0)

    signal.signal(signal.SIGINT, signal_handler)

    i = 0
    while i < amount_of_sockets:
        print('Creating thread: socket_{0}'.format(i))
        socket_client = SocketClient(server_host, server_port, 'socket_{}'.format(i), max_delay)
        socket_clients.append(socket_client)
        print('  Client created')
        thread = Thread(target=socket_client.run)
        thread.start()

        print('Created thread: socket_{0}'.format(i))
        i = i + 1

    time.sleep(5)
