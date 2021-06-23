package services

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nnset/iot-cloud-connector/bus"
	"github.com/nnset/iot-cloud-connector/events"
	"gotest.tools/assert"
)

func TestCreatingNewInMemoryConnectionsStorageServiceShouldReturnAnInstanceOfIt(t *testing.T) {
	eventBus, _ := bus.NewInMemoryEventBus()

	service, err := NewInMemoryConnectionsStorageService(eventBus)

	assert.NilError(t, err)
	assert.Assert(t, len(service.Id()) > 0)
}

func TestEstablishedConnectionsShouldBeCounted(t *testing.T) {
	eventBus, _ := bus.NewInMemoryEventBus()

	service, _ := NewInMemoryConnectionsStorageService(eventBus)

	shutdownService := make(chan bool)

	service.Init(shutdownService)

	go service.Start()

	time.Sleep(200 * time.Millisecond)

	assert.Assert(t, service.ActiveConnectionsCount() == 0)

	m := events.NewMessage("{\"device_id\": \"abc-123\"}", "192.168.1.100", events.Default)

	eventBus.Publish(events.ConnectionEstablished, m)

	time.Sleep(20 * time.Millisecond)

	assert.Assert(t, service.ActiveConnectionsCount() == 1)
	shutdownService <- true
}

func TestClosedConnectionsShouldBeDiscounted(t *testing.T) {
	eventBus, _ := bus.NewInMemoryEventBus()

	service, _ := NewInMemoryConnectionsStorageService(eventBus)

	shutdownService := make(chan bool)

	service.Init(shutdownService)

	go service.Start()

	time.Sleep(200 * time.Millisecond)

	assert.Assert(t, service.ActiveConnectionsCount() == 0)

	m := events.NewMessage("{\"device_id\": \"abc-123\"}", "192.168.1.100", events.Default)

	eventBus.Publish(events.ConnectionEstablished, m)

	time.Sleep(20 * time.Millisecond)

	assert.Assert(t, service.ActiveConnectionsCount() == 1)

	eventBus.Publish(events.ConnectionClosed, m)

	time.Sleep(20 * time.Millisecond)

	assert.Assert(t, service.ActiveConnectionsCount() == 0)
	shutdownService <- true
}

func TestEstablishingMultipleConnectionsShouldBeThreadSafe(t *testing.T) {
	eventBus, _ := bus.NewInMemoryEventBus()

	service, _ := NewInMemoryConnectionsStorageService(eventBus)

	shutdownService := make(chan bool)

	service.Init(shutdownService)

	go service.Start()

	time.Sleep(200 * time.Millisecond)

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		for i := 0; i < 10; i++ {
			payload := fmt.Sprintf("{\"device_id\": \"%s\"}", uuid.New().String())
			m := events.NewMessage(payload, "192.168.1.100", events.Default)

			eventBus.Publish(events.ConnectionEstablished, m)
		}

		wg.Done()
	}()

	go func() {
		for i := 0; i < 10; i++ {
			payload := fmt.Sprintf("{\"device_id\": \"%s\"}", uuid.New().String())
			m := events.NewMessage(payload, "192.168.1.100", events.Default)

			eventBus.Publish(events.ConnectionEstablished, m)
		}

		wg.Done()
	}()

	go func() {
		for i := 0; i < 40; i++ {
			payload := fmt.Sprintf("{\"device_id\": \"%s\"}", uuid.New().String())
			m := events.NewMessage(payload, "192.168.1.100", events.Default)

			eventBus.Publish(events.ConnectionEstablished, m)
		}

		wg.Done()
	}()

	c := make(chan struct{})

	go func() {
		defer close(c)

		wg.Wait()
	}()

	select {
	case <-c:
		assert.Assert(t, service.ActiveConnectionsCount() == 60)
		assert.Assert(t, len(service.ActiveConnections()) == 60)
		shutdownService <- true
	case <-time.After(4 * time.Second):
		shutdownService <- true
		assert.Assert(t, false)
	}
}
