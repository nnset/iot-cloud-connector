package servers

type cloudConnectorStatusPayload struct {
	Metrics struct {
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
		SSESubscribers            uint                `json:"sse_subscribers"`
	} `json:"metrics"`
	Units struct {
		ServerCurrentState        string `json:"server_current_state"`
		Connections               string `json:"connections"`
		Uptime                    string `json:"uptime"`
		ReceivedMessages          string `json:"received_messages"`
		ReceivedMessagesPerSecond string `json:"received_messages_per_second"`
		SentMessages              string `json:"sent_messages"`
		SentMessagesPerSecond     string `json:"sent_messages_per_second"`
		CommandsWaiting           string `json:"commands_waiting"`
		QueriesWaiting            string `json:"queries_waiting"`
		GoRoutines                string `json:"go_routines"`
		SystemMemory              string `json:"system_memory"`
		AllocatedMemory           string `json:"allocated_memory"`
		HeapAllocatedMemory       string `json:"heap_allocated_memory"`
		SSESubscribers            string `json:"sse_subscribers"`
	} `json:"units"`
}

type showDevicePayload struct {
	Metrics struct {
		Uptime                    int64   `json:"uptime"`
		ReceivedMessages          uint    `json:"received_messages"`
		ReceivedMessagesPerSecond float64 `json:"received_messages_per_second"`
		SentMessages              uint    `json:"sent_messages"`
		SentMessagesPerSecond     float64 `json:"sent_messages_per_second"`
	} `json:"metrics"`
	Units struct {
		Uptime                    string `json:"uptime"`
		ReceivedMessages          string `json:"received_messages"`
		ReceivedMessagesPerSecond string `json:"received_messages_per_second"`
		SentMessages              string `json:"sent_messages"`
		SentMessagesPerSecond     string `json:"sent_messages_per_second"`
	} `json:"units"`
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
