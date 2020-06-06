# Default cloud connector API

## Source code

Check the [source code](/servers/defaultCloudConnectorAPI.go)

## Endpoints

### Cloud Connector

#### Current Status

> **GET** `/cloud-connector/status`

Some metrics on how Cloud Connector instance is performing.

**Headers**

TODO

**Success**

> HTTP/1.1 **200** OK

```json
{
  "metrics": {
    "server_current_state": "started",
    "connections": 300,
    "uptime": 897,
    "received_messages": 300,
    "received_messages_per_second": 0,
    "sent_messages": 0,
    "sent_messages_per_second": 0,
    "commands_waiting": 0,
    "queries_waiting": 0,
    "go_routines": 309,
    "system_memory": 69,
    "allocated_memory": 4,
    "heap_allocated_memory": 4
  },
  "units": {
    "server_current_state": "",
    "connections": "",
    "uptime": "secs",
    "received_messages": "",
    "received_messages_per_second": "",
    "sent_messages": "",
    "sent_messages_per_second": "",
    "commands_waiting": "",
    "queries_waiting": "",
    "go_routines": "",
    "system_memory": "Mb",
    "allocated_memory": "Mb",
    "heap_allocated_memory": "Mb"
  }
}
```

| Field                                     |  Type  | Description |
| ------                                    | ------ |------ |
|  metrics                                  | object | Server's metrics. |
|  **metrics**.server_current_state         | string | Server's current state. |
|  **metrics**.connections                  | int    | How many connections are currently open. |
|  **metrics**.uptime                       | int    | Server uptime in seconds. |
|  **metrics**.received_messages            | int    | How may messages the server received. |
|  **metrics**.received_messages_per_second | int    | How may messages the server is receiving per second. |
|  **metrics**.sent_messages                | int    | How may messages the server sent to the connected clients. |
|  **metrics**.sent_messages_per_second     | int    | How may messages the server is sending per second. |
|  **metrics**.commands_waiting             | int    | How many commands to devices are currently waiting feedback from the device. |
|  **metrics**.queries_waiting              | int    | How many queries to devices are currently waiting device's response. |
|  **metrics**.go_routines                  | int    | How may Go routines are current spawned. |
|  **metrics**.system_memory                | int    | Total mega bytes of memory obtained from the OS. |
|  **metrics**.allocated_memory             | int    | Mega bytes allocated for heap objects. |
|  **metrics**.heap_allocated_memory        | int    | Mega bytes of allocated heap objects. |
|  units                                    | object | Server's metrics units. |
|  **units**.server_current_state           | string | "" |
|  **units**.connections                    | string | "" |
|  **units**.uptime                         | string | "secs" |
|  **units**.received_messages              | string | "" |
|  **units**.received_messages_per_second   | string | "" |
|  **units**.sent_messages                  | string | "" |
|  **units**.sent_messages_per_second       | string | "" |
|  **units**.commands_waiting               | string | "" |
|  **units**.queries_waiting                | string | "" |
|  **units**.go_routines                    | string | "" |
|  **units**.system_memory                  | string | "Mb" |
|  **units**.allocated_memory               | string | "Mb" |
|  **units**.heap_allocated_memory          | string | "Mb" |


#### Current Status Stream (SSE)

> **GET** `/cloud-connector/status/stream`

Data Stream ([Server Sent Events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events)) with changes on Cloud Connector server.

**Headers**

Response will have these headers:

    "Connection: keep-alive"
    "Cache-Control: no-cache"
    "Content-Type: text/event-stream"

**Success**

> HTTP/1.1 **200** OK

Each Event will be like this :

```
id: "" (optional)
retry: 1 (optional)
event: system_status
data: {"metric":"go_routines","value":"17"}

```

| Field   |  Type  | Description |
| ------  | ------ |------ |
|  id     | string | Event ID. |
|  retry  | integer | Event retry number. |
|  event  | string | Event's name. |
|  data   | string | Event's payload. |


**Client (javascript) example**

Read [Mozilla Docs](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events) regarding **Server Sent Events** class.

