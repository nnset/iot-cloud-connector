package servers

import (
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/nnset/iot-cloud-connector/connectionshandlers"
	"github.com/sirupsen/logrus"
)

/*
CloudConnector is the main process, once you Start() it, these processes are spawned :

	- An instance of connectionshandlers.ConnectionsHandlerInterface
		This is the instance you coded, there you handle your connections and business logic.
		Check connectionshandlers/sample*.go files for some examples.

	- An instance of StatusAPIInterface
		This opens a http/s server serving a JSON API where you can fetch the status of your connected
		devices and interact with them in case you need it (this is still a TODO).
*/
type CloudConnector struct {
	id                      string
	address                 string
	port                    string
	network                 string
	startTime               int64
	log                     *logrus.Logger
	serverShutdownWaitGroup sync.WaitGroup
	connectionsHandler      connectionshandlers.ConnectionsHandlerInterface
	statusAPI               StatusAPIInterface
	state                   CloudConnectorState
}

/*
CloudConnectorState Server's state
*/
type CloudConnectorState string

/*
	CloudConnectors go across some status :
	- CloudConnectorCreated
	- CloudConnectorStarted
	- CloudConnectorStopped
*/
const (
	CloudConnectorCreated CloudConnectorState = "created"
	CloudConnectorStarted CloudConnectorState = "started"
	CloudConnectorStopped CloudConnectorState = "stopped"
)

/*
NewCloudConnector Creates a new instance of CloudConnector
*/
func NewCloudConnector(
	address, port, network string,
	log *logrus.Logger,
	connectionsHandler connectionshandlers.ConnectionsHandlerInterface,
	statusAPI StatusAPIInterface) *CloudConnector {
	return &CloudConnector{
		id:                      uuid.New().String(),
		address:                 address,
		port:                    port,
		network:                 network,
		log:                     log,
		serverShutdownWaitGroup: sync.WaitGroup{},
		connectionsHandler:      connectionsHandler,
		statusAPI:               statusAPI,
		startTime:               time.Now().Unix(),
		state:                   CloudConnectorCreated,
	}
}

/*
Start Starts the server on the given host and port
*/
func (server *CloudConnector) Start() {
	server.log.Debugf("Starting CloudConnector #%s at %s:%s", server.id, server.address, server.port)

	connectionsHandlerShutdownIsComplete, shutdownConnectionsHandler := make(chan bool), make(chan bool)

	server.serverShutdownWaitGroup.Add(1)
	go server.waitForShutdownSignal()

	go server.connectionsHandler.Listen(&shutdownConnectionsHandler, &connectionsHandlerShutdownIsComplete, server.log)

	server.startAPI()

	server.state = CloudConnectorStarted

	server.serverShutdownWaitGroup.Wait()

	server.log.Info("CloudConnector received shutdown signal")

	server.shutdown(shutdownConnectionsHandler, connectionsHandlerShutdownIsComplete)
}

func (server *CloudConnector) waitForShutdownSignal() {
	operatingSystemSignal := make(chan os.Signal)

	signal.Notify(operatingSystemSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	go func() {
		sig := <-operatingSystemSignal
		server.log.Debugf("Signal received : %s", sig)
		server.log.Debug("Shutting down main")

		server.serverShutdownWaitGroup.Done()
	}()
}

func (server *CloudConnector) startAPI() {
	if server.statusAPI == nil {
		server.statusAPI = NewDefaultStatusAPI(server.address, server.port, server.log, server)
	}

	go server.statusAPI.Start()
}

func (server *CloudConnector) shutdown(shutdownConnectionsHandler, connectionsHandlerShutdownIsComplete chan bool) {
	shutdownConnectionsHandler <- true

	select {
	case <-connectionsHandlerShutdownIsComplete:
		server.log.Debug("Connections handler successfully shutdown")
	case <-time.After(8 * time.Second):
		server.log.Error("Connections handler shutdown time out")
	}

	server.statusAPI.Stop()

	server.state = CloudConnectorStopped

	server.log.Info("CloudConnector stopped.")
	server.log.Info("  Total incoming messages processed: ", server.connectionsHandler.Stats().TotalIncomingMessages())
	server.log.Info("  Total outgoing messages processed: ", server.connectionsHandler.Stats().TotalOutgoingMessages())
	server.log.Infof("  Uptime: %d seconds", server.Uptime(""))
}

/*
ID Server's uuid
*/
func (server *CloudConnector) ID() string {
	return server.id
}

/*
Uptime how many seconds the server has been up
*/
func (server *CloudConnector) Uptime(deviceID string) int64 {

	if deviceID == "" {
		if server.startTime == 0 {
			return 0
		}

		return time.Now().Unix() - server.startTime
	}

	connection, err := server.connectionsHandler.Stats().Get(deviceID)

	if err != nil {
		// TODO should we return error or just 0 ?
		return 0
	}

	uptime, err := connection.Uptime()

	if err != nil {
		// TODO should we return error or just 0 ?
		return 0
	}

	return uptime
}

/*
OpenConnections How many connections are currently open on this server
*/
func (server *CloudConnector) OpenConnections() uint {
	return server.connectionsHandler.Stats().OpenConnections()
}

/*
ConnectedDevices Which devices are currently connected
*/
func (server *CloudConnector) ConnectedDevices() []string {
	return server.connectionsHandler.Stats().ConnectedDevices()
}

// Devices []string `json:"devices"`

/*
IncomingMessages How many incoming messages were processed by a Device or globally if deviceID is empty
*/
func (server *CloudConnector) IncomingMessages(deviceID string) uint {
	if deviceID == "" {
		return server.connectionsHandler.Stats().TotalIncomingMessages()
	}

	return server.connectionsHandler.Stats().IncomingMessages(deviceID)
}

/*
OutgoingMessages How many messages this server sent to a Device or globally if deviceID is empty
*/
func (server *CloudConnector) OutgoingMessages(deviceID string) uint {
	if deviceID == "" {
		return server.connectionsHandler.Stats().TotalOutgoingMessages()
	}

	return server.connectionsHandler.Stats().OutgoingMessages(deviceID)
}

/*
SystemMemory total mega bytes of memory obtained from the OS.
*/
func (server *CloudConnector) SystemMemory() uint {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return uint(m.Sys / 1024 / 1024)
}

/*
AllocatedMemory mega bytes allocated for heap objects.
*/
func (server *CloudConnector) AllocatedMemory() uint {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return uint(m.Alloc / 1024 / 1024)
}

/*
HeapAllocatedMemory mega bytes of allocated heap objects.
*/
func (server *CloudConnector) HeapAllocatedMemory() uint {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return uint(m.HeapAlloc / 1024 / 1024)
}

/*
GoRoutinesSpawned How many Go routines are currently being executed
*/
func (server *CloudConnector) GoRoutinesSpawned() int {
	return runtime.NumGoroutine()
}

/*
State CloudConnector current state
*/
func (server *CloudConnector) State() CloudConnectorState {
	return server.state
}

/*
Kill Begins shutdown procedure
*/
func (server *CloudConnector) Kill() {
	server.serverShutdownWaitGroup.Done()
}
