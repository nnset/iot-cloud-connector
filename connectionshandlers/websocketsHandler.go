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
WebSocketsHandler
*/
type WebSocketsHandler struct {
	id                         string
	address                    string
	port                       string
	network                    string
	keyFile                    string
	certFile                   string
	log                        *logrus.Logger
	startTime                  int64
	shutdownInProgress         bool
	activeConnections          storage.DeviceConnectionsStorageInterface
	openWebsockets             map[string]*websocket.Conn
	queriesToDevicesWaiting    map[string]chan deviceMessage
	commandsToDevicesWaiting   map[string]chan deviceMessage
	dataMutex                  *sync.Mutex
	shutdownOutgoingDeliveries chan bool
	handleIncomingMessage      func(deviceID string, messageType int, p []byte) error
	authenticateConnection     func(r *http.Request) error
}

/*
NewWebSocketsHandler TODO
*/
func NewWebSocketsHandler(
	address, port, network, keyFile, certFile string,
	handleIncomingMessage func(deviceID string, messageType int, p []byte) error,
	authenticateConnection func(r *http.Request) error,
) *WebSocketsHandler {

	return &WebSocketsHandler{
		id:                         uuid.New().String(),
		address:                    address,
		port:                       port,
		network:                    network,
		startTime:                  time.Now().Unix(),
		openWebsockets:             make(map[string]*websocket.Conn),
		queriesToDevicesWaiting:    make(map[string]chan deviceMessage),
		commandsToDevicesWaiting:   make(map[string]chan deviceMessage),
		dataMutex:                  &sync.Mutex{},
		certFile:                   certFile,
		keyFile:                    keyFile,
		shutdownInProgress:         false,
		shutdownOutgoingDeliveries: make(chan bool),
		handleIncomingMessage:      handleIncomingMessage,
		authenticateConnection:     authenticateConnection,
	}
}

/*
Listen TODO
*/
func (handler *WebSocketsHandler) Start(
	shutdownChannel, shutdownIsCompleteChannel *chan bool,
	activeConnections storage.DeviceConnectionsStorageInterface,
	log *logrus.Logger,
) error {

	handler.log = log
	handler.activeConnections = activeConnections

	go handler.gracefullShutdown(shutdownChannel, shutdownIsCompleteChannel)

	http.HandleFunc("/connect", handler.handleConnection)

	// TODO Gracefully shutdown http.server
	if handler.keyFile != "" {
		handler.log.Debugf("Serving websockets via wss (TLS ON) at %s:%s", handler.address, handler.port)
		handler.log.Debugf("  Connect endpoint wss://%s:%s/connect", handler.address, handler.port)

		err := http.ListenAndServeTLS(handler.address+":"+handler.port, handler.certFile, handler.keyFile, nil)

		if err != nil {
			handler.log.Fatal("ListenAndServe failed", err)

			return err
		}
	} else {
		handler.log.Debugf("Serving websockets via ws (TLS OFF) at %s:%s", handler.address, handler.port)
		handler.log.Debugf("  Connect endpoint ws://%s:%s/connect", handler.address, handler.port)

		err := http.ListenAndServe(handler.address+":"+handler.port, nil)

		if err != nil {
			handler.log.Fatal("ListenAndServe failed ", err)

			return err
		}
	}

	return nil
}

func (handler *WebSocketsHandler) gracefullShutdown(shutdownChannel, shutdownIsCompleteChannel *chan bool) {
	<-*shutdownChannel
	handler.log.Debugf("SampleWebsocketsHandler shutdown signal received. Proceeding.")

	handler.shutdownInProgress = true

	handler.closeAllConnections()

	handler.log.Debugf("SampleWebsocketsHandler all shutdown steps have finished")

	*shutdownIsCompleteChannel <- true
}

func (handler *WebSocketsHandler) closeAllConnections() error {
	handler.dataMutex.Lock()
	defer handler.dataMutex.Unlock()

	handler.log.Debugf("SampleWebsocketsHandler closing open connections")

	for deviceID, wsConnection := range handler.openWebsockets {
		handler.log.Debugf("  Closing connection to device #%s (%s)", deviceID, wsConnection.RemoteAddr().String())

		wsConnection.Close()
		delete(handler.openWebsockets, deviceID)
	}

	handler.log.Debugf("  All connections closed")

	return nil
}