Also [this article](https://community.hetzner.com/tutorials/real-time-apps-with-go-and-reactjs) from Hetzner, about SSE using Go, helped a lot.

```javascript

const evtSource = new EventSource('http://localhost:9090/cloud-connector/status/stream');

evtSource.onerror = function(error) {
  console.error("EventSource failed:", error);
};

evtSource.addEventListener("system_status", function(e) {
  const data = JSON.parse(e.data);

  if (data.metric) {
    console.error(`Metric ${data.metric} changed to ${data.value}.`);
  }
})

evtSource.onopen = function() {
  console.log("Connection to server opened.");
};
```

### IoT Devices

#### Device status

> **GET** `/devices/:deviceID/show`

**Headers**

TODO

**Parameters**

| Name | Description |
| ------------- | ------------- |
| deviceID | IoT Device's unique identifier which was used to establish a connection to Cloud Connector. |


**Success**

> HTTP/1.1 **200** OK

```json
{
  "metrics": {
    "uptime": 100,
    "received_messages": 0,
    "received_messages_per_second": 0,
    "sent_messages": 0,
    "sent_messages_per_second": 0
  },
  "units": {
    "uptime": "secs",
    "received_messages": "",
    "received_messages_per_second": "",
    "sent_messages": "",
    "sent_messages_per_second": ""
  }
}
```

| Field                                     |  Type  | Description |
| ------                                    | ------ |------ |
|  metrics                                  | Object | Device's metrics. |
|  **metrics**.uptime                       | int    | Device connection uptime in seconds. |
|  **metrics**.received_messages            | int    | How may messages the device sent to the server. |
|  **metrics**.received_messages_per_second | int    | How many messages the device is sending to the server per second. |
|  **metrics**.sent_messages                | int    | How may messages the device received from the server. |
|  **metrics**.sent_messages_per_second     | int    | How may messages the device is receiving from the server per second. |
|  units                                    | Object | Device's metrics units. |
|  **units**.uptime                         | string | "secs" |
|  **units**.received_messages              | string | "" |
|  **units**.received_messages_per_second   | string | "" |
|  **units**.sent_messages                  | string | "" |
|  **units**.sent_messages_per_second       | string | "" |

**Error**

|                | Response code | Message |
| -------------  | ------------- |  ------------- |
| DeviceNotFound | 404           | DeviceNotFound The <code>deviceID</code> of the Device was not found. |

> HTTP/1.1 **404** Not found

```json
{
    "error": "Device not found"
}
```

#### Devices list

> **GET** `/status`

Which devices are currently connected.

**Headers**

TODO

**Success**

> HTTP/1.1 **200** OK

```json
{
    "devices": [
        "device_id_1", "device_id_2"
    ]
}
```

| Field     |  Type    | Description |
| ------    | ------   |------ |
|  devices  | string[] | Connected IoT Devices IDs. |


#### Send a command

> **POST** `/devices/command/:deviceID`

Send a **command** to a connected IoT device. Submitted content will be forwarded to the device.

**Headers**

TODO

**Parameters**

| Name | Description |
| ------------- | ------------- |
| deviceID | IoT Device's unique identifier which was used to establish a connection to Cloud Connector. |
| payload | Payload content to be delivered to IoT Devices. |

```json
{
  "payload": "string"
}
```

**Success**

> HTTP/1.1 **200** OK

```json
{
  "response": "string",
  "errors": ""
}
```

| Field              |  Type  | Description |
| ------             | ------ |------ |
|  response | string | Device's response to the command. |
|  errors   | string | Possible error messages. |


**Error**

|                | Response code | Message |
| -------------  | ------------- |  ------------- |
| DeviceNotFound | 404           | The <code>deviceID</code> of the Device was not found |
| TimeOut        | 408           | Command to Device timed out |


> HTTP/1.1 **404** Not found

```json
{
    "response": "",
    "error": "Device not found"
}
```

> HTTP/1.1 **408** Time out

```json
{
    "response": "",
    "error": "Device command timeout"
}
```

#### Send a query

> **POST** `/devices/query/:deviceID`

Send a **query** to a connected IoT device. Submitted content will be forwarded to the device.

**Headers**

TODO

**Parameters**

| Name | Description |
| ------------- | ------------- |
| deviceID | IoT Device's unique identifier which was used to establish a connection to Cloud Connector. |
| payload | Payload content to be delivered to IoT Devices. |

```json
{
  "payload": "string"
}
```

**Success**

> HTTP/1.1 **200** OK

```json
{
  "response": "string",
  "errors": ""
}
```

| Field              |  Type  | Description |
| ------             | ------ |------ |
|  response | string | Device's response to the query. |
|  errors   | string | Possible error messages. |


**Error**

|                | Response code | Message |
| -------------  | ------------- |  ------------- |
| DeviceNotFound | 404           | The <code>deviceID</code> of the Device was not found |
| TimeOut        | 408           | Query to Device timed out |


> HTTP/1.1 **404** Not found

```json
{
    "response": "",
    "error": "Device not found"
}
```

> HTTP/1.1 **408** Time out

```json
{
    "response": "",
    "error": "Device query timeout"
}
```
