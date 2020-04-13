import sys
import json
import time
import signal
import random
# pip install websocket-client (https://github.com/websocket-client/websocket-client)
import websocket
from threading import Thread


class WebSocketClient:

    def __init__(self, host: str, port: str, socket_id: str, max_delay: int):
        self.client_name = 'Websockets client'
        self.__server_host = host
        self.__server_port = port
        self.__web_socket = None
        self.__web_socket_id = socket_id
        self.__max_delay = max_delay
        websocket.enableTrace(False)

    def run(self):
        self.__web_socket = websocket.WebSocketApp(
            'ws://{0}:{1}/connect'.format(self.__server_host, self.__server_port),
            on_message=self.on_message,
            on_error=self.on_error,
            header=["device_id: {0}".format(self.__web_socket_id)]
        )

        print("  Connecting with device_id = {}".format(self.__web_socket_id))
        self.__web_socket.on_open = self.on_open
        self.__web_socket.run_forever()

    def stop(self):
        print('Stopping {0} #{1}'.format(self.client_name, self.__web_socket_id))
        self.__web_socket.close()

    @staticmethod
    def on_message(ws, message):
        print('Message received {0}'.format(message))

    @staticmethod
    def on_error(ws, error):
        print(error)

    @staticmethod
    def on_open(ws):
        def run(*args):
            try:
                while True:
                    ws.send("Hello %d" % i)
                    time.sleep(random.randint(0, 10))
            except websocket.WebSocketConnectionClosedException:
                pass

        Thread(target=run).start()


if __name__ == "__main__":
    if len(sys.argv) < 4:
        print(
            'Missing arguments, usage:\n python websockets_client.py server_host server_port amount_of_sockets max_delay (optional)')
        sys.exit(1)

    server_host = (sys.argv + [''])[1]
    server_port = (sys.argv + [''])[2]
    amount_of_sockets = int((sys.argv + [''])[3])
    max_delay = int((sys.argv + ['0'])[4])
    websocket_clients = []

    finished = False

    def signal_handler(sig, frame):
        global finished
        print('Closing clients')

        for i, wsclient in enumerate(websocket_clients):
            wsclient.stop()

        finished = True
        print('All clients closed. Bye! (may take up to 10 second to exit, some Threads may be on time.sleep)')

    signal.signal(signal.SIGINT, signal_handler)

    i = 0
    while i < amount_of_sockets:
        print('Creating thread: websocket_{0}'.format(i))
        websocket_client = WebSocketClient(server_host, server_port, 'socket_{}'.format(i), max_delay)
        websocket_clients.append(websocket_client)
        thread = Thread(target=websocket_client.run)
        thread.start()
        i = i + 1

    # Do not use active waits in production code. Check python asyncio instead.
    while finished is False:
        time.sleep(1)

    sys.exit(0)
