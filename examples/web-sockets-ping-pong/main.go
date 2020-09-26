package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/nnset/iot-cloud-connector/connectionshandlers"
	"github.com/nnset/iot-cloud-connector/servers"
	"github.com/nnset/iot-cloud-connector/storage"
	"github.com/sirupsen/logrus"
)

func main() {
	log := createLogger()

	connectionsHandler := connectionshandlers.NewWebSocketsHandler(
		"localhost", "8080", "tcp", "", "",
		pingPongHandler,
		authenticateConnection,
	)

	cors := servers.CrossOriginResourceSharing{
		Headers: "Content-Type, Access-Control-Request-Method, Authorization",
		Origin:  "*",
	}

	defaultAPI := servers.NewDefaultCloudConnectorAPI(
		"localhost",
		"9090",
		"",
		"",
		&servers.APINoAuthenticationMiddleware{},
		&cors,
	)

	s := servers.NewCloudConnector(
		log, connectionsHandler, storage.NewInMemoryDeviceConnectionsStorage(), defaultAPI, nil,
	)

	s.Start(15)

	os.Exit(0)
}

func pingPongHandler(deviceID string, messageType int, p []byte) error {
	return nil
}

func authenticateConnection(r *http.Request) error {
	// Authentication: Bearer ____token____
	authToken := strings.Split(r.Header.Get("Authentication"), " ")[1]

	registeredDevicesTokens := make([]string, 3)

	registeredDevicesTokens[0] = "abc123"
	registeredDevicesTokens[1] = "abc124"
	registeredDevicesTokens[2] = "abc125"

	for _, token := range registeredDevicesTokens {
		if authToken == token {
			return nil
		}
	}

	return errors.New("Unauthorized")
}

func createLogger() *logrus.Logger {
	var log = logrus.New()

	log.SetLevel(logrus.DebugLevel)
	log.Out = os.Stderr

	file, err := os.OpenFile("ping-pong-server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err == nil {
		log.Out = file
	} else {
		fmt.Println("Using stdErr for log")
	}

	return log
}
