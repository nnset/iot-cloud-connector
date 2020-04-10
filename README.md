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

### Initial Configuration

Add this to your [go.mod](https://blog.golang.org/using-go-modules) file:

```
go 1.13

require (
    github.com/nnset/iot-cloud-connector
    github.com/sirupsen/logrus v1.4.2
)

```

### Quick example

Now you can start using IoT Cloud Connector, here's an example of main function, using one 
of our sample connections handlers (SampleWebSocketsHandler)

```go
package main

import (
    "os"
    "fmt"
    "syscall"
    "os/signal"

    "github.com/nnset/iot-cloud-connector/connectionshandlers"
    "github.com/nnset/iot-cloud-connector/servers"
    "github.com/sirupsen/logrus"
)

func main() {
    log := createLogger()
    operatingSystemSignal := make(chan os.Signal)
    shutdownServer := make(chan bool)

    signal.Notify(operatingSystemSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

    go func(log *logrus.Logger) {
        sig := <-operatingSystemSignal
        shutdownServer <- true
    }(log)

    connectionsHandler := connectionshandlers.NewSampleWebSocketsHandler(
        "localhost", "8080", "tcp", "", "",
    )

    s := servers.NewCloudConnector(
        "localhost", "9090", "tcp", log, &shutdownServer, connectionsHandler, nil
    )

    s.Start()
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
```

And run your app.


## Developing

Lets talk about the basic structs and interfaces.


### Structs

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
    serverShutdown     *chan bool
    connectionsHandler connectionshandlers.ConnectionsHandlerInterface
    statusAPI          StatusAPIInterface
    state              CloudConnectorState
}
```

CloudConnector offers some basic functionalities for handling IoT devices.
In order to initiate it, you can use its named constructor and call *Start()*.

```go

    // Init log, shutdown channel and connectionsHandler before.

    connector := servers.NewCloudConnector(
        "localhost", "9090", "tcp", log, &shutdownServerChannel, connectionsHandler, nil
    )

    connector.Start()  // Blocks flow until shutdownServerChannel is triggered
```

Once started your IoT devices may connect using localhost port 9090 and using any protocol you defined 
in your own connectionsHandlerInterface implementation, for example websockets.

**Features**

Starts all functional layers:

* ConnectionsHandler (Your business logic)
* StatusAPIInterface (Your own or a default one)
* ControlAPIInterface (Not available yet)

Handles service shutdown when a message is received from shutdownServerChannel.
CloudConnector performs a graceful shutdown for all layers but, if this timeouts, shutdown is forced.


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

#### connections.DeviceConnection

```go
type DeviceConnection struct {
    id            string
    deviceID      string
    deviceType    string
    userAgent     string
    remoteAddress string
    createdAt     int64
    connection    NetworkConnection
}

type NetworkConnection interface {
    Close(statusCode ConnectionStatusCode, reason string) error
}

```

### Interfaces

#### connectionshandlers.ConnectionsHandlerInterface

> Implementing your own ConnectionsHandlerInterface allows you to code any 
kind of business rules yo need to manage your IoT devices.

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

> You code how individual IoT devices connections stats are stored and also offering some global stats.

Check a ready to use thread safe in memory implementation at [storage.InMemoryDeviceConnectionsStatsStorage](storage/inMemoryDeviceConnectionsStatsStorage.go)


#### servers.StatusAPIInterface


### Connections Handlers samples


### How to write your own business logic






## Features

What's all the bells and whistles this project can perform?
* What's the main functionality
* You can also do another thing
* If you get really randy, you can even do this


## Links

Even though this information can be found inside the project on machine-readable
format like in a .json file, it's good to include a summary of most useful
links to humans using your project. You can include links like:

- Project homepage: https://your.github.com/awesome-project/
- Repository: https://github.com/your/awesome-project/
- Issue tracker: https://github.com/your/awesome-project/issues
  - In case of sensitive bugs like security vulnerabilities, please contact
    my@email.com directly instead of using issue tracker. We value your effort
    to improve the security and privacy of this project!
- Related projects:
  - Your other project: https://github.com/your/other-project/
  - Someone else's project: https://github.com/someones/awesome-project/


## Licensing

Under MIT License, check [License file](./LICENSE)











