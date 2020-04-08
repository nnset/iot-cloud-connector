package servers

import (
	"runtime"
	"time"

	"github.com/google/uuid"
	"github.com/nnset/iot-cloud-connector/connectionshandlers"
	"github.com/sirupsen/logrus"
)

/*
CloudServer Helps you controling the status of incoming connections to
your server and also when to start and shut it down.
*/
type CloudServer struct {
	id                         string
	address                    string
	port                       string
	network                    string
	startTime                  int64
	log                        *logrus.Logger
	connectionsHandlerShutdown *chan bool
	connectionsHandler         connectionshandlers.ConnectionsHandlerInterface
	api                        *CloudServerAPI
	state                      CloudServerState
}

/*
CloudServerState used to know on which state is the server
*/
type CloudServerState int

/*
	CloudServers go across some status :
	- CloudServerCreated
	- CloudServerStarted
	- CloudServerStopped
*/
const (
	CloudServerCreated CloudServerState = 1
	CloudServerStarted CloudServerState = 2
	CloudServerStopped CloudServerState = 3
)

/*
NewCloudServer Creates a new instance of CloudServer
*/
func NewCloudServer(
	address, port, network string,
	log *logrus.Logger,
	connectionsHandlerShutdown *chan bool,
	connectionsHandler connectionshandlers.ConnectionsHandlerInterface) *CloudServer {
	return &CloudServer{
		id:                         uuid.New().String(),
		address:                    address,
		port:                       port,
		network:                    network,
		log:                        log,
		connectionsHandlerShutdown: connectionsHandlerShutdown,
		connectionsHandler:         connectionsHandler,
		startTime:                  time.Now().Unix(),
		state:                      CloudServerCreated,
	}
}

/*
Start Starts the server on the given host and port
*/
func (server *CloudServer) Start() {
	server.log.Debugf("Starting CloudServer #%s at %s:%s", server.id, server.address, server.port)

	if server.api == nil {
		server.log.Debugf("Starting CloudServerAPI")
		server.api = NewCloudServerAPI(server.address, server.port, server.log, server)
		go server.api.Start()
	}

	server.log.Debugf("Starting CloudServer connections handler")

	closingConnectionsIsComplete := make(chan bool)

	go server.connectionsHandler.Listen(server.connectionsHandlerShutdown, &closingConnectionsIsComplete, server.log)

	server.state = CloudServerStarted

	<-*server.connectionsHandlerShutdown

	server.log.Info("CloudServer received shutdown signal")

	*server.connectionsHandlerShutdown <- true

	select {
	case <-closingConnectionsIsComplete:
		server.log.Debug("Connections handler successfully shutdown")
	case <-time.After(8 * time.Second):
		server.log.Error("Connections handler shutdown time out")
	}

	server.log.Debug("Shutting down CloudServerAPI")
	server.api.Stop()

	uptime := server.Uptime()
	server.state = CloudServerStopped

	server.log.Info("CloudServer stopped.")
	server.log.Info("  Total incoming messages processed: ", server.connectionsHandler.Stats().TotalIncomingMessages())
	server.log.Infof("  Uptime: %d seconds", uptime)
}

/*
ID Server's uuid
*/
func (server *CloudServer) ID() string {
	return server.id
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

/*
OpenConnections How many connections are currently open against this server
*/
func (server *CloudServer) OpenConnections() uint {
	return server.connectionsHandler.Stats().OpenConnections()
}

/*
IncomingMessages How many incoming messages were processed
*/
func (server *CloudServer) IncomingMessages() uint {
	return server.connectionsHandler.Stats().TotalIncomingMessages()
}

/*
OutgoingMessages How many messages this server sent to the connected clients
*/
func (server *CloudServer) OutgoingMessages() uint {
	return server.connectionsHandler.Stats().TotalOutgoingMessages()
}

/*
SystemMemory total mega bytes of memory obtained from the OS.
*/
func (server *CloudServer) SystemMemory() uint {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return uint(m.Sys / 1024 / 1024)
}

/*
AllocatedMemory mega bytes allocated for heap objects.
*/
func (server *CloudServer) AllocatedMemory() uint {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return uint(m.Alloc / 1024 / 1024)
}

/*
HeapAllocatedMemory mega bytes of allocated heap objects.
*/
func (server *CloudServer) HeapAllocatedMemory() uint {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return uint(m.HeapAlloc / 1024 / 1024)
}

/*
GoRoutinesSpawned How many Go routines are currently being executed
*/
func (server *CloudServer) GoRoutinesSpawned() int {
	return runtime.NumGoroutine()
}

/*
State CloudServer current state
*/
func (server *CloudServer) State() CloudServerState {
	return server.state
}
