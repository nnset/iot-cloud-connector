package connectionshandlers

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nnset/iot-cloud-connector/storage"
	"github.com/sirupsen/logrus"
)

/*
SamplePingPongHandler Simple example of a connections handler that receives
PING and returns PONG on each request. Connections are not permanent.
*/
type SamplePingPongHandler struct {
	id        string
	address   string
	port      string
	network   string
	log       *logrus.Logger
	startTime int64
	messages  uint

	activeConnections storage.DeviceConnectionsStatsStorageInterface
	dataMutex         *sync.Mutex
	httpServer        *http.Server
}

/*
NewSamplePingPongHandler Creates a new instance of SamplePingPongHandler
*/
func NewSamplePingPongHandler(address, port, network string) *SamplePingPongHandler {
	return &SamplePingPongHandler{
		id:                uuid.New().String(),
		address:           address,
		port:              port,
		network:           network,
		startTime:         time.Now().Unix(),
		activeConnections: storage.NewInMemoryDeviceConnectionsStatsStorage(),
		dataMutex:         &sync.Mutex{},
	}
}

/*
Listen Starts a Go http.Server and attends to all incoming connections expecting
the string PING as the requests payload and responding with PONG string, until
shutdownChannel is triggered, closing all connections then.
*/
func (handler *SamplePingPongHandler) Listen(shutdownChannel, shutdownIsCompleteChannel *chan bool, log *logrus.Logger) error {
	handler.log = log

	handler.log.Debugf("SamplePingPongHandler listening to %s:%s", handler.address, handler.port)

	portListener, err := net.Listen(handler.network, handler.address+":"+handler.port)

	if err != nil {
		return err
	}

	defer portListener.Close()

	handler.httpServer = &http.Server{
		Handler:      http.HandlerFunc(handler.handleConnection),
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	defer handler.httpServer.Close()

	go func() {
		handler.log.Debugf("SamplePingPongHandler is serving requests")
		err = handler.httpServer.Serve(portListener)

		if err != http.ErrServerClosed {
			handler.log.Error("SamplePingPongHandler closed ", err)
			return
		}
	}()

	<-*shutdownChannel

	*shutdownIsCompleteChannel <- true

	return nil
}

func (handler *SamplePingPongHandler) handleConnection(w http.ResponseWriter, r *http.Request) {
	handler.log.Debugf("New connection from %s", r.RemoteAddr)

	handler.dataMutex.Lock()
	defer handler.dataMutex.Unlock()
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	payload := string(b)

	handler.log.Debugf("New message %s", payload)

	if payload == "PING" {
		fmt.Fprint(w, "PONG")
	} else {
		http.Error(w, "Invalid payload", 400)
	}

	handler.messages++
}

/*
Stats TODO
*/
func (handler *SamplePingPongHandler) Stats() storage.DeviceConnectionsStatsStorageInterface {
	return handler.activeConnections
}
