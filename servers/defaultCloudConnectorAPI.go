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

/*
DefaultCloudConnectorAPI This REST API will let you check stats, status and interact with all your
IoT devices
*/
type DefaultCloudConnectorAPI struct {
	address        string
	port           string
	cloudConnector *CloudConnector
	httpServer     *http.Server
	auth           APIAuthMiddleWare
}

/*
NewDefaultCloudConnectorAPI Creates a new API server
*/
func NewDefaultCloudConnectorAPI(address, port string, auth APIAuthMiddleWare) *DefaultCloudConnectorAPI {
	return &DefaultCloudConnectorAPI{
		address: address,
		port:    port,
		auth:    auth,
	}
}

/*
Start Starts the API server on the configured port
*/
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

func (api *DefaultCloudConnectorAPI) authRequest(r *http.Request) error {
	// TODO security layer
	return nil
}

func (api *DefaultCloudConnectorAPI) unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	json.NewEncoder(w).Encode(
		errorPayload{Error: "Unauthorized"},
	)
}

/**
 * @api {get} /cloud-connector/status Cloud Connector status
 * @apiName ServerStatus
 * @apiDescription Stats and status from this Cloud Connector instance
 * @apiGroup Status
 *
 * @apiSuccess {string=created, started, stopped} server_current_state Server's current state
 * @apiSuccess {Integer} connections How many connections are currently open.
 * @apiSuccess {Integer} uptime Server uptime in seconds.
 * @apiSuccess {Integer} incoming_messages How may messages the server received.
 * @apiSuccess {Integer} incoming_messages_per_second How may messages the server is receiving per second.
 * @apiSuccess {Integer} outgoing_messages How may messages the server sent to the connected clients.
 * @apiSuccess {Integer} outgoing_messages_per_second How may messages the server is sending per second.
 * @apiSuccess {Integer} commands_waiting How many commands to devices are currently waiting feedback from the device.
 * @apiSuccess {Integer} queries_waiting How many queries to devices are currently waiting device's response.
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
 *       "commands_waiting": int,
 *       "queries_waiting": int,
 *       "go_routines": int,
 *       "system_memory": int,
 *       "allocated_memory": int
 *     }
 */
