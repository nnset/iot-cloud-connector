package servers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/nnset/iot-cloud-connector/connectionshandlers"

	"github.com/gorilla/mux"
)

// DefaultCloudConnectorAPI We provide a simple REST API to interact with CloudConenctor.
// Available endpoints:
// - {GET} /cloud-connector/status
// - {GET} /cloud-connector/status/stream
// - {GET} /devices/status/{deviceID}
// - {GET} /devices
// - {POST} /devices/command/{deviceID}
// - {POST} /devices/query/{deviceID}
//
// Read API docs at /docs/default-cloud-connector-api.md
//
type DefaultCloudConnectorAPI struct {
	address         string
	port            string
	certificatePath string
	keyPath         string
	cloudConnector  *CloudConnector
	httpServer      *http.Server
	auth            APIAuthMiddleWare
	cors            *CrossOriginResourceSharing
}

// NewDefaultCloudConnectorAPI Creates a new DefaultCloudConnectorAPI
func NewDefaultCloudConnectorAPI(
	address, port, certificatePath, keyPath string,
	auth APIAuthMiddleWare,
	cors *CrossOriginResourceSharing,
) *DefaultCloudConnectorAPI {
	return &DefaultCloudConnectorAPI{
		address:         address,
		port:            port,
		certificatePath: certificatePath,
		keyPath:         keyPath,
		auth:            auth,
		cors:            cors,
	}
}

// Start Starts DefaultCloudConnectorAPI using the configured port and TLS certificates.
// You don't have to invoke this method, CloudConnector will.
func (api *DefaultCloudConnectorAPI) Start(cloudConnector *CloudConnector) error {
	if cloudConnector == nil {
		return errors.New("Missing Cloud Connector required instance")
	}

	api.cloudConnector = cloudConnector
	api.cloudConnector.log.Debugf("Starting Default REST API")

	api.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%s", api.address, api.port),
		Handler: api.router(),
	}

	api.cloudConnector.log.Infof("Default REST API available at %s:%s", api.address, api.port)

	if api.certificatePath == "" {
		return api.startServer()
	}

	return api.startTLSserver()
}

func (api *DefaultCloudConnectorAPI) startTLSserver() error {
	api.cloudConnector.log.Info("  Default REST API will use TLS")

	if err := api.httpServer.ListenAndServeTLS(api.certificatePath, api.keyPath); err != nil && err != http.ErrServerClosed {
		api.cloudConnector.log.Fatalf("%v\n", err)

		return err
	}

	return nil
}

func (api *DefaultCloudConnectorAPI) startServer() error {
	if err := api.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		api.cloudConnector.log.Fatalf("%v\n", err)

		return err
	}

	return nil
}

func (api *DefaultCloudConnectorAPI) router() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/cloud-connector/status", api.status).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/cloud-connector/status/stream", api.systemMetricsStream).Methods(http.MethodGet)
	router.HandleFunc("/devices/{deviceID}/show", api.showDevice).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/devices", api.connectedDevices).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/devices/command/{deviceID}", api.sendCommand).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/devices/query/{deviceID}", api.sendQuery).Methods(http.MethodPost, http.MethodOptions)

	if api.cors != nil {
		router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Access-Control-Allow-Origin", api.cors.Origin)
				w.Header().Set("Access-Control-Allow-Headers", api.cors.Headers)

				if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
					w.WriteHeader(http.StatusOK)
				} else {
					next.ServeHTTP(w, r)
				}
			})
		})
	}

	router.Use(api.auth.Middleware)

	return router
}

func (api *DefaultCloudConnectorAPI) unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	json.NewEncoder(w).Encode(
		errorPayload{Error: "Unauthorized"},
	)
}

func (api *DefaultCloudConnectorAPI) status(w http.ResponseWriter, r *http.Request) {
	incomingMessages := api.cloudConnector.ReceivedMessages("")
	outgoingMessages := api.cloudConnector.SentMessages("")
	uptimeSeconds := api.cloudConnector.Uptime("") + 1

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(
		cloudConnectorStatusPayload{
			Metrics: struct {
				ServerCurrentState        CloudConnectorState `json:"server_current_state"`
				Connections               uint                `json:"connections"`
				StartTime                 int64               `json:"start_time"`
				ReceivedMessages          uint                `json:"received_messages"`
				ReceivedMessagesPerSecond float64             `json:"received_messages_per_second"`
				SentMessages              uint                `json:"sent_messages"`
				SentMessagesPerSecond     float64             `json:"sent_messages_per_second"`
				CommandsWaiting           uint                `json:"commands_waiting"`
				QueriesWaiting            uint                `json:"queries_waiting"`
				GoRoutines                int                 `json:"go_routines"`
				SystemMemory              uint                `json:"system_memory"`
				AllocatedMemory           uint                `json:"allocated_memory"`
				HeapAllocatedMemory       uint                `json:"heap_allocated_memory"`
				SSESubscribers            uint                `json:"sse_subscribers"`
			}{
				Connections:               api.cloudConnector.OpenConnections(),
				StartTime:                 api.cloudConnector.StartTime(),
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
				SSESubscribers:            api.cloudConnector.SystemMetricsStreamSubscriptions(),
			},
			Units: struct {
				ServerCurrentState        string `json:"server_current_state"`
				Connections               string `json:"connections"`
				Uptime                    string `json:"uptime"`
				ReceivedMessages          string `json:"received_messages"`
				ReceivedMessagesPerSecond string `json:"received_messages_per_second"`
				SentMessages              string `json:"sent_messages"`
				SentMessagesPerSecond     string `json:"sent_messages_per_second"`
				CommandsWaiting           string `json:"commands_waiting"`
				QueriesWaiting            string `json:"queries_waiting"`
				GoRoutines                string `json:"go_routines"`
				SystemMemory              string `json:"system_memory"`
				AllocatedMemory           string `json:"allocated_memory"`
				HeapAllocatedMemory       string `json:"heap_allocated_memory"`
				SSESubscribers            string `json:"sse_subscribers"`
			}{
				ServerCurrentState:        "",
				Connections:               "",
				Uptime:                    "secs",
				ReceivedMessages:          "",
				ReceivedMessagesPerSecond: "",
				SentMessages:              "",
				SentMessagesPerSecond:     "",
				CommandsWaiting:           "",
				QueriesWaiting:            "",
				GoRoutines:                "",
				SystemMemory:              "Mb",
				AllocatedMemory:           "Mb",
				HeapAllocatedMemory:       "Mb",
				SSESubscribers:            "",
			},
		},
	)
}

