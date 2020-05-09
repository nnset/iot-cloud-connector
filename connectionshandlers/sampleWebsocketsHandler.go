package connectionshandlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/nnset/iot-cloud-connector/storage"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

/*
SampleWebSocketsHandler
*/
type SampleWebSocketsHandler struct {
	id                         string
	address                    string
	port                       string
	network                    string
	keyFile                    string
	certFile                   string
	log                        *logrus.Logger
	startTime                  int64
	shutdownInProgress         bool
	connectionsStats           storage.DeviceConnectionsStatsStorageInterface
	connections                map[string]*websocket.Conn
	queriesToDevicesWaiting    map[string]chan deviceMessage
	commandsToDevicesWaiting   map[string]chan deviceMessage
	dataMutex                  *sync.Mutex
	shutdownOutgoingDeliveries chan bool
}

type deviceMessage struct {
	ID      string `json:"id"`
	Payload string `json:"payload"`
	Time    int64  `json:"timestamp"`
}

/*
NewSampleWebSocketsHandler TODO
*/
func NewSampleWebSocketsHandler(address, port, network, keyFile, certFile string) *SampleWebSocketsHandler {
	return &SampleWebSocketsHandler{
		id:                         uuid.New().String(),
		address:                    address,
		port:                       port,
		network:                    network,
		startTime:                  time.Now().Unix(),
		connections:                make(map[string]*websocket.Conn),
		queriesToDevicesWaiting:    make(map[string]chan deviceMessage),
		commandsToDevicesWaiting:   make(map[string]chan deviceMessage),
		dataMutex:                  &sync.Mutex{},
		certFile:                   certFile,
		keyFile:                    keyFile,
		shutdownInProgress:         false,
		shutdownOutgoingDeliveries: make(chan bool),
	}
}

/*
Listen TODO
*/
func (handler *SampleWebSocketsHandler) Listen(shutdownChannel, shutdownIsCompleteChannel *chan bool, connectionsStats storage.DeviceConnectionsStatsStorageInterface, log *logrus.Logger) error {
	handler.log = log
	handler.connectionsStats = connectionsStats

	go handler.gracefullShutdown(shutdownChannel, shutdownIsCompleteChannel)

	http.HandleFunc("/connect", handler.handleConnection)

	// TODO Gracefully shutdown http.server
	if handler.keyFile != "" {
		handler.log.Debug("Serving websockets via wss (TLS ON) at " + handler.address + ":" + handler.port)
		handler.log.Debugf("  Connect endpoint wss://%s:%s/connect", handler.address, handler.port)

		err := http.ListenAndServeTLS(handler.address+":"+handler.port, handler.certFile, handler.keyFile, nil)

		if err != nil {
			handler.log.Fatal("ListenAndServe failed ", err)
			return err
		}
	} else {
		handler.log.Debug("Serving websockets via ws (TLS OFF) at " + handler.address + ":" + handler.port)
		handler.log.Debugf("  Connect endpoint ws://%s:%s/connect", handler.address, handler.port)

		err := http.ListenAndServe(handler.address+":"+handler.port, nil)

		if err != nil {
			handler.log.Fatal("ListenAndServe failed ", err)
			return err
		}
	}

	return nil
}

func (handler *SampleWebSocketsHandler) gracefullShutdown(shutdownChannel, shutdownIsCompleteChannel *chan bool) {
	<-*shutdownChannel
	handler.log.Debugf("SampleWebsocketsHandler shutdown signal received. Proceeding.")

	handler.shutdownInProgress = true

	handler.closeAllConnections()

	handler.log.Debugf("SampleWebsocketsHandler all shutdown steps have finished")

	*shutdownIsCompleteChannel <- true
}

func (handler *SampleWebSocketsHandler) closeAllConnections() error {
	handler.dataMutex.Lock()
	defer handler.dataMutex.Unlock()

	handler.log.Debugf("SampleWebsocketsHandler closing open connections")

	for deviceID, wsConnection := range handler.connections {
		handler.log.Debugf("  Closing connection to device #%s (%s)", deviceID, wsConnection.RemoteAddr().String())

		wsConnection.Close()
		delete(handler.connections, deviceID)
	}

	handler.log.Debugf("  All connections closed")

	return nil
}

func (handler *SampleWebSocketsHandler) handleConnection(w http.ResponseWriter, r *http.Request) {
	wsConn, err := handler.upgradeConnectionToWebSocket(w, r)

	if err != nil {
		handler.log.Errorf("Unable to upgrade to websocket %s", err)
		return
	}

	defer wsConn.Close()

	handler.log.Debugf("Websocket from %s accepted", r.RemoteAddr)

	deviceID := r.Header.Get("device_id")
	deviceType := r.Header.Get("device_type")
	userAgent := r.Header.Get("User-Agent")
	connectionID := uuid.New().String()

	handler.saveConnection(deviceID, wsConn)
	defer handler.deleteConnection(deviceID)

	handler.connectionsStats.Add(connectionID, deviceID, deviceType, userAgent, r.RemoteAddr)
	defer handler.connectionsStats.Delete(deviceID)

	handler.handleIncomingMessages(deviceID, wsConn)
}

