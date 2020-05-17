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
  "server_current_state": "string",
  "connections" : 100,
  "uptime": 300,
  "incoming_messages": 300,
  "incoming_messages_per_second": 1,
  "outgoing_messages": 0,
  "outgoing_messages_per_second": 0,
  "commands_waiting": 0,
  "queries_waiting": 0,
  "go_routines": 0,
  "system_memory": 9,
  "allocated_memory": 12
}
```

| Field                         |  Type  | Description |
| ------                        | ------ |------ |
|  server_current_state         | string | Server's current state. |
|  connections                  | int    | How many connections are currently open. |
|  uptime                       | int    | Server uptime in seconds. |
|  received_messages            | int    | How may messages the server received. |
|  received_messages_per_second | int    | How may messages the server is receiving per second. |
|  sent_messages                | int    | How may messages the server sent to the connected clients. |
|  sent_messages_per_second     | int    | How may messages the server is sending per second. |
|  commands_waiting             | int    | How many commands to devices are currently waiting feedback from the device. |
|  queries_waiting              | int    | How many queries to devices are currently waiting device's response. |
|  go_routines                  | int    | How may Go routines are current spawned. |
|  system_memory                | int    | Total mega bytes of memory obtained from the OS. |
|  allocated_memory             | int    | Mega bytes allocated for heap objects. |
|  heap_allocated_memory        | int    | Mega bytes of allocated heap objects. |


### IoT Devices

#### Device status

> **GET** `/devices/status/:deviceID`

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
  "uptime": 100,
  "received_messages": 0,
  "received_messages_per_second": 0,
  "sent_messages": 0,
  "sent_messages_per_second": 0
}
```

| Field                         |  Type  | Description |
| ------                        | ------ |------ |
|  uptime                       | int    | Device connection uptime in seconds. |
|  received_messages            | int    | How may messages the device sent to the server. |
|  received_messages_per_second | int    | How many messages the device is sending to the server per second. |
|  sent_messages                | int    | How may messages the device received from the server. |
|  sent_messages_per_second     | int    | How may messages the device is receiving from the server per second. |


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
