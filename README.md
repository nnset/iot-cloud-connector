# Project still under development do not use in production

# IoT Cloud Connector
> Realtime communications with your IoT devices over the Internet.

![Problem to solve](docs/images/problem-brief.jpg)

Monitor and control your IoT devices using a simple and tiny helper cloud tool, that lets you 
code your own business logic using [Go](https://golang.org/) programming language.


## Main Concepts

| Name | Description |
| ------------- | ------------- |
| IoT Device  | Sensors, actuators, gateways or any IoT device you want to monitor or even control, with Internet access |
| Cloud Connector  | Handles: start and shutdown actions, also monitors memory usage from connections.  |
| REST API | A simple REST API where you query and send commands to your devices and also check Cloud Connector performance info. You can code your own API, however we provide a default one. |
| Connections Handler  | This is your code, here you define communications protocol and business logic. We have coded some samples such as a Web sockets handler in order to help you with this. |
| Command  | Send an order to an IoT Device. Commands may change device's state. |
| Query  | Fetch information from an IoT Device. Queries shall not change device's state, therefore they are easily cacheable in case you need it. Queries should not alter device's state. |
| SSE (Event stream)  | [Server Sent Events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events) that allows you to listen in real time to what is happening with your IoT devices. |

## Conventions

- All communications are asynchronous by default.
- If you want to alter the current state of your IoT devices you **send them a Command**.
- If you want to retrieve data from your IoT devices you **send them a Query**.
- IoT devices, may send data to the cloud whenever they want, you just **handle these messages**.

## Features

| Name | Description |
| ------------- | ------------- |
| Monitor  | Monitor how many messages your cloud and your IoT devices ar exchanging and how many server memory connections are consuming. |
| Control  | Send commands to your IoT devices and Query them for retrieving data at your will.  |
| Customize  | Code your own business logic, Cloud Connector is just a low footprint facilitator. |
| Default API | We provide a default REST API to handle your IoT devices and server. |
| Default Websockets handler | If websockets its your thing, we provide a connections handler for you so you only have to worry about business logic regarding messages. |
| Vanilla UI  | We provide a simple UI using vanilla Javascript that helps you check if your server is doing fine, also allows to send Commands and Queries to your IoT devices with the payload you want. |

![Vanilla UI Dashboard](docs/images/vanilla-ui-dashboard.jpg)
![Vanilla UI Devices](docs/images/vanilla-ui-device.jpg)

## Usage

![Global diagram](docs/images/global-diagram.jpg)

What you **must** code to make it work:

1. A **connections handler** that will define the communications protocol between your IoT devices and Cloud Connector.

Your handlers must implement [connectionshandlers.ConnectionsHandlerInterface](connectionshandlers/connectionsHandlerInterface.go). Check [websockets sample](connectionshandlers/sampleWebsocketsHandler.go) for an example.


2. Your own **package** or main.go file:

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
    defaultAPI := servers.NewDefaultCloudConnectorAPI(
        "localhost", 
        "9090", 
        "", 
        "", 
        &servers.APINoAuthenticationMiddleware{}, 
        &cors
    )

    s := servers.NewCloudConnector(
        log, connectionsHandler, storage.NewInMemoryDeviceConnectionsStorage(), defaultAPI,
    )

    s.Start() // Will block flow until an operating system signal is received

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

3. A **modules definition** for your project [go.mod](https://blog.golang.org/using-go-modules) file:

```
go 1.13

require (
    github.com/nnset/iot-cloud-connector
    github.com/sirupsen/logrus v1.4.2
)

```

4. Finally type:

```
    $ go get

    $ go build

    $ ./your_app
```

**Optional steps**

- An implementation of [storage.DeviceConnectionsStorageInterface](storage/deviceConnectionsStorageInterface.go)
    - Check [in memory implementation](storage/inMemoryDeviceConnectionsStorage.go) for an example.
- An implementation of servers.CloudConnectorAPIInterface
    - Check [default REST API](servers/defaultCloudConnectorAPI.go) for an example.

## Developing your own logic

### Structs

#### servers.CloudConnector

> You can see this as the control layer. It starts and stops the service and helps CloudConnectorAPIInterface to fetch information.

Source [cloudConnector.go](servers/cloudConnector.go)

**Features**

Starts all functional layers:

* ConnectionsHandler (Your business logic)
* CloudConnectorAPI (Your own or a [default one](/docs/default-cloud-connector-api.md))

Handles:

* Server shutdown when a os.signal is sent to stop the process. CloudConnector performs a graceful shutdown for all layers but, if this timeouts, shutdown is enforced.
* Holds Connections stats (via an instance of storage.DeviceConnectionsStorageInterface)
* Start REST API server and supports it making connection stats data available to it.


#### connections.DeviceConnection

> Information regarding each IoT device connection. All these fields are then used to report
how a connection is behaving.

Source [deviceConnection.go](connections/deviceConnection.go)


#### connections.DeviceConnectionDTO

> A Data Transfer Object that encapsulates a device connection information.

Source [deviceConnectionDTO.go](connections/deviceConnectionDTO.go)


### Interfaces

Use all these interfaces to fully customize your logic. We also offer some ready to use 
implementations.


#### connectionshandlers.ConnectionsHandlerInterface

> Implementing your own ConnectionsHandlerInterface allows you to code any 
kind of business rules you need to manage your IoT devices.

Source [connectionsHandlerInterface.go](connectionshandlers/connectionsHandlerInterface.go)


For Start() method, these two channels are important:

| Name | Description |
| ------------- | ------------- |
| shutdownChannel  | ConnectionsHandler only reads from it. When a message is received means that service must shutdown so you should perform a graceful shutdown of all the connections and if you need, report the in progress shutdown to other services in your infrastructure. |
| shutdownIsCompleteChannel  | ConnectionsHandler on writes on it. A message must be added only when ConnectionsHandler has completed its shutdown procedure. |


#### storage.DeviceConnectionsStorageInterface

> How individual IoT devices connections stats are stored offering some global stats as well.

Source [deviceConnectionsStorageInterface.go](storage/deviceConnectionsStorageInterface.go)

Check a ready to use thread safe in memory implementation at [storage.InMemoryDeviceConnectionsStorage](storage/inMemoryDeviceConnectionsStorage.go)


#### servers.CloudConnectorAPIInterface

> Each Cloud Connector instance has a build in REST API where users may fetch information regarding
connector status and current connections.

Source [cloudConnectorAPIInterface.go](servers/cloudConnectorAPIInterface.go)

However you may want to offer your own API, thats fine just implement this interface and pass the instance
to Cloud Connector.

For authentication methods, we provide some options compatible with 
the Default API. Check [authentication.go](servers/authentication.go).

For [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) settings use
[CrossOriginResourceSharing](servers/authentication.go) struct

We provide a default REST API for Cloud Connector check the [documentation here](/docs/default-cloud-connector-api.md).

# Links

Cloud connector uses these amazing projects:

- Gorilla Toolkit: https://www.gorillatoolkit.org/
- Logrus: https://github.com/sirupsen/logrus
- gotest.tools: https://github.com/gotestyourself/gotest.tools

# Licensing

Under MIT License, check [License file](./LICENSE)
