package servers

type statusPayload struct {
	ServerCurrentState        CloudConnectorState `json:"server_current_state"`
	Connections               uint                `json:"connections"`
	Uptime                    int64               `json:"uptime"`
	IncomingMessages          uint                `json:"incoming_messages"`
	IncomingMessagesPerSecond float64             `json:"incoming_messages_per_second"`
	OutgoingMessages          uint                `json:"outgoing_messages"`
	OutgoingMessagesPerSecond float64             `json:"outgoing_messages_per_second"`
	CommandsWaiting           uint                `json:"commands_waiting"`
	QueriesWaiting            uint                `json:"queries_waiting"`
	GoRoutines                int                 `json:"go_routines"`
	SystemMemory              uint                `json:"system_memory"`
	AllocatedMemory           uint                `json:"allocated_memory"`
	HeapAllocatedMemory       uint                `json:"heap_allocated_memory"`
}

type deviceStatusPayload struct {
	Uptime                    int64   `json:"uptime"`
	IncomingMessages          uint    `json:"incoming_messages"`
	IncomingMessagesPerSecond float64 `json:"incoming_messages_per_second"`
	OutgoingMessages          uint    `json:"outgoing_messages"`
	OutgoingMessagesPerSecond float64 `json:"outgoing_messages_per_second"`
}

type devicesListPayload struct {
	Devices []string `json:"devices"`
}

type deviceResponsePayload struct {
	Response string `json:"response"`
	Errors   string `json:"errors"`
}

type errorPayload struct {
	Error string `json:"error"`
}
