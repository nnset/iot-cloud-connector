package servers

import(
    "time"
    "runtime"
    "github.com/google/uuid"
    "github.com/sirupsen/logrus"
    "github.com/nnset/iot-cloud-connector/handlers"
)

/*
CloudServer Helps you controling the status of incoming connections to
your server and also when to start and shut it down.
*/
type CloudServer struct {
    id                      string
    address                 string
    port                    string
    network                 string
    startTime               int64
    log                     *logrus.Logger
    connectionsHandlerShutdown  *chan bool
    connectionsHandler      handlers.ConnectionsHandlerInterface
}

/*
NewCloudServer Creates a new instance of CloudServer
*/
func NewCloudServer(
    address, port, network string, 
    log *logrus.Logger, 
    connectionsHandlerShutdown *chan bool, 
    connectionsHandler handlers.ConnectionsHandlerInterface) *CloudServer {
    return &CloudServer {
        id: uuid.New().String(),
        address: address,
        port: port,
        network: network,
        log: log,
        connectionsHandlerShutdown: connectionsHandlerShutdown,
        connectionsHandler: connectionsHandler,
        startTime: time.Now().Unix(),
    }
}

/*
Start Starts the server on the given host and port
*/
func (server *CloudServer) Start() {
    server.log.Debugf("Starting CloudServer #%s at %s:%s", server.id, server.address, server.port)

    closingConnectionsIsComplete := make(chan bool, 1)

    go server.connectionsHandler.Listen(server.connectionsHandlerShutdown, &closingConnectionsIsComplete, server.log)
    
    <-*server.connectionsHandlerShutdown

    server.log.Info("CloudServer received shutdown signal, proceding to close open connections")

    *server.connectionsHandlerShutdown <- true

    // TODO Timeout does not work
    select {
        case <-closingConnectionsIsComplete:
            server.log.Debug("Connections handler successfully shutdown")
        case <-time.After(2 * time.Second):
            server.log.Error("Connections handler shutdown time out")
    }

    uptime := server.Uptime()

    server.log.Info("CloudServer stopped.")
    server.log.Info("  Total messages processed: ", server.connectionsHandler.TotalMessagesProcessed())
    server.log.Infof("  Uptime: %d seconds", uptime)
}

/*
Uptime how many seconds the server has been up
*/
func (server *CloudServer) Uptime() int64 {
    if server.startTime == 0 {
        return 0
    }

    return time.Now().Unix() - server.startTime
}

// total mega bytes of memory obtained from the OS.
func (server *CloudServer) systemMemory() uint {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    return uint(m.Sys / 1024 / 1024)
}

// Cumulative mega bytes allocated for heap objects.
func (server *CloudServer) totalAllocatedMemory() uint {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    return uint(m.TotalAlloc / 1024 / 1024)
}

func (server *CloudServer) totalGoRoutinesSpawned() int {
    return runtime.NumGoroutine()
}

// Servidor API REST para consultar las estadísticas

