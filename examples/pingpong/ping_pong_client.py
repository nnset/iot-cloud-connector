import time
import sys
import random
import signal
import http.client


class PingPongClient:
    def __init__(self, host: str, port: str):
        self.client_name = 'Python ping pong client'
        self.__server_host = host
        self.__server_port = port
        self.__alive = True

    def run(self):
        print('Connecting to {0}:{1} as {2}'.format(self.__server_host, self.__server_port, self.client_name))

        while self.__alive:
            try:
                response = self.__send_ping()

                if response.status != 200:
                    print('Server response code: {0}, {1}'.format(response.status, response.read().decode('utf-8')))
                else:               
                    print('Response: {0}'.format(response.read().decode('utf-8')))
                
                time.sleep(random.randint(0, 3))

            except Exception as e:
                print(e)
                self.stop()

    def __send_ping(self) -> http.client.HTTPResponse:
        self.__connection = http.client.HTTPConnection(host=self.__server_host, port=self.__server_port)
        self.__connection.request('POST', '', 'PING')
        
        return self.__connection.getresponse()

    def stop(self):
        self.__connection.close()
        self.__alive = False


if __name__ == "__main__":
    server_host = (sys.argv + [''])[1]
    server_port = (sys.argv + [''])[2]

    if len(sys.argv) < 3:
        print('Missing arguments, usage:\n python ping_pong_client.py server_host server_port')
        sys.exit(1)

    client = PingPongClient(server_host, server_port)

    def signal_handler(sig, frame):
        print('Closing client')
        client.stop()
        print('Client closed. Bye!')
        sys.exit(0)

    signal.signal(signal.SIGINT, signal_handler)

    client.run()
