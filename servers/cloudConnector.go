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
	"github.com/nnset/iot-cloud-connector/storage"
	"github.com/sirupsen/logrus"
)

/*
CloudConnector is the main process, once you Start() it, these processes are spawned :

	- An instance of connectionshandlers.ConnectionsHandlerInterface
		This is the instance you coded, there you handle your connections and business logic.
		Check connectionshandlers/sample*.go files for some examples.

	- An instance of CloudConnectorAPIInterface
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
	statusAPI               CloudConnectorAPIInterface
	state                   CloudConnectorState
	connectionsStats        storage.DeviceConnectionsStatsStorageInterface
	auth                    APIAuthMiddleWare
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
	log *logrus.Logger,
	connectionsHandler connectionshandlers.ConnectionsHandlerInterface,
	connectionsStats storage.DeviceConnectionsStatsStorageInterface,
	statusAPI CloudConnectorAPIInterface) *CloudConnector {
	return &CloudConnector{
		id:                      uuid.New().String(),
		log:                     log,
		serverShutdownWaitGroup: sync.WaitGroup{},
		connectionsHandler:      connectionsHandler,
		statusAPI:               statusAPI,
		startTime:               time.Now().Unix(),
		state:                   CloudConnectorCreated,
		connectionsStats:        connectionsStats,
	}
}

/*
Start Starts the server on the given host and port
*/
func (cc *CloudConnector) Start() {
	cc.log.Debugf("Starting CloudConnector #%s", cc.id)

	connectionsHandlerShutdownIsComplete, shutdownConnectionsHandler := make(chan bool), make(chan bool)

	cc.serverShutdownWaitGroup.Add(1)
	go cc.waitForShutdownSignal()

	go cc.connectionsHandler.Listen(&shutdownConnectionsHandler, &connectionsHandlerShutdownIsComplete, cc.connectionsStats, cc.log)

	if cc.statusAPI != nil {
		go cc.statusAPI.Start(cc)
	}

	cc.state = CloudConnectorStarted

	cc.serverShutdownWaitGroup.Wait()

	cc.log.Info("CloudConnector received shutdown signal")

	cc.shutdown(shutdownConnectionsHandler, connectionsHandlerShutdownIsComplete)
}

func (cc *CloudConnector) waitForShutdownSignal() {
	operatingSystemSignal := make(chan os.Signal)

	signal.Notify(operatingSystemSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	go func() {
		sig := <-operatingSystemSignal
		cc.log.Debugf("Signal received : %s", sig)
		cc.log.Debug("Shutting down main")

		cc.serverShutdownWaitGroup.Done()
	}()
}

func (cc *CloudConnector) shutdown(shutdownConnectionsHandler, connectionsHandlerShutdownIsComplete chan bool) {
	shutdownConnectionsHandler <- true

	select {
	case <-connectionsHandlerShutdownIsComplete:
		cc.log.Debug("Connections handler successfully shutdown")
	case <-time.After(8 * time.Second):
		cc.log.Error("Connections handler shutdown time out")
	}

	if cc.statusAPI != nil {
		cc.statusAPI.Stop()
	}

	cc.state = CloudConnectorStopped

	cc.log.Info("CloudConnector stopped.")
	cc.log.Info("  Total incoming messages processed: ", cc.connectionsStats.TotalIncomingMessages())
	cc.log.Info("  Total outgoing messages processed: ", cc.connectionsStats.TotalOutgoingMessages())
	cc.log.Infof("  Uptime: %d seconds", cc.Uptime(""))
}

/*
ID Server's uuid
*/
func (cc *CloudConnector) ID() string {
	return cc.id
}

/*
Uptime how many seconds the server has been up
*/
func (cc *CloudConnector) Uptime(deviceID string) int64 {

	if deviceID == "" {
		if cc.startTime == 0 {
			return 0
		}

		return time.Now().Unix() - cc.startTime
	}

	connection, err := cc.connectionsStats.Get(deviceID)

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
func (cc *CloudConnector) OpenConnections() uint {
	return cc.connectionsStats.OpenConnections()
}

/*
ConnectedDevices Which devices are currently connected
*/
func (cc *CloudConnector) ConnectedDevices() []string {
	return cc.connectionsStats.ConnectedDevices()
}

/*
IncomingMessages How many incoming messages were processed by a Device or globally if deviceID is empty
*/
func (cc *CloudConnector) IncomingMessages(deviceID string) uint {
	if deviceID == "" {
		return cc.connectionsStats.TotalIncomingMessages()
	}

	return cc.connectionsStats.IncomingMessages(deviceID)
}

/*
OutgoingMessages How many messages this server sent to a Device or globally if deviceID is empty
*/
func (cc *CloudConnector) OutgoingMessages(deviceID string) uint {
	if deviceID == "" {
		return cc.connectionsStats.TotalOutgoingMessages()
	}

	return cc.connectionsStats.OutgoingMessages(deviceID)
}

/*
SendACommand TODO
*/
func (cc *CloudConnector) SendCommand(payload, deviceID string) (string, int, error) {
	return cc.connectionsHandler.SendCommand(payload, deviceID)
}

/*
SendQuery TODO
*/
func (cc *CloudConnector) SendQuery(payload, deviceID string) (string, int, error) {
	return cc.connectionsHandler.SendQuery(payload, deviceID)
}

/*
SystemMemory total mega bytes of memory obtained from the OS.
*/
func (cc *CloudConnector) SystemMemory() uint {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return uint(m.Sys / 1024 / 1024)
}

/*
AllocatedMemory mega bytes allocated for heap objects.
*/
func (cc *CloudConnector) AllocatedMemory() uint {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return uint(m.Alloc / 1024 / 1024)
}

/*
HeapAllocatedMemory mega bytes of allocated heap objects.
*/
func (cc *CloudConnector) HeapAllocatedMemory() uint {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return uint(m.HeapAlloc / 1024 / 1024)
}

/*
GoRoutinesSpawned How many Go routines are currently being executed
*/
func (cc *CloudConnector) GoRoutinesSpawned() int {
	return runtime.NumGoroutine()
}

/*
State CloudConnector current state
*/
func (cc *CloudConnector) State() CloudConnectorState {
	return cc.state
}

/*
Kill Begins shutdown procedure
*/
func (cc *CloudConnector) Kill() {
	cc.serverShutdownWaitGroup.Done()
}

func (cc *CloudConnector) QueriesWaiting() uint {
	return cc.connectionsHandler.QueriesWaiting()
}

func (cc *CloudConnector) CommandsWaiting() uint {
	return cc.connectionsHandler.CommandsWaiting()
}
