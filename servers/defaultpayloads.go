package servers

type statusPayload struct {
	ServerCurrentState        CloudConnectorState `json:"server_current_state"`
	Connections               uint                `json:"connections"`
	Uptime                    int64               `json:"uptime"`
	ReceivedMessages          uint                `json:"received_messages"`
	ReceivedMessagesPerSecond float64             `json:"received_messages_per_second"`
	SentMessages              uint                `json:"sent_messages"`
	SentMessagesPerSecond     float64             `json:"sent_messages_per_second"`
	CommandsWaiting           uint                `json:"commands_waiting"`
	QueriesWaiting            uint                `json:"queries_waiting"`
	GoRoutines                int                 `json:"go_routines"`
	SystemMemory              uint                `json:"system_memory"`
	AllocatedMemory           uint                `json:"allocated_memory"`
	HeapAllocatedMemory       uint                `json:"heap_allocated_memory"`
}

type deviceStatusPayload struct {
	Uptime                    int64   `json:"uptime"`
	ReceivedMessages          uint    `json:"received_messages"`
	ReceivedMessagesPerSecond float64 `json:"received_messages_per_second"`
	SentMessages              uint    `json:"sent_messages"`
	SentMessagesPerSecond     float64 `json:"sent_messages_per_second"`
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
