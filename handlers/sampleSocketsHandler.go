package handlers

import (
	"encoding/json"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nnset/iot-cloud-connector/storage"
	"github.com/sirupsen/logrus"
)

/*
SampleSocketsHandler Simple example of a ConnectionHandler that handles
incoming permanent socket connections expecting messages using a JSON format
with fields defined in payload struct.
For each message received this handler will reply with a 'OK' also using payload
struct.
*/
type SampleSocketsHandler struct {
	id        string
	address   string
	port      string
	network   string
	log       *logrus.Logger
	startTime int64

	activeConnections storage.DeviceConnectionsStorageInterface
	dataMutex         *sync.Mutex
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
		activeConnections: storage.NewInMemoryDeviceConnectionsStorage(),
		dataMutex:         &sync.Mutex{},
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

	for {
		conn, err := portListener.Accept()

		if err != nil {
			log.Fatalln(err)
			handler.log.Errorf("Error accepting connection: %s", err)
			continue // TODO Continue or break
		}

		go handler.handleConnection(conn)

		if <-*shutdownChannel {
			handler.log.Debugf("SampleSocketsHandler received shutdown signal")
			break
		}
	}

	handler.log.Debugf("SampleSocketsHandler is closing connections")

	// TODO

	*shutdownIsCompleteChannel <- true

	handler.log.Debugf("SampleSocketsHandler all connections closed")

	return nil
}

func (handler *SampleSocketsHandler) handleConnection(connection net.Conn) {
	handler.log.Debugf("New connection from %s", connection.RemoteAddr().String())
	defer connection.Close()

	connectionID := uuid.New().String()

	handler.activeConnections.Add(
		connectionID,
		"SampleSockets",
		"SampleSockets",
		"SampleSockets",
		connection.RemoteAddr().String(),
	)

	for {
		var requestPayload payload
		decoder := json.NewDecoder(connection)
		err := decoder.Decode(&requestPayload)

		if err != nil {
			handler.log.Debugf("Unable to decode payload: %s", err)
			break
		}

		handler.log.Debugf("Mesage received from %s: %s", requestPayload.Sender, requestPayload.Body)
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
IncomingMessagesProcessed How many messages from all connections have been processed
*/
func (handler *SampleSocketsHandler) IncomingMessagesProcessed() uint {
	return handler.activeConnections.TotalIncomingMessages()
}

/*
OpenConnections How open connections are currently registered
*/
func (handler *SampleSocketsHandler) OpenConnections() uint {
	return handler.activeConnections.OpenConnections()
}
