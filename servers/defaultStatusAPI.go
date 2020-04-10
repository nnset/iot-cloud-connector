package servers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

/*
DefaultStatusAPI This REST API server will let you check for stats and status regarding the server
and all its open connections
*/
type DefaultStatusAPI struct {
	address        string
	port           string
	log            *logrus.Logger
	cloudConnector *CloudConnector
	httpServer     *http.Server
}

/*
NewDefaultStatusAPI Creates a new API server
*/
func NewDefaultStatusAPI(address, port string, log *logrus.Logger, cloudConnector *CloudConnector) *DefaultStatusAPI {
	return &DefaultStatusAPI{
		address:        address,
		port:           port,
		log:            log,
		cloudConnector: cloudConnector,
	}
}

/*
Start Starts the API server on the configured port
*/
func (api *DefaultStatusAPI) Start() error {
	api.cloudConnector.log.Debugf("Starting StatusAPI")

	listenAddr := fmt.Sprintf("%s:%s", api.address, api.port)

	api.httpServer = &http.Server{
		Addr:         listenAddr,
		Handler:      api.router(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	api.cloudConnector.log.Debugf("StatusAPI available at %s:%s", api.address, api.port)

	// TODO api.httpServer.ListenAndServeTLS

	if err := api.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		api.cloudConnector.log.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
	}

	return nil
}

func (api *DefaultStatusAPI) router() *mux.Router {
	// TODO does mux have a middleware in order to perform auth ?
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/status", api.status)

	return router
}

//
// Payloads
//

type statusPayload struct {
	ServerCurrentState        CloudConnectorState `json:"server_current_state"`
	Connections               uint                `json:"connections"`
	Uptime                    int64               `json:"uptime"`
	IncomingMessages          uint                `json:"incoming_messages"`
	IncomingMessagesPerSecond float64             `json:"incoming_messages_per_second"`
	OutgoingMessages          uint                `json:"outgoing_messages"`
	OutgoingMessagesPerSecond float64             `json:"outgoing_messages_per_second"`
	GoRoutines                int                 `json:"go_routines"`
	SystemMemory              uint                `json:"system_memory"`
	AllocatedMemory           uint                `json:"allocated_memory"`
	HeapAllocatedMemory       uint                `json:"heap_allocated_memory"`
}

type errorPayload struct {
	Error string `json:"error"`
}

func (api *DefaultStatusAPI) restAPIHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (api *DefaultStatusAPI) authRequest(r *http.Request) error {
	// TODO security layer
	return nil
}

func (api *DefaultStatusAPI) unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	json.NewEncoder(w).Encode(
		errorPayload{Error: "Unauthorized"},
	)
}

/**
 * @api {get} /status Cloud server status
 * @apiName ServerStatus
 * @apiDescription Stats and status of the server
 * @apiGroup Status
 *
 * @apiSuccess {string=created, started, stopped} server_current_state Server's current state
 * @apiSuccess {Integer} connections How many connections are currently open.
 * @apiSuccess {Integer} uptime Server uptime in seconds.
 * @apiSuccess {Integer} incoming_messages How may messages the server received.
 * @apiSuccess {Integer} incoming_messages_per_second How may messages the server is receiving per second.
 * @apiSuccess {Integer} outgoing_messages How may messages the server sent to the connected clients.
 * @apiSuccess {Integer} outgoing_messages_per_second How may messages the server is sending per second.
 * @apiSuccess {Integer} go_routines How may Go routines are current spawned.
 * @apiSuccess {Integer} system_memory Total mega bytes of memory obtained from the OS.
 * @apiSuccess {Integer} allocated_memory Mega bytes allocated for heap objects.
 * @apiSuccess {Integer} heap_allocated_memory Mega bytes of allocated heap objects.
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "server_current_state": string,
 *       "connections" : int,
 *       "uptime": int,
 *       "incoming_messages": int,
 *       "incoming_messages_per_second": int,
 *       "outgoing_messages": int,
 *       "outgoing_messages_per_second": int,
 *       "go_routines": int,
 *       "system_memory": int,
 *       "allocated_memory": int
 *     }
 */
func (api *DefaultStatusAPI) status(w http.ResponseWriter, r *http.Request) {
	if api.authRequest(r) != nil {
		api.unauthorized(w)
		return
	}

	api.restAPIHeaders(w)

	incomingMessages := api.cloudConnector.IncomingMessages()
	outgoingMessages := api.cloudConnector.OutgoingMessages()
	uptimeSeconds := api.cloudConnector.Uptime() + 1

	json.NewEncoder(w).Encode(
		statusPayload{
			Connections:               api.cloudConnector.OpenConnections(),
			Uptime:                    api.cloudConnector.Uptime(),
			IncomingMessages:          incomingMessages,
			IncomingMessagesPerSecond: float64(int64(incomingMessages) / uptimeSeconds),
			OutgoingMessages:          outgoingMessages,
			OutgoingMessagesPerSecond: float64(int64(outgoingMessages) / uptimeSeconds),
			GoRoutines:                api.cloudConnector.GoRoutinesSpawned(),
			SystemMemory:              api.cloudConnector.SystemMemory(),
			AllocatedMemory:           api.cloudConnector.AllocatedMemory(),
			HeapAllocatedMemory:       api.cloudConnector.HeapAllocatedMemory(),
			ServerCurrentState:        api.cloudConnector.State(),
		},
	)
}

/*
Stop Stops the API server
@see https://marcofranssen.nl/go-webserver-with-graceful-shutdown/
*/
func (api *DefaultStatusAPI) Stop() {
	api.log.Debug("Shutting down StatusAPI")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	api.httpServer.SetKeepAlivesEnabled(false)

	if err := api.httpServer.Shutdown(ctx); err != nil {
		api.log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	} else {
		api.log.Debug("StatusAPI is shutdown")
	}
}