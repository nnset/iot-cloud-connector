package connectionshandlers

import (
	"errors"
	"math/rand"
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
	dataMutex                  *sync.Mutex
	shutdownOutgoingDeliveries chan bool
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
		connectionsStats:           storage.NewInMemoryDeviceConnectionsStatsStorage(),
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
func (handler *SampleWebSocketsHandler) Listen(shutdownChannel, shutdownIsCompleteChannel *chan bool, log *logrus.Logger) error {
	handler.log = log
	outgoingMessagesDeliverTicker := time.NewTicker(3 * time.Second)

	go handler.gracefullShutdown(shutdownChannel, shutdownIsCompleteChannel, outgoingMessagesDeliverTicker)
	go handler.outgoingMessagesDeliver(outgoingMessagesDeliverTicker)

	http.HandleFunc("/connect", handler.handleConnection)

	// TODO Gracefully shutdown http.server
	if handler.keyFile != "" {
		handler.log.Debug("Serving websockets via wss (TLS ON) at " + handler.address + ":" + handler.port)

		err := http.ListenAndServeTLS(handler.address+":"+handler.port, handler.certFile, handler.keyFile, nil)

		if err != nil {
			handler.log.Fatal("ListenAndServe failed ", err)
			return err
		}
	} else {
		handler.log.Debug("Serving websockets via ws (TLS OFF) at " + handler.address + ":" + handler.port)

		err := http.ListenAndServe(handler.address+":"+handler.port, nil)

		if err != nil {
			handler.log.Fatal("ListenAndServe failed ", err)
			return err
		}
	}

	return nil
}

func (handler *SampleWebSocketsHandler) gracefullShutdown(shutdownChannel, shutdownIsCompleteChannel *chan bool, outgoingDeliveriesTicker *time.Ticker) {
	handler.log.Debugf("SampleWebsocketsHandler OK waiting for shutdown signal")

	<-*shutdownChannel

	handler.dataMutex.Lock()
	handler.shutdownInProgress = true
	handler.dataMutex.Unlock()

	outgoingDeliveriesTicker.Stop()
	handler.closeAllConnections()

	handler.log.Debugf("SampleWebsocketsHandler all shutdown steps have finished")

	*shutdownIsCompleteChannel <- true
}

func (handler *SampleWebSocketsHandler) closeAllConnections() error {
	handler.dataMutex.Lock()
	defer handler.dataMutex.Unlock()

	handler.log.Debugf("SampleWebsocketsHandler closing open connections")

	for connectionID, wsConnection := range handler.connections {
		handler.log.Debugf("  Closing connection #%s (%s)", connectionID, wsConnection.RemoteAddr().String())

		wsConnection.Close()
		delete(handler.connections, connectionID)
	}

	handler.log.Debugf("  All connections closed")

	return nil
}

func (handler *SampleWebSocketsHandler) outgoingMessagesDeliver(ticker *time.Ticker) {
	for {
		select {
		case <-ticker.C:
			r := rand.New(rand.NewSource(time.Now().UnixNano()))

			for connectionID, wsConnection := range handler.connections {
				if r.Intn(20) < 5 {
					handler.log.Debugf("Sending message to connection #%s (%s)", connectionID, wsConnection.RemoteAddr().String())

					err := wsConnection.WriteMessage(1, []byte("Hello from server"))

					if err == nil {
						handler.connectionsStats.OutgoingMessageSent(connectionID)
					}
				}
			}
		}
	}
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

	handler.saveConnection(connectionID, wsConn)
	defer handler.deleteConnection(connectionID)

	handler.connectionsStats.Add(connectionID, deviceID, deviceType, userAgent, r.RemoteAddr)
	defer handler.connectionsStats.Delete(connectionID)

	handler.handleIncommingMessages(connectionID, wsConn)
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

func (handler *SampleWebSocketsHandler) saveConnection(connectionID string, ws *websocket.Conn) error {
	handler.dataMutex.Lock()
	defer handler.dataMutex.Unlock()

	_, alreadyConnected := handler.connections[connectionID]

	if alreadyConnected {
		return errors.New("Connection already established")
	}

	handler.connections[connectionID] = ws

	return nil
}

func (handler *SampleWebSocketsHandler) deleteConnection(connectionID string) {
	handler.dataMutex.Lock()
	defer handler.dataMutex.Unlock()

	delete(handler.connections, connectionID)
}

func (handler *SampleWebSocketsHandler) handleIncommingMessages(connectionID string, wsConn *websocket.Conn) {
	for {
		messageType, message, err := wsConn.ReadMessage()

		if err != nil {
			if handler.shutdownInProgress == false {
				handler.log.Error("Error reading message ", err)
			}

			return
		}

		handler.log.Debugf("recv: Type: %d, %s", messageType, message)

		handler.connectionsStats.IncomingMessageReceived(connectionID)
	}
}

/*
Stats TODO
*/
func (handler *SampleWebSocketsHandler) Stats() storage.DeviceConnectionsStatsStorageInterface {
	return handler.connectionsStats
}
