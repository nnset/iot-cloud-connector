# IoT Cloud Connector
A cloud server where you code how IoT devices send and receive messages.

## What IoT Cloud Connector does

Allows you to keep permanent connections between IoT devices (or any other kind of device) and a cloud server.

## Who may use it

This tool is for developers as it does not have any user interface and its functionality
has to be coded in [Go programming language](https://golang.org/).

## How to use it

There are 3 main concepts :
- CloudConnector
- Server
- DeviceConnections
- Messages


### CloudConnector

This struct has all the control logic regarding the IoT devices conenctios, such as:
    - Active connections
    - Memory usage
    - Network usage
    - Incoming messages and outgoing messages stats
    - Server start and shutdow
    - Closing connections
    - Logging

### Server

This is where you implement all your logic; defining how conenctions are established,
authenticated, how messages are handled and even what kind of network protocol is used.
All servers MUST implement the *ServerInterface* interface in order to be compatible
with any *CloudConnector* instance.

We provide some ready to use servers, such as a Web sockets server where yo just pass
a couple of functions to handle authentication, messages and disconnections. 
Check the example below.

### DeviceConnections

All IoT connections handled by a *CloudConnector* instance, must be *DeviceConnection* structs.
*DeviceConnection* structs have a field that MUST implement *NetworkConnection* interface,
so *CloudConnector* may close the connection, this interface will allow you to use
any kind of communication protocol.

### Messages

These are the payloads that clients (IoT devices) and your server (*ServerInterface* handled by a *CloudConnector*) exchange.

## Build in servers

TODO

### Websockets server

TODO

## How to code your own servers

TODO

## Usage

TODO
