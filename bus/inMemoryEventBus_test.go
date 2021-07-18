package bus

import (
	"testing"
	"time"

	"github.com/nnset/iot-cloud-connector/events"
	"gotest.tools/assert"
)

func TestCreatingAnInMemoryEventBusShouldReturnPointer(t *testing.T) {
	eventBus, errors := NewInMemoryEventBus()

	assert.Assert(t, eventBus != nil)
	assert.Assert(t, errors == nil)
	assert.Assert(t, eventBus.TotalSubscriptions == 0)

	// TODO play with reflection
	//assert.Assert(t, (reflect.TypeOf(eventBus) == reflect.PtrTo(InMemoryEventBus{})) == true)
}

func TestSubscribingToTopicsShouldNotReturnError(t *testing.T) {
	eventBus, _ := NewInMemoryEventBus()
	ch := make(chan events.Message)

	err := eventBus.Subscribe("topic", &ch)
	assert.Assert(t, err == nil)
}

func TestSubscribingToTopicsShouldIncreaseTotalSubscriptionsCount(t *testing.T) {
	eventBus, _ := NewInMemoryEventBus()
	ch := make(chan events.Message)

	eventBus.Subscribe("topic", &ch)

	assert.Assert(t, eventBus.TotalSubscriptions == 1)
}

func TestUnSubscribingToTopicsShouldDecreaseTotalSubscriptionsCount(t *testing.T) {
	eventBus, _ := NewInMemoryEventBus()
	ch := make(chan events.Message)

	eventBus.Subscribe("topic", &ch)
	assert.Assert(t, eventBus.TotalSubscriptions == 1)

	eventBus.Unsubscribe("topic", &ch)
	assert.Assert(t, eventBus.TotalSubscriptions == 0)
}

func TestTotalSubscriptionsCountMinValueShouldBeZero(t *testing.T) {
	eventBus, _ := NewInMemoryEventBus()
	ch := make(chan events.Message)

	err := eventBus.Unsubscribe("topic", &ch)
	assert.Assert(t, err != nil)
	assert.Assert(t, eventBus.TotalSubscriptions == 0)

	err = eventBus.Unsubscribe("topic", &ch)
	assert.Assert(t, err != nil)
	assert.Assert(t, eventBus.TotalSubscriptions == 0)
}

func TestSubscribingToEventBusTopicShouldReceiveMessages(t *testing.T) {
	eventBus, _ := NewInMemoryEventBus()
	ch := make(chan events.Message)
	ch2 := make(chan events.Message)

	eventBus.Subscribe("topic", &ch)
	eventBus.Subscribe("topic", &ch2)

	go func() {
		time.Sleep(50 * time.Millisecond)
		eventBus.Publish("topic", events.NewMessage("payload", "address", events.Query)) // todo Query
	}()

	select {
	case event := <-ch:
		assert.Assert(t, event.Payload == "payload")
	case <-time.After(1 * time.Second):
		// Message was not received
		assert.Assert(t, 1 == 0)
	}

	select {
	case event := <-ch2:
		assert.Assert(t, event.Payload == "payload")
	case <-time.After(1 * time.Second):
		// Message was not received
		assert.Assert(t, 1 == 0)
	}
}