func (handler *SampleWebSocketsHandler) upgradeConnectionToWebSocket(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	wsConn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		return nil, err
	}

	return wsConn, nil
}

func (handler *SampleWebSocketsHandler) saveConnection(deviceID string, ws *websocket.Conn) error {
	handler.dataMutex.Lock()
	defer handler.dataMutex.Unlock()

	_, alreadyConnected := handler.connections[deviceID]

	if alreadyConnected {
		return errors.New("Device already connected")
	}

	handler.connections[deviceID] = ws

	return nil
}

func (handler *SampleWebSocketsHandler) deleteConnection(deviceID string) {
	handler.dataMutex.Lock()
	defer handler.dataMutex.Unlock()

	delete(handler.connections, deviceID)
}

func (handler *SampleWebSocketsHandler) handleIncomingMessages(deviceID string, wsConn *websocket.Conn) {
	for {
		messageType, message, err := wsConn.ReadMessage()

		if err != nil {
			if handler.shutdownInProgress == false {
				handler.log.Error("Error reading message ", err)
			}

			return
		}

		handler.connectionsStats.IncomingMessageReceived(deviceID)

		var m deviceMessage
		err = json.Unmarshal([]byte(message), &m)

		if err == nil && m.ID != "" {
			handler.dataMutex.Lock()

			channel, exists := handler.queriesToDevicesWaiting[m.ID]

			if exists {
				channel <- m
			} else {
				channel, exists = handler.commandsToDevicesWaiting[m.ID]

				if exists {
					channel <- m
				}
			}

			handler.dataMutex.Unlock()
		} else {
			handler.log.Debugf("recv: Type: %d, %s", messageType, message)
		}
	}
}

/*
SendCommand TODO
*/
func (handler *SampleWebSocketsHandler) SendCommand(payload, deviceID string) (string, int, error) {
	handler.dataMutex.Lock()
	_, isConnected := handler.connections[deviceID]
	handler.dataMutex.Unlock()

	if !isConnected {
		return "", http.StatusNotFound, errors.New("Device not connected")
	}

	select {
	case r := <-handler.synchMessageToDevice(payload, deviceID, handler.commandsToDevicesWaiting):
		return fmt.Sprint(r), http.StatusOK, nil
	case <-time.After(8 * time.Second):
		return "", http.StatusRequestTimeout, errors.New("Device command timeout")
	}
}

/*
SendQuery TODO
*/
func (handler *SampleWebSocketsHandler) SendQuery(payload, deviceID string) (string, int, error) {
	handler.dataMutex.Lock()
	_, isConnected := handler.connections[deviceID]
	handler.dataMutex.Unlock()

	if !isConnected {
		return "", http.StatusNotFound, errors.New("Device not connected")
	}

	select {
	case r := <-handler.synchMessageToDevice(payload, deviceID, handler.queriesToDevicesWaiting):
		return fmt.Sprint(r), http.StatusOK, nil
	case <-time.After(8 * time.Second):
		return "", http.StatusRequestTimeout, errors.New("Device query timeout")
	}
}

func (handler *SampleWebSocketsHandler) synchMessageToDevice(payload, deviceID string, queue map[string]chan deviceMessage) <-chan string {
	r := make(chan string)

	go func() {
		defer close(r)

		messageID := uuid.New().String()
		asyncResponseWaitChannel := make(chan deviceMessage)
		message := deviceMessage{messageID, payload, time.Now().Unix()}
		marshalledMessage, err := json.Marshal(message) // TODO error checks on json marshalling

		if err != nil {
			r <- fmt.Sprintf("Error unable to json encode message to device: %s", err)
			return
		}

		handler.dataMutex.Lock()
		queue[messageID] = asyncResponseWaitChannel
		handler.dataMutex.Unlock()

		defer close(asyncResponseWaitChannel)
		defer func() {
			handler.dataMutex.Lock()
			delete(queue, messageID)
			handler.dataMutex.Unlock()
		}()

		handler.dataMutex.Lock()
		connection, _ := handler.connections[deviceID]
		err = connection.WriteMessage(1, marshalledMessage)
		handler.dataMutex.Unlock()

		if err != nil {
			r <- fmt.Sprintf("Error sending message to deviceID #%s. %s", deviceID, err)
			return
		}

		deviceQueryResponse := <-asyncResponseWaitChannel

		handler.connectionsStats.OutgoingMessageSent(deviceID)

		r <- deviceQueryResponse.Payload
	}()

	return r
}

func (handler *SampleWebSocketsHandler) QueriesWaiting() uint {
	return uint(len(handler.queriesToDevicesWaiting))
}

func (handler *SampleWebSocketsHandler) CommandsWaiting() uint {
	return uint(len(handler.commandsToDevicesWaiting))
}
