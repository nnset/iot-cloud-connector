package bus

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/nnset/iot-cloud-connector/events"
)

type InMemoryEventBus struct {
	subscriptions      map[string][]*chan events.Message // subscriptions[topic] => [chan events.Message]
	lock               sync.Mutex
	TotalSubscriptions int
}

// NewInMemoryEventBus
func NewInMemoryEventBus() (*InMemoryEventBus, error) {
	return &InMemoryEventBus{
		subscriptions:      make(map[string][]*chan events.Message),
		lock:               sync.Mutex{},
		TotalSubscriptions: 0,
	}, nil
}

func (bus *InMemoryEventBus) Subscribe(topic string, channel *chan events.Message) error {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	// TODO check for repeated subscriptions
	bus.subscriptions[topic] = append(bus.subscriptions[topic], channel)

	bus.TotalSubscriptions++

	return nil
}

func (bus *InMemoryEventBus) Publish(topic string, message events.Message) error {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	if _, ok := bus.subscriptions[topic]; !ok {
		return fmt.Errorf("topic %s doesn't exist", topic)
	}

	for _, subs := range bus.subscriptions[topic] {
		*subs <- message
	}

	return nil
}

func (bus *InMemoryEventBus) Unsubscribe(topic string, channel *chan events.Message) error {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	if _, ok := bus.subscriptions[topic]; ok && len(bus.subscriptions[topic]) > 0 {
		bus.unsubscribeChannel(topic, bus.findChannelIdx(topic, reflect.ValueOf(channel)))

		return nil
	}

	return fmt.Errorf("topic %s doesn't exist", topic)
}

func (bus *InMemoryEventBus) findChannelIdx(topic string, channel reflect.Value) int {

	if _, ok := bus.subscriptions[topic]; ok {
		for idx, subs := range bus.subscriptions[topic] {
			if reflect.ValueOf(subs).Type() == channel.Type() &&
				reflect.ValueOf(subs).Pointer() == channel.Pointer() {

				return idx
			}
		}
	}

	return -1
}

func (bus *InMemoryEventBus) unsubscribeChannel(topic string, channelIndex int) {
	if _, ok := bus.subscriptions[topic]; !ok {
		return
	}

	l := len(bus.subscriptions[topic])

	if channelIndex < 0 || channelIndex >= l {
		return
	}

	copy(bus.subscriptions[topic][channelIndex:], bus.subscriptions[topic][channelIndex+1:])
	bus.subscriptions[topic][l-1] = nil // or the zero value of T
	bus.subscriptions[topic] = bus.subscriptions[topic][:l-1]
	bus.TotalSubscriptions--
}
