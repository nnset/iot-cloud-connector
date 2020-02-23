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
CloudServerAPI This REST API server will let you check for stats and status regarding the server
and all its open connections
*/
type CloudServerAPI struct {
	address     string
	port        string
	log         *logrus.Logger
	cloudServer *CloudServer
	httpServer  *http.Server
}

/*
NewCloudServerAPI Creates a new API server
*/
func NewCloudServerAPI(address, port string, log *logrus.Logger, cloudServer *CloudServer) *CloudServerAPI {

	return &CloudServerAPI{
		address:     address,
		port:        port,
		log:         log,
		cloudServer: cloudServer,
	}
}

/*
Start Starts the API server on the configured port
*/
func (api *CloudServerAPI) Start() {
	api.cloudServer.log.Debug("CloudServerAPI seting up routes")

	listenAddr := fmt.Sprintf("%s:%s", api.address, api.port)

	api.httpServer = &http.Server{
		Addr:         listenAddr,
		Handler:      api.router(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	api.cloudServer.log.Debugf("CloudServerAPI available at %s:%s", api.address, api.port)

	// TODO api.httpServer.ListenAndServeTLS

	if err := api.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		api.cloudServer.log.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
	}
}

func (api *CloudServerAPI) router() *mux.Router {
	// TODO does mux have a middleware in order to perform auth ?
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/status", api.status)

	return router
}

//
// Payloads
//

type statusPayload struct {
	Connections               uint    `json:"connections"`
	Uptime                    int64   `json:"uptime"`
	IncomingMessages          uint    `json:"incoming_messages"`
	IncomingMessagesPerSecond float64 `json:"incoming_messages_per_second"`
	OutgoingMessages          uint    `json:"outgoing_messages"`
	OutgoingMessagesPerSecond float64 `json:"outgoing_messages_per_second"`
	GoRoutines                int     `json:"go_routines"`
	SystemMemory              uint    `json:"system_memory"`
	AllocatedMemory           uint    `json:"allocated_memory"`
	HeapAllocatedMemory       uint    `json:"heap_allocated_memory"`
}

type errorPayload struct {
	Error string `json:"error"`
}

func (api *CloudServerAPI) restAPIHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (api *CloudServerAPI) authRequest(r *http.Request) error {
	// TODO security layer
	return nil
}

func (api *CloudServerAPI) unauthorized(w http.ResponseWriter) {
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
func (api *CloudServerAPI) status(w http.ResponseWriter, r *http.Request) {
	if api.authRequest(r) != nil {
		api.unauthorized(w)
		return
	}

	api.restAPIHeaders(w)

	incomingMessages := api.cloudServer.IncomingMessages()
	outgoingMessages := api.cloudServer.OutgoingMessages()
	uptimeSeconds := api.cloudServer.Uptime() + 1

	json.NewEncoder(w).Encode(
		statusPayload{
			Connections:               api.cloudServer.OpenConnections(),
			Uptime:                    api.cloudServer.Uptime(),
			IncomingMessages:          incomingMessages,
			IncomingMessagesPerSecond: float64(int64(incomingMessages) / uptimeSeconds),
			OutgoingMessages:          outgoingMessages,
			OutgoingMessagesPerSecond: float64(int64(outgoingMessages) / uptimeSeconds),
			GoRoutines:                api.cloudServer.GoRoutinesSpawned(),
			SystemMemory:              api.cloudServer.SystemMemory(),
			AllocatedMemory:           api.cloudServer.AllocatedMemory(),
			HeapAllocatedMemory:       api.cloudServer.HeapAllocatedMemory(),
		},
	)
}

/*
Stop Stops the API server
@see https://marcofranssen.nl/go-webserver-with-graceful-shutdown/
*/
func (api *CloudServerAPI) Stop() {
	api.log.Debug("Shutting down CloudServerAPI")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	api.httpServer.SetKeepAlivesEnabled(false)

	if err := api.httpServer.Shutdown(ctx); err != nil {
		api.log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	} else {
		api.log.Debug("CloudServerAPI is shutdown")
	}
}