func (handler *WebSocketsHandler) handleConnection(w http.ResponseWriter, r *http.Request) {

	authError := handler.authenticateConnection(r)

	if authError != nil {
		handler.log.Debugf("Unauthorized connection from %s", r.RemoteAddr)
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	wsConn, err := handler.upgradeConnectionToWebSocket(w, r)

	if err != nil {
		handler.log.Errorf("Unable to upgrade to websocket %s", err)

		return
	}

	defer wsConn.Close()

	handler.log.Debugf("Websocket from %s accepted", r.RemoteAddr)

	deviceID := r.Header.Get("Device-id")
	deviceName := r.Header.Get("Device-Name")
	deviceType := r.Header.Get("Device-Type")
	userAgent := r.Header.Get("User-Agent")
	connectionID := uuid.New().String()

	handler.saveConnection(deviceID, wsConn)
	defer handler.deleteConnection(deviceID)

	handler.activeConnections.Add(connectionID, deviceID, deviceName, deviceType, userAgent, r.RemoteAddr)
	defer handler.activeConnections.Delete(deviceID)

	handler.handleIncomingMessages(deviceID, wsConn)
}

func (handler *WebSocketsHandler) upgradeConnectionToWebSocket(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
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

func (handler *WebSocketsHandler) saveConnection(deviceID string, ws *websocket.Conn) error {
	handler.dataMutex.Lock()
	defer handler.dataMutex.Unlock()

	_, alreadyConnected := handler.openWebsockets[deviceID]

	if alreadyConnected {
		return errors.New("Device already connected")
	}

	handler.openWebsockets[deviceID] = ws

	return nil
}

func (handler *WebSocketsHandler) deleteConnection(deviceID string) {
	handler.dataMutex.Lock()
	defer handler.dataMutex.Unlock()

	delete(handler.openWebsockets, deviceID)
}

func (handler *WebSocketsHandler) handleIncomingMessages(deviceID string, wsConn *websocket.Conn) {
	for {

		messageType, message, err := wsConn.ReadMessage()

		if err != nil {
			if handler.shutdownInProgress == false {
				handler.log.Error("Error reading message ", err)
			}

			return
		}

		if handler.messageRespondsToAQueryOrACommand(message) {
			err = handler.handleResponseToCommandOrQuery(message)
		} else {
			err = handler.handleIncomingMessage(deviceID, messageType, message)
		}

		handler.activeConnections.MessageWasReceived(deviceID)
	}
}

func (handler *WebSocketsHandler) handleResponseToCommandOrQuery(message []byte) error {
	var m deviceMessage
	err := json.Unmarshal([]byte(message), &m)

	if err != nil {
		return err
	}

	handler.dataMutex.Lock()
	defer handler.dataMutex.Unlock()

	if handler.isAQueryWaitingForMessage(m.ID) {
		channel, exists := handler.queriesToDevicesWaiting[m.ID]

		if exists {
			channel <- m
		}
	} else if handler.isACommandWaitingForMessage(m.ID) {
		channel, exists := handler.commandsToDevicesWaiting[m.ID]

		if exists {
			channel <- m
		}
	}

	return nil
}

func (handler *WebSocketsHandler) messageRespondsToAQueryOrACommand(message []byte) bool {

	var m deviceMessage
	err := json.Unmarshal([]byte(message), &m)

	return err == nil && m.ID != ""
}

func (handler *WebSocketsHandler) isAQueryWaitingForMessage(messageID string) bool {
	_, exists := handler.queriesToDevicesWaiting[messageID]

	return exists
}

func (handler *WebSocketsHandler) isACommandWaitingForMessage(messageID string) bool {
	_, exists := handler.commandsToDevicesWaiting[messageID]

	return exists
}

/*
SendCommand TODO
*/
func (handler *WebSocketsHandler) SendCommand(payload, deviceID string) (string, int, error) {
	handler.dataMutex.Lock()
	_, isConnected := handler.openWebsockets[deviceID]
	handler.dataMutex.Unlock()

	if !isConnected {
		return "", http.StatusNotFound, errors.New("Device not connected")
	}

	select {
	case r := <-handler.sendMessageToDeviceSynchronously(payload, deviceID, handler.commandsToDevicesWaiting):
		return fmt.Sprint(r), http.StatusOK, nil
	case <-time.After(8 * time.Second):
		return "", http.StatusRequestTimeout, errors.New("Device command timeout")
	}
}

/*
SendQuery TODO
*/
func (handler *WebSocketsHandler) SendQuery(payload, deviceID string) (string, int, error) {
	handler.dataMutex.Lock()
	_, isConnected := handler.openWebsockets[deviceID]
	handler.dataMutex.Unlock()

	if !isConnected {
		return "", http.StatusNotFound, errors.New("Device not connected")
	}

	select {
	case r := <-handler.sendMessageToDeviceSynchronously(payload, deviceID, handler.queriesToDevicesWaiting):
		return fmt.Sprint(r), http.StatusOK, nil
	case <-time.After(8 * time.Second):
		return "", http.StatusRequestTimeout, errors.New("Device query timeout")
	}
}

func (handler *WebSocketsHandler) sendMessageToDeviceSynchronously(payload, deviceID string, queue map[string]chan deviceMessage) <-chan string {
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
		connection, _ := handler.openWebsockets[deviceID]
		err = connection.WriteMessage(1, marshalledMessage)
		handler.dataMutex.Unlock()

		if err != nil {
			r <- fmt.Sprintf("Error sending message to deviceID #%s. %s", deviceID, err)
			return
		}

		deviceQueryResponse := <-asyncResponseWaitChannel

		handler.activeConnections.MessageWasSent(deviceID)

		r <- deviceQueryResponse.Payload
	}()

	return r
}

func (handler *WebSocketsHandler) QueriesWaiting() uint {
	return uint(len(handler.queriesToDevicesWaiting))
}

func (handler *WebSocketsHandler) CommandsWaiting() uint {
	return uint(len(handler.commandsToDevicesWaiting))
}
