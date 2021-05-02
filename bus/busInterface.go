package bus

import "github.com/nnset/iot-cloud-connector/events"

type BusSubscriber interface {
	Subscribe(topic string, channel *chan events.Message) error
	Unsubscribe(topic string, channel *chan events.Message) error
}

type BusPublisher interface {
	Publish(topic string, message events.Message) error
}

type MessageBus interface {
	BusSubscriber
	BusPublisher
}
