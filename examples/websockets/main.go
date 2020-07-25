package main

import (
	"fmt"
	"os"

	"github.com/nnset/iot-cloud-connector/connectionshandlers"
	"github.com/nnset/iot-cloud-connector/servers"
	"github.com/nnset/iot-cloud-connector/storage"
	"github.com/sirupsen/logrus"
)

func main() {
	log := createLogger()

	connectionsHandler := connectionshandlers.NewSampleWebSocketsHandler(
		"localhost", "8080", "tcp", "", "",
	)

	cors := servers.CrossOriginResourceSharing{
		Headers: "Content-Type, Access-Control-Request-Method, Authorization",
		Origin:  "*",
	}

	defaultAPI := servers.NewDefaultCloudConnectorAPI("localhost", "9090", &servers.APINoAuthenticationMiddleware{}, &cors)

	s := servers.NewCloudConnector(
		log, connectionsHandler, storage.NewInMemoryDeviceConnectionsStorage(), defaultAPI,
	)

	s.Start()

	log.Debug("Finished shutdown")

	os.Exit(0)
}

func createLogger() *logrus.Logger {
	var log = logrus.New()

	log.SetLevel(logrus.DebugLevel)
	log.Out = os.Stderr

	file, err := os.OpenFile("../var/log/sockets.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err == nil {
		log.Out = file
	} else {
		fmt.Println("Using stdErr for log")
	}

	return log
}