func (api *DefaultCloudConnectorAPI) status(w http.ResponseWriter, r *http.Request) {
	if api.authRequest(r) != nil {
		api.unauthorized(w)
		return
	}

	api.restAPIHeaders(w, http.StatusOK)

	incomingMessages := api.cloudConnector.IncomingMessages("")
	outgoingMessages := api.cloudConnector.OutgoingMessages("")
	uptimeSeconds := api.cloudConnector.Uptime("") + 1

	json.NewEncoder(w).Encode(
		statusPayload{
			Connections:               api.cloudConnector.OpenConnections(),
			Uptime:                    api.cloudConnector.Uptime(""),
			IncomingMessages:          incomingMessages,
			IncomingMessagesPerSecond: float64(int64(incomingMessages) / uptimeSeconds),
			OutgoingMessages:          outgoingMessages,
			OutgoingMessagesPerSecond: float64(int64(outgoingMessages) / uptimeSeconds),
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

/**
 * @api {get} /devices/status/:deviceID Device status
 * @apiName DeviceStatus
 * @apiDescription Stats and status from a Device connection to the server
 * @apiGroup Devices
 *
 * @apiParam {String} deviceID Device's unique identifier
 *
 * @apiSuccess {Integer} uptime Device connection uptime in seconds.
 * @apiSuccess {Integer} incoming_messages How may messages the device sent to the server.
 * @apiSuccess {Integer} incoming_messages_per_second How many messages the device is sending to the server per second.
 * @apiSuccess {Integer} outgoing_messages How may messages the device received from the server.
 * @apiSuccess {Integer} outgoing_messages_per_second How may messages the device is receiving from the server per second.
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "uptime": int,
 *       "incoming_messages": int,
 *       "incoming_messages_per_second": int,
 *       "outgoing_messages": int,
 *       "outgoing_messages_per_second": int
 *     }
 *
 * @apiError DeviceNotFound The <code>deviceID</code> of the Device was not found.
 *
 * @apiErrorExample {json} Error-Response:
 *     HTTP/1.1 404 Not Found
 *     {
 *       "error": "Device not found"
 *     }
 *
 */
func (api *DefaultCloudConnectorAPI) deviceStatus(w http.ResponseWriter, r *http.Request) {
	if api.authRequest(r) != nil {
		api.unauthorized(w)
		return
	}

	vars := mux.Vars(r)
	incomingMessages := api.cloudConnector.IncomingMessages(vars["deviceID"])
	outgoingMessages := api.cloudConnector.OutgoingMessages(vars["deviceID"])
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
			IncomingMessages:          incomingMessages,
			IncomingMessagesPerSecond: float64(int64(incomingMessages) / uptimeSeconds),
			OutgoingMessages:          outgoingMessages,
			OutgoingMessagesPerSecond: float64(int64(outgoingMessages) / uptimeSeconds),
		},
	)
}

/**
 * @api {get} /devices Connected devices
 * @apiName DeviceStatus
 * @apiDescription A list of all connected devices
 * @apiGroup Devices
 *
 * @apiSuccess {String[]} devices Connected Devices IDs
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *        "devices": [
 *          "device_id_1", "device_id_2"
 *        ]
 *     }
 */
func (api *DefaultCloudConnectorAPI) devicesList(w http.ResponseWriter, r *http.Request) {
	if api.authRequest(r) != nil {
		api.unauthorized(w)
		return
	}

	api.restAPIHeaders(w, http.StatusOK)

	devices := api.cloudConnector.ConnectedDevices()
	sort.Strings(devices)

	json.NewEncoder(w).Encode(
		devicesListPayload{
			Devices: devices,
		},
	)
}

/**
 * @api {post} /devices/command/:deviceID
 * @apiName DeviceCommand
 * @apiDescription Send a command to a connected device. Sumbitted content will be forwarded to the device.
 * @apiGroup Devices
 *
 * @apiParam {string} payload
 *
 * @apiSuccess {String="OK",""} response
 * @apiSuccess {String} errors
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *        "response": "string",
 *        "errors": ""
 *     }
 *
 * @apiError DeviceNotFound Device with <code>deviceID</code> is not connected.
 *
 * @apiErrorExample {json} Error-Response:
 *     HTTP/1.1 404 Not Found
 *     {
 *       "response": ""
 *       "errors": "Device :deviceID is not connected"
 *     }
 *
 * @apiError DeviceTimeout Command to Device timed out
 *
 * @apiErrorExample {json} Error-Response:
 *     HTTP/1.1 408 Time Out
 *     {
 *       "response": ""
 *       "errors": "Device command timeout"
 *     }
 *
 */
func (api *DefaultCloudConnectorAPI) sendCommand(w http.ResponseWriter, r *http.Request) {
	if api.authRequest(r) != nil {
		api.unauthorized(w)
		return
	}

	vars := mux.Vars(r)

	commandResponse, responseCode, err := api.cloudConnector.SendCommand(api.rawRequestBody(r), vars["deviceID"])

	api.responseFromDevice(w, commandResponse, responseCode, err)
}

/**
 * @api {post} /devices/query/:deviceID
 * @apiName DeviceQuery
 * @apiDescription Send a query to a connected device. Sumbitted content will be forwarded to the device.
 * @apiGroup Devices
 *
 * @apiParam {string} payload
 *
 * @apiSuccess {String} response Device's response to the query
 * @apiSuccess {String} errors
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *        "response": "string",
 *        "errors": ""
 *     }
 *
 * @apiError DeviceNotFound Device with <code>deviceID</code> is not connected.
 *
 * @apiErrorExample {json} Error-Response:
 *     HTTP/1.1 404 Not Found
 *     {
 *       "response": ""
 *       "errors": "Device :deviceID is not connected"
 *     }
 *
 * @apiError DeviceTimeout Query to Device timed out
 *
 * @apiErrorExample {json} Error-Response:
 *     HTTP/1.1 408 Time Out
 *     {
 *       "response": ""
 *       "errors": "Device query timeout"
 *     }
 *
 */
func (api *DefaultCloudConnectorAPI) sendQuery(w http.ResponseWriter, r *http.Request) {
	if api.authRequest(r) != nil {
		api.unauthorized(w)
		return
	}

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
