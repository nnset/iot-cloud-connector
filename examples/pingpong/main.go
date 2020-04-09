package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nnset/iot-cloud-connector/connectionshandlers"
	"github.com/nnset/iot-cloud-connector/servers"
	"github.com/sirupsen/logrus"
)

func main() {
	log := createLogger()
	operatingSystemSignal := make(chan os.Signal, 1)
	shutdownServer := make(chan bool, 1)

	signal.Notify(operatingSystemSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	go func(log *logrus.Logger) {
		sig := <-operatingSystemSignal
		log.Debugf("Signal received : %s", sig)
		log.Debug("Shutting down main")
		shutdownServer <- true
	}(log)

	connectionsHandler := connectionshandlers.NewSamplePingPongHandler(
		"localhost", "8080", "tcp",
	)

	s := servers.NewCloudServer(
		"localhost", "9090", "tcp", log, &shutdownServer, connectionsHandler,
	)

	s.Start()

	log.Debug("Finished shutdown")

	os.Exit(0)
}

func createLogger() *logrus.Logger {
	var log = logrus.New()

	log.SetLevel(logrus.DebugLevel)
	log.Out = os.Stderr

	file, err := os.OpenFile("../var/log/ping-pong.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err == nil {
		log.Out = file
	} else {
		fmt.Println("Using stdErr for log")
	}

	return log
}
