package connectionshandlers

import (
	"encoding/json"
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nnset/iot-cloud-connector/storage"
	"github.com/sirupsen/logrus"
)

/*
SampleSocketsHandler Simple example of a ConnectionHandler that handles
incoming permanent socket connections. Expected messages should use JSON format
with fields defined in payload struct.
For each message received, this handler will reply with a 'OK' also using payload
struct.
*/
type SampleSocketsHandler struct {
	id        string
	address   string
	port      string
	network   string
	log       *logrus.Logger
	startTime int64

	activeConnections storage.DeviceConnectionsStatsStorageInterface
	dataMutex         *sync.Mutex
	connections       map[string]net.Conn
}

type payload struct {
	Sender string
	Body   string
	Time   int64
}

/*
NewSampleSocketsHandler Creates a new instance of SampleSocketsHandler
*/
func NewSampleSocketsHandler(address, port, network string) *SampleSocketsHandler {
	return &SampleSocketsHandler{
		id:                uuid.New().String(),
		address:           address,
		port:              port,
		network:           network,
		startTime:         time.Now().Unix(),
		activeConnections: storage.NewInMemoryDeviceConnectionsStatsStorage(),
		dataMutex:         &sync.Mutex{},
		connections:       make(map[string]net.Conn),
	}
}

/*
Listen Starts listening to sockets connections using Go net.Listen
*/
func (handler *SampleSocketsHandler) Listen(shutdownChannel, shutdownIsCompleteChannel *chan bool, log *logrus.Logger) error {
	handler.log = log

	handler.log.Debugf("SampleSocketsHandler listening to %s:%s", handler.address, handler.port)

	portListener, err := net.Listen(handler.network, handler.address+":"+handler.port)

	if err != nil {
		return err
	}

	defer portListener.Close()

	go handler.acceptConnections(portListener)

	<-*shutdownChannel

	handler.log.Debugf("SampleSocketsHandler received shutdown signal")
	handler.log.Debugf("SampleSocketsHandler is closing connections")

	for id, connection := range handler.connections {
		handler.log.Debugf("Closing connection #%s", id)
		err = connection.Close()

		if err != nil {
			handler.log.Debugf("Unable to close connection %s", err)
		}
	}

	*shutdownIsCompleteChannel <- true

	handler.log.Debugf("SampleSocketsHandler all connections closed")

	return nil
}

func (handler *SampleSocketsHandler) acceptConnections(portListener net.Listener) {
	for {
		conn, err := portListener.Accept()

		if err != nil {
			log.Fatalln(err)
			handler.log.Errorf("Error accepting connection: %s", err)

			return
		}

		go handler.handleConnection(conn)
	}
}

func (handler *SampleSocketsHandler) handleConnection(connection net.Conn) {
	connectionID := uuid.New().String()
	handler.log.Debugf("New connection %s from %s", connectionID, connection.RemoteAddr().String())
	defer connection.Close()

	handler.activeConnections.Add(
		connectionID,
		"SampleSockets",
		"SampleSockets",
		"SampleSockets",
		connection.RemoteAddr().String(),
	)

	handler.dataMutex.Lock()
	handler.connections[connectionID] = connection
	handler.dataMutex.Unlock()

	for {
		var requestPayload payload
		decoder := json.NewDecoder(connection)
		err := decoder.Decode(&requestPayload)

		if err != nil {
			handler.log.Debugf("Unable to process message: %s", err)
			break
		}

		handler.log.Debugf("Connection #%s Mesage received from '%s': %s", connectionID, requestPayload.Sender, requestPayload.Body)
		handler.activeConnections.IncomingMessageReceived(connectionID) // We ignore errors on this example code

		responsePayload := payload{
			Sender: "Sample Sockets Handler",
			Body:   "OK",
			Time:   time.Now().Unix(),
		}

		encoder := json.NewEncoder(connection)
		err = encoder.Encode(&responsePayload)

		if err != nil {
			handler.log.Debugf("Unable to encode response: %s", err)
		}

		handler.activeConnections.OutgoingMessageSent(connectionID)
	}

	handler.log.Debugf("Closing connection from %s", connection.RemoteAddr().String())

	handler.activeConnections.Delete(connectionID) // We ignore errors on this example code
}

/*
Stats TODO
*/
func (handler *SampleSocketsHandler) Stats() storage.DeviceConnectionsStatsStorageInterface {
	return handler.activeConnections
}
