package services

import (
	"strconv"
	"testing"
	"time"

	"github.com/nnset/iot-cloud-connector/bus"
	"github.com/nnset/iot-cloud-connector/events"

	"gotest.tools/assert"
)

func TestCreatingNewDefaultSystemMetricsServiceShouldNotPublishOnEventBus(t *testing.T) {
	eventBus, _ := bus.NewInMemoryEventBus()

	publishMetricsInterval := 1

	NewDefaultSystemMetricsService(eventBus, publishMetricsInterval)

	topicChannel := make(chan events.Message)

	eventBus.Subscribe(events.SystemMetricsNumGoRoutinesTopic, &topicChannel)

	select {
	case <-topicChannel:
		assert.Assert(t, false)
	case <-time.After(2 * time.Second):
		assert.Assert(t, true)
	}
}

func TestStaringDefaultSystemMetricsServiceShouldPublishOnEventBus(t *testing.T) {
	eventBus, _ := bus.NewInMemoryEventBus()
	shutdownChannel := make(chan bool)

	publishMetricsInterval := 1

	service := NewDefaultSystemMetricsService(eventBus, publishMetricsInterval)
	service.Init(shutdownChannel)

	topicChannel := make(chan events.Message)

	eventBus.Subscribe(events.SystemMetricsNumGoRoutinesTopic, &topicChannel)

	go service.Start()

	select {
	case message := <-topicChannel:
		_, err := strconv.ParseInt(message.Payload, 10, 64)

		assert.Assert(t, err == nil)
		assert.Assert(t, message.MessagType == events.Default)

	case <-time.After(2 * time.Second):
		assert.Assert(t, false)
	}
}
