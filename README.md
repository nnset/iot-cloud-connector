# Project still under development do not use in production

# IoT Cloud Connector
> Realtime communications with your IoT devices over the Internet.

Monitor and control your IoT devices using a simple and tiny helper tool that lets you 
code your own business logic using [Go](https://golang.org/).

![Global diagram](docs/images/global-diagram.jpg)

Current architecture is suitable for a single server instance.

## Quick start

### Concepts

| Name | Description |
| ------------- | ------------- |
| IoT device  | Sensors, actuators, gateways or any IoT device you want to monitor or even control, with Internet access |
| Cloud connector  | Handles, start and shutdown actions, also monitors memory usage from connections.  |
| REST API | A simple REST API where you query and send commands to your devices and also check Cloud Connector performance info. You can code your own API however we provide a default one. |
| Connections handler  | This is your code, here you define communications protocol and business logic. We have coded some samples such as a Web sockets handler in order to help you with this. |
| Command  | Send an order to an IoT Device. Commands may change device's state. |
| Query  | Fetch information from an IoT Device. Queries shall not change device's state, therefore they are easily cacheable in case you need it. |

If you have a strong Go or software development background, you may skip the rest and go to [how to write your own logic](#how-to-write-your-own-business-logic).

Also check our **default REST API** for Cloud Connector. [Documentation available here](/docs/default-cloud-connector-api.md).

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
    "github.com/nnset/iot-cloud-connector/storage"
    "github.com/sirupsen/logrus"
)

func main() {
    log := createLogger()

    // IoT devices will connect to localhost:8080
    connectionsHandler := connectionshandlers.NewSampleWebSocketsHandler(
        "localhost", "8080", "tcp", "", "",
    )

    cors := servers.CrossOriginResourceSharing{
        Headers: "Content-Type, Access-Control-Request-Method, Authorization",
        Origin:  "mysite.com",
    }

    // A default REST API will be available at localhost:9090
    // with no authorization required.
    // Check servers.authentication.go for more build in authentication methods
    defaultAPI := servers.NewDefaultCloudConnectorAPI("localhost", "9090", "", "", &servers.APINoAuthenticationMiddleware{}, &cors)

    s := servers.NewCloudConnector(
        log, connectionsHandler, storage.NewInMemoryDeviceConnectionsStorage(), defaultAPI,
    )

    s.Start()

    log.Debug("Finished shutdown")

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
    DEBU[0000] Starting REST API
    DEBU[0000] REST API available at localhost:9090        
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

> You can see this as the control layer. It starts and stops the service and helps CloudConnectorAPIInterface to fetch information.

```go
type CloudConnector struct {
    id                               string
    address                          string
    port                             string
    network                          string
    startTime                        int64
    log                              *logrus.Logger
    serverShutdownWaitGroup          sync.WaitGroup
    connectionsHandler               connectionshandlers.ConnectionsHandlerInterface
    statusAPI                        CloudConnectorAPIInterface
    state                            CloudConnectorState
    activeConnections                storage.DeviceConnectionsStorageInterface
    auth                             APIAuthMiddleWare
    systemMetricsStreamTicker        *time.Ticker
    systemMetricsStreamTickerDone    chan bool
    systemMetricsStreamSubscriptions map[chan SystemMetricChangedMessage]bool
}
```

**Features**

Starts all functional layers:

* ConnectionsHandler (Your business logic)
* CloudConnectorAPI (Your own or a [default one](/docs/default-cloud-connector-api.md))

Handles:

* Server shutdown when a os.signal is sent to stop the process. CloudConnector performs a graceful shutdown for all layers but, if this timeouts, shutdown is enforced.
* Holds Connections stats (via an instance of storage.DeviceConnectionsStorageInterface)
* Start REST API server and supports it making connection stats data available to it.

**How to instantiate it**

In order to instantiate it, you can use its named constructor and call *Start()*.

```go

    // Init log, connectionsHandler and defaultAPI before.

    connector := servers.NewCloudConnector(
        log, connectionsHandler, storage.NewInMemoryDeviceConnectionsStorage(), defaultAPI,
    )

    connector.Start()  // Blocks flow until a kill signal is sent to the process
```

Once started your IoT devices may connect using localhost port 9090 and using any protocol you defined 
in your own connectionsHandlerInterface implementation, for example websockets.

#### connections.DeviceConnection

> Information regarding each IoT device connection. All these fields are then used to report
how a connection is behaving.

```go
type DeviceConnection struct {
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

```go
type ConnectionsHandlerInterface interface {
    Start(shutdownChannel, shutdownIsCompleteChannel *chan bool, connections storage.DeviceConnectionsStorageInterface, log *logrus.Logger) error
    SendCommand(payload, deviceID string) (string, int, error)
    SendQuery(payload, deviceID string) (string, int, error)
    QueriesWaiting() uint
    CommandsWaiting() uint
}
```

On Start method these two channels are important:

| Name | Description |
| ------------- | ------------- |
| shutdownChannel  | ConnectionsHandler only reads from it. When a message is received means that service must shutdown so you should perform a graceful shutdown of all the connections and if you need, report the in progress shutdown to other services in your infrastructure. |
| shutdownIsCompleteChannel  | ConnectionsHandler on writes on it. A message must be added only when ConnectionsHandler has completed its shutdown procedure. |


#### storage.DeviceConnectionsStorageInterface

> How individual IoT devices connections stats are stored offering some global stats as well.

Check a ready to use thread safe in memory implementation at [storage.InMemoryDeviceConnectionsStorage](storage/inMemoryDeviceConnectionsStorage.go)


#### servers.CloudConnectorAPIInterface

> Each Cloud Connector instance has a build in REST API where users may fetch information regarding
connector status and current connections.

However you may want to offer your own API, thats fine just implement this interface and pass the instance
to Cloud Connector.

For authentication methods, we provide some options compatible with 
the Default API. Check [authentication.go](servers/authentication.go).

For [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) settings use
[CrossOriginResourceSharing](servers/authentication.go) struct

We provide a default REST API for Cloud Connector check the [documentation here](/docs/default-cloud-connector-api.md).

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

- An implementation of [storage.DeviceConnectionsStorageInterface](storage/deviceConnectionsStorageInterface.go)
    - Check [in memory implementation](storage/inMemoryDeviceConnectionsStorage.go) for an example.
- An implementation of servers.CloudConnectorAPIInterface
    - Check [default REST API](servers/defaultCloudConnectorAPI.go) for an example.


Finally type 

```
    $ go get

    $ go build

    $ ./your_app
```

# Links

Cloud connector uses these amazing projects:

- Gorilla Toolkit: https://www.gorillatoolkit.org/
- Logrus: https://github.com/sirupsen/logrus
- gotest.tools: https://github.com/gotestyourself/gotest.tools


# Licensing

Under MIT License, check [License file](./LICENSE)
