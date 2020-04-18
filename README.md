# Project still under development do not use in production

# IoT Cloud Connector
> Realtime communications with your IoT devices over the Internet.

Monitor and control your IoT devices using a simple and tiny helper tool that lets you 
code your own business logic using [Go](https://golang.org/).

![Global diagram](docs/images/global-diagram.jpg)


## Quick start

### Concepts

| Name | Description |
| ------------- | ------------- |
| IoT device  | Sensors, actuators, gateways or any IoT device you want to monitor or even control, with Internet access |
| Cloud connector  | Handles, start and shutdown actions, also monitors memory usage from connections.  |
| Status API | A simple REST API where you can fetch data from connected devices and theirs current status. |
| Control API | To do.  |
| Connections handler  | This is your code, here you define communications protocol and business rules. We have coded some samples such as a Web sockets handler in order to help you with this. |

If you have a strong Go or software development background, you may skip the rest and go to [how to write your own logic](#how-to-write-your-own-business-logic).

### Quick example

Now you can start using IoT Cloud Connector, here's an example of main function, using one 
of our sample connections handlers (SampleWebSocketsHandler)

```go
package main

import (
	"fmt"
	"os"

	"github.com/nnset/iot-cloud-connector/connectionshandlers"
	"github.com/nnset/iot-cloud-connector/servers"
	"github.com/sirupsen/logrus"
)

func main() {
	log := createLogger()

    // IoT devices will connect to localhost:8080
	connectionsHandler := connectionshandlers.NewSampleWebSocketsHandler(
		"localhost", "8080", "tcp", "", "",
	)

    // A default Status REST API will be available at localhost:9090
	s := servers.NewCloudConnector(
		"localhost", "9090", "tcp", log, connectionsHandler, nil,
	)

	s.Start()  // Blocks flow until server is shutdown via an os.signal

	os.Exit(0)
}

func createLogger() *logrus.Logger {
	var log = logrus.New()

	log.SetLevel(logrus.DebugLevel)
	log.Out = os.Stderr

	file, err := os.OpenFile("../var/log/sockets.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err == nil {
		log.Out = file
	} else {
		fmt.Println("Using stdErr for log")
	}

	return log
}

```

Run *go build* so all your dependencies are downloaded.

```shell
$ go build
$ ./websockets

    Using stdErr for log
    DEBU[0000] Starting CloudConnector #d24def51-0220-4a71-8da7-8eb3d4e93bb3 at localhost:9090 
    DEBU[0000] Starting StatusAPI                           
    DEBU[0000] StatusAPI available at localhost:9090        
    DEBU[0000] Serving websockets via ws (TLS OFF) at localhost:8080 
    DEBU[0000]   Connect endpoint ws://localhost:8080/connect 

```

Now you can use the [sample websocket client](examples/websockets/websockets_client.py) written in Python.

```shell
$ python3 websockets_client.py localhost 8080 3

    Creating thread: websocket_0
    Creating thread: websocket_1
    Creating thread: websocket_2

```

## Developing

### Initial Configuration

Add this to your [go.mod](https://blog.golang.org/using-go-modules) file:

```
go 1.13

require (
    github.com/nnset/iot-cloud-connector
    github.com/sirupsen/logrus v1.4.2
)

```

### Structs

Lets talk about the basic structs.

#### servers.CloudConnector

> You can see this as the control layer. It starts and stops the service and helps StatusAPIInterface to fetch information.

```go
type CloudConnector struct {
	id                 string
	address            string
	port               string
	network            string
	startTime          int64
	log                *logrus.Logger
	serverShutdown     chan bool
	connectionsHandler connectionshandlers.ConnectionsHandlerInterface
	statusAPI          StatusAPIInterface
	state              CloudConnectorState
}
```

**Features**

Starts all functional layers:

* ConnectionsHandler (Your business logic)
* StatusAPIInterface (Your own or a default one)
* ControlAPIInterface (Not available yet)

Handles service shutdown when a os.signal is send to stop the process.
CloudConnector performs a graceful shutdown for all layers but, if this timeouts, shutdown is enforced.

**How to instantiate it**

In order to instantiate it, you can use its named constructor and call *Start()*.

```go

    // Init log and connectionsHandler before.

    connector := servers.NewCloudConnector(
        "localhost", "9090", "tcp", log, connectionsHandler, nil
    )

    connector.Start()  // Blocks flow until a kill signal is sent to the process
```

Once started your IoT devices may connect using localhost port 9090 and using any protocol you defined 
in your own connectionsHandlerInterface implementation, for example websockets.

#### connections.DeviceConnectionStats

> Information regarding each IoT device connection. All these fields are then used to report
how a connection is behaving.

```go
type DeviceConnectionStats struct {
    connectionID                 string
    deviceID                     string
    deviceType                   string
    userAgent                    string
    remoteAddress                string
    createdAt                    int64
    lastReceivedMessageTimeStamp int64
    lastSentMessageTimeStamp     int64
    receivedMessages             uint
    sentMessages                 uint
}
```

### Interfaces

Use all these interfaces to fully customize your logic. We also offer some ready to use 
implementations.


#### connectionshandlers.ConnectionsHandlerInterface

> Implementing your own ConnectionsHandlerInterface allows you to code any 
kind of business rules you need to manage your IoT devices.

Only 2 methods are required by this interface:

```go
type ConnectionsHandlerInterface interface {
	Listen(shutdownChannel, shutdownIsCompleteChannel *chan bool, log *logrus.Logger) error

	Stats() storage.DeviceConnectionsStatsStorageInterface
}
```

On Listen method these two channels are important:

| Name | Description |
| ------------- | ------------- |
| shutdownChannel  | ConnectionsHandler only reads from it. When a message is received means that service must shutdown so you should perform a graceful shutdown of all the connections and if you need, report the in progress shutdown to other services in your infrastructure. |
| shutdownIsCompleteChannel  | ConnectionsHandler on writes on it. A message must be added only when ConnectionsHandler has completed its shutdown procedure. |


#### storage.DeviceConnectionsStatsStorageInterface

> How individual IoT devices connections stats are stored offering some global stats as well.

Check a ready to use thread safe in memory implementation at [storage.InMemoryDeviceConnectionsStatsStorage](storage/inMemoryDeviceConnectionsStatsStorage.go)


#### servers.StatusAPIInterface

> Each Cloud Connector instance has a build in REST API where users may fetch information regarding
connector status and current connections.

However you may want to offer your own API, thats fine just implement this interface and pass the instance
to Clod Connector.


### Connections Handlers samples

**websockets**

> Bidirectional communication using websockets

Cloud Connector [source code](examples/websockets/main.go).

Client [source code](examples/websockets/websockets_client.py).

Start Cloud Connector :

```shell
    ~/iot-cloud-connector/examples/websockets $ go build
    
    ~/iot-cloud-connector/examples/websockets $ ./websockets
```

Start client with 3 connections threads:

```shell
    ~/iot-cloud-connector/examples/websockets $ python3 websockets_client.py localhost 8080 3
```

# How to write your own business logic

**Required steps**

Add this to your [go.mod](https://blog.golang.org/using-go-modules) file:

```
go 1.13

require (
    github.com/nnset/iot-cloud-connector
    github.com/sirupsen/logrus v1.4.2
)

```

And then you must code :

- An implementation of [connectionshandlers.ConnectionsHandlerInterface](connectionshandlers/connectionsHandlerInterface.go)
    - Check [websockets sample](connectionshandlers/sampleWebsocketsHandler.go) for an example.

**Optional steps**

- An implementation of [storage.DeviceConnectionsStatsStorageInterface](storage/deviceConnectionsStatsStorageInterface.go)
    - Check [in memory implementation](storage/inMemoryDeviceConnectionsStatsStorage.go) for an example.
- An implementation of servers.StatusAPIInterface
    - Check [default status API](servers/defaultStatusAPI.go) for an example.


# Links

Cloud connector uses these amazing projects:

- Gorilla Toolkit: https://www.gorillatoolkit.org/
- Logrus: https://github.com/sirupsen/logrus
- gotest.tools: https://github.com/gotestyourself/gotest.tools


# Licensing

Under MIT License, check [License file](./LICENSE)
