package bus

type BusHandler interface {
	Start(bus *MessageBus) error
}
