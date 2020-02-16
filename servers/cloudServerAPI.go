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
CloudServerAPI
*/
type CloudServerAPI struct {
	address     string
	port        string
	log         *logrus.Logger
	cloudServer *CloudServer
	httpServer  *http.Server
}

/*
NewCloudServerAPI
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
Start
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

	api.cloudServer.log.Debugf("CloudServerAPI listening to %s:%s", api.address, api.port)

	// TODO api.httpServer.ListenAndServeTLS

	if err := api.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		api.cloudServer.log.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
	}
}

func (api *CloudServerAPI) router() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/status/connections/count", api.connectionsCount)
	router.HandleFunc("/status/uptime", api.uptime)
	router.HandleFunc("/status/messages/incoming-processed", api.incomingMessagesProcessed)

	return router
}

//
// Payloads
//

type connectionsCountPayload struct {
	Count uint `json:"count"`
}

type uptimePayload struct {
	Uptime int64 `json:"uptime"`
}

type totalMessagesProcessedPayload struct {
	Count uint `json:"count"`
}

/**
 * @api {get} /status/connections/count Open connections
 * @apiName OpenConnections
 * @apiDescription How many open connections are currently connected to this instance
 * @apiGroup Status
 *
 * @apiSuccess {String} count Value in seconds
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "count": "7541"
 *     }
 */
func (api *CloudServerAPI) connectionsCount(w http.ResponseWriter, r *http.Request) {
	// TODO Auth request

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	payload := connectionsCountPayload{Count: api.cloudServer.OpenConnections()}

	json.NewEncoder(w).Encode(payload)
}

/**
 * @api {get} /status/uptime Uptime
 * @apiDescription How many seconds the server has been up
 * @apiName Uptime
 * @apiGroup Status
 *
 * @apiSuccess {String} uptime seconds
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "uptime": "3605"
 *     }
 */
func (api *CloudServerAPI) uptime(w http.ResponseWriter, r *http.Request) {
	// TODO Auth request

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	payload := uptimePayload{Uptime: api.cloudServer.Uptime()}

	json.NewEncoder(w).Encode(payload)
}

/**
 * @api {get} /status/messages/incoming-processed Incoming messages
 * @apiDescription How many messages (from client to this server) have been processed so far
 * @apiName IncomingMessages
 * @apiGroup Status
 *
 * @apiSuccess {String} count
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "count": "125784"
 *     }
 */
func (api *CloudServerAPI) incomingMessagesProcessed(w http.ResponseWriter, r *http.Request) {
	// TODO Auth request

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	payload := totalMessagesProcessedPayload{Count: api.cloudServer.IncomingMessagesProcessed()}

	json.NewEncoder(w).Encode(payload)
}

/*
Stop
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