func (api *DefaultCloudConnectorAPI) showDevice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	incomingMessages := api.cloudConnector.ReceivedMessages(vars["deviceID"])
	outgoingMessages := api.cloudConnector.SentMessages(vars["deviceID"])
	uptimeSeconds := api.cloudConnector.Uptime(vars["deviceID"])

	if uptimeSeconds == 0 {
		// TODO return JSON
		http.Error(w, "Device not found", http.StatusNotFound)

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(
		showDevicePayload{
			Metrics: struct {
				Uptime                    int64   `json:"uptime"`
				ReceivedMessages          uint    `json:"received_messages"`
				ReceivedMessagesPerSecond float64 `json:"received_messages_per_second"`
				SentMessages              uint    `json:"sent_messages"`
				SentMessagesPerSecond     float64 `json:"sent_messages_per_second"`
			}{
				Uptime:                    uptimeSeconds,
				ReceivedMessages:          incomingMessages,
				ReceivedMessagesPerSecond: float64(int64(incomingMessages) / uptimeSeconds),
				SentMessages:              outgoingMessages,
				SentMessagesPerSecond:     float64(int64(outgoingMessages) / uptimeSeconds),
			},
			Units: struct {
				Uptime                    string `json:"uptime"`
				ReceivedMessages          string `json:"received_messages"`
				ReceivedMessagesPerSecond string `json:"received_messages_per_second"`
				SentMessages              string `json:"sent_messages"`
				SentMessagesPerSecond     string `json:"sent_messages_per_second"`
			}{
				Uptime:                    "secs",
				ReceivedMessages:          "",
				ReceivedMessagesPerSecond: "",
				SentMessages:              "",
				SentMessagesPerSecond:     "",
			},
		},
	)
}

func (api *DefaultCloudConnectorAPI) connectedDevices(w http.ResponseWriter, r *http.Request) {
	devices := api.cloudConnector.ConnectedDevices()

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(
		devicesListPayload{
			Devices: devices,
		},
	)
}

func (api *DefaultCloudConnectorAPI) sendCommand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	c := connectionshandlers.NewCommand(vars["deviceID"], api.rawRequestBody(r))

	commandResponse, responseCode, err := api.cloudConnector.SendCommand(c)

	api.responseFromDevice(w, commandResponse, responseCode, err)
}

func (api *DefaultCloudConnectorAPI) sendQuery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	q := connectionshandlers.NewQuery(vars["deviceID"], api.rawRequestBody(r))

	queryResponse, responseCode, err := api.cloudConnector.SendQuery(q)

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

		w.WriteHeader(http.StatusBadRequest)

	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(payload)
}

func (api *DefaultCloudConnectorAPI) systemMetricsStream(w http.ResponseWriter, r *http.Request) {
	api.cloudConnector.log.Debug("systemMetricsStream subscription")

	flusher, ok := w.(http.Flusher)

	if !ok {
		fmt.Println("Streaming unsupported")
		// TODO
		return
	}

	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/event-stream")

	messageChannel := make(chan SystemMetricChangedMessage)

	api.cloudConnector.SubscribeToSystemMetricsStream(messageChannel)

	for {
		select {
		case systemMetricChangedMessage := <-messageChannel:

			// TODO check for errors
			payload, _ := json.Marshal(systemMetricChangedMessage)

			w.Write(api.formatSSE("", "system_status", string(payload), 1))
			flusher.Flush()
		case <-r.Context().Done():
			api.cloudConnector.UnSubscribeToSystemMetricsStream(messageChannel)
			return
		}
	}
}

func (api *DefaultCloudConnectorAPI) formatSSE(id, event, data string, retry int) []byte {

	eventPayload := "id: " + id + "\n" + "retry: " + string(retry) + "\n" + "event: " + event + "\n"

	dataLines := strings.Split(data, "\n")

	for _, line := range dataLines {
		eventPayload = eventPayload + "data: " + line + "\n"
	}

	return []byte(eventPayload + "\n")
}

func (api *DefaultCloudConnectorAPI) rawRequestBody(r *http.Request) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)

	return buf.String()
}

// Stop Stops the API server
// @see https://marcofranssen.nl/go-webserver-with-graceful-shutdown/
func (api *DefaultCloudConnectorAPI) Stop() {
	api.cloudConnector.log.Debug("Shutting down defaultCloudConnectorAPI")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	api.httpServer.SetKeepAlivesEnabled(false)

	// TODO close all systemMetricsStream SSE connections

	if err := api.httpServer.Shutdown(ctx); err != nil {
		api.cloudConnector.log.Fatalf("Could not gracefully shutdown defaultCloudConnectorAPI http server %v\n", err)
	} else {
		api.cloudConnector.log.Debug("defaultCloudConnectorAPI is shutdown")
	}
}
