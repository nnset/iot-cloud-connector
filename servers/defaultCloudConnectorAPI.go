package servers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
)

// DefaultCloudConnectorAPI We provide a simple REST API to interact with CloudConenctor.
// These are the default available endpoints:
// - {GET} /cloud-connector/status
// - {GET} /devices/status/{deviceID}
// - {GET} /devices
// - {POST} /devices/command/{deviceID}
// - {POST} /devices/query/{deviceID}
//
// Read API docs at /docs/default-cloud-connector-api.md
//
type DefaultCloudConnectorAPI struct {
	address        string
	port           string
	cloudConnector *CloudConnector
	httpServer     *http.Server
	auth           APIAuthMiddleWare
}

// NewDefaultCloudConnectorAPI Creates a new DefaultCloudConnectorAPI
func NewDefaultCloudConnectorAPI(address, port string, auth APIAuthMiddleWare) *DefaultCloudConnectorAPI {
	return &DefaultCloudConnectorAPI{
		address: address,
		port:    port,
		auth:    auth,
	}
}

// Start Starts DefaultCloudConnectorAPI using the configured port.
// You don't have to invoke this method, CloudConnector will.
func (api *DefaultCloudConnectorAPI) Start(cloudConnector *CloudConnector) error {
	if cloudConnector == nil {
		return errors.New("Missing Cloud Connector required instance")
	}

	api.cloudConnector = cloudConnector
	api.cloudConnector.log.Debugf("Starting Default REST API")

	listenAddr := fmt.Sprintf("%s:%s", api.address, api.port)

	api.httpServer = &http.Server{
		Addr:         listenAddr,
		Handler:      api.router(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	api.cloudConnector.log.Debugf("Default REST API available at %s:%s", api.address, api.port)

	// TODO api.httpServer.ListenAndServeTLS

	if err := api.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		api.cloudConnector.log.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
	}

	return nil
}

func (api *DefaultCloudConnectorAPI) router() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/cloud-connector/status", api.status).Methods("GET")
	router.HandleFunc("/devices/status/{deviceID}", api.deviceStatus).Methods("GET")
	router.HandleFunc("/devices", api.devicesList).Methods("GET")
	router.HandleFunc("/devices/command/{deviceID}", api.sendCommand).Methods("POST")
	router.HandleFunc("/devices/query/{deviceID}", api.sendQuery).Methods("POST")

	router.Use(api.auth.Middleware)

	return router
}

func (api *DefaultCloudConnectorAPI) restAPIHeaders(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
}

func (api *DefaultCloudConnectorAPI) unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	json.NewEncoder(w).Encode(
		errorPayload{Error: "Unauthorized"},
	)
}

func (api *DefaultCloudConnectorAPI) status(w http.ResponseWriter, r *http.Request) {
	api.restAPIHeaders(w, http.StatusOK)

	incomingMessages := api.cloudConnector.ReceivedMessages("")
	outgoingMessages := api.cloudConnector.SentMessages("")
	uptimeSeconds := api.cloudConnector.Uptime("") + 1

	json.NewEncoder(w).Encode(
		statusPayload{
			Connections:               api.cloudConnector.OpenConnections(),
			Uptime:                    api.cloudConnector.Uptime(""),
			ReceivedMessages:          incomingMessages,
			ReceivedMessagesPerSecond: float64(int64(incomingMessages) / uptimeSeconds),
			SentMessages:              outgoingMessages,
			SentMessagesPerSecond:     float64(int64(outgoingMessages) / uptimeSeconds),
			CommandsWaiting:           api.cloudConnector.CommandsWaiting(),
			QueriesWaiting:            api.cloudConnector.QueriesWaiting(),
			GoRoutines:                api.cloudConnector.GoRoutinesSpawned(),
			SystemMemory:              api.cloudConnector.SystemMemory(),
			AllocatedMemory:           api.cloudConnector.AllocatedMemory(),
			HeapAllocatedMemory:       api.cloudConnector.HeapAllocatedMemory(),
			ServerCurrentState:        api.cloudConnector.State(),
		},
	)
}

func (api *DefaultCloudConnectorAPI) deviceStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	incomingMessages := api.cloudConnector.ReceivedMessages(vars["deviceID"])
	outgoingMessages := api.cloudConnector.SentMessages(vars["deviceID"])
	uptimeSeconds := api.cloudConnector.Uptime(vars["deviceID"])

	if uptimeSeconds == 0 {
		// TODO return JSON
		http.Error(w, "Device not found", http.StatusNotFound)

		return
	}

	api.restAPIHeaders(w, http.StatusOK)

	json.NewEncoder(w).Encode(
		deviceStatusPayload{
			Uptime:                    uptimeSeconds,
			ReceivedMessages:          incomingMessages,
			ReceivedMessagesPerSecond: float64(int64(incomingMessages) / uptimeSeconds),
			SentMessages:              outgoingMessages,
			SentMessagesPerSecond:     float64(int64(outgoingMessages) / uptimeSeconds),
		},
	)
}

func (api *DefaultCloudConnectorAPI) devicesList(w http.ResponseWriter, r *http.Request) {
	api.restAPIHeaders(w, http.StatusOK)

	devices := api.cloudConnector.ConnectedDevices()
	sort.Strings(devices)

	json.NewEncoder(w).Encode(
		devicesListPayload{
			Devices: devices,
		},
	)
}

func (api *DefaultCloudConnectorAPI) sendCommand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	commandResponse, responseCode, err := api.cloudConnector.SendCommand(api.rawRequestBody(r), vars["deviceID"])

	api.responseFromDevice(w, commandResponse, responseCode, err)
}

func (api *DefaultCloudConnectorAPI) sendQuery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	queryResponse, responseCode, err := api.cloudConnector.SendQuery(api.rawRequestBody(r), vars["deviceID"])

	api.responseFromDevice(w, queryResponse, responseCode, err)
}

func (api *DefaultCloudConnectorAPI) responseFromDevice(w http.ResponseWriter, commandResponse string, responseCode int, err error) {

	payload := deviceResponsePayload{
		Response: commandResponse,
	}

	if err != nil {
		payload = deviceResponsePayload{
			Response: "",
			Errors:   err.Error(),
		}
	}

	api.restAPIHeaders(w, responseCode)
	json.NewEncoder(w).Encode(payload)
}

func (api *DefaultCloudConnectorAPI) rawRequestBody(r *http.Request) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)

	return buf.String()
}

/*
Stop Stops the API server
@see https://marcofranssen.nl/go-webserver-with-graceful-shutdown/
*/
func (api *DefaultCloudConnectorAPI) Stop() {
	api.cloudConnector.log.Debug("Shutting down defaultCloudConnectorAPI")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	api.httpServer.SetKeepAlivesEnabled(false)

	if err := api.httpServer.Shutdown(ctx); err != nil {
		api.cloudConnector.log.Fatalf("Could not gracefully shutdown defaultCloudConnectorAPI http server %v\n", err)
	} else {
		api.cloudConnector.log.Debug("defaultCloudConnectorAPI is shutdown")
	}
}
