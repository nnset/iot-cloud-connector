package connections

import(
    "errors"
    "time"
)

/*
DeviceConnectionStatus Device's connections information used to keep track on how
connections are behaving
*/
type DeviceConnectionStatus struct {
    connectionID   string
    deviceID       string
    deviceType     string
    userAgent      string
    remoteAddress  string
    createdAt      int64
    lastReceivedMessageTimeStamp int64
    lastSentMessageTimeStamp int64
    receivedMessages     uint64
    sentMessages     uint64
}

/*
NewDeviceConnectionStatus Creates a new instance of DeviceConnectionStatus
*/
func NewDeviceConnectionStatus(connectionID, deviceID, deviceType, userAgent,remoteAddress string) *DeviceConnectionStatus {
    return &DeviceConnectionStatus {
        connectionID: connectionID,
        deviceID: deviceID,
        deviceType: deviceType,
        userAgent: userAgent,
        remoteAddress: remoteAddress,
        createdAt: time.Now().Unix(),
    }
}

/*
ConnectionID Connection's UUID
*/
func (status *DeviceConnectionStatus) ConnectionID() string {
    return status.connectionID
}

/*
DeviceID Connection's Device UUID
*/
func (status *DeviceConnectionStatus) DeviceID() string {
    return status.deviceID
}

/*
DeviceType Connection's Device type
*/
func (status *DeviceConnectionStatus) DeviceType() string {
    return status.deviceType
}

/*
UserAgent Connection's user agent
*/
func (status *DeviceConnectionStatus) UserAgent() string {
    return status.userAgent
}

/*
RemoteAddress Connection's remote address
*/
func (status *DeviceConnectionStatus) RemoteAddress() string {
    return status.remoteAddress
}

/*
Uptime how many seconds the connection has been active
*/
func (status *DeviceConnectionStatus) Uptime() (int64, error) {
    if status.createdAt == 0 {
        return -1, errors.New("Connection has not been initiated")
    }

    return time.Now().Unix() - status.createdAt, nil
}

/*
LastReceivedMessageTimeStamp When was the last unix time when we received a message from the connection
*/
func (status *DeviceConnectionStatus) LastReceivedMessageTimeStamp() int64 {
    return status.lastReceivedMessageTimeStamp
}

/*
LastSentMessageTimeStamp When was the last unix time when we sent a message to the connection
*/
func (status *DeviceConnectionStatus) LastSentMessageTimeStamp() int64 {
    return status.lastSentMessageTimeStamp
}

/*
ReceivedMessages How many messages have we received from the connection
*/
func (status *DeviceConnectionStatus) ReceivedMessages() uint64 {
    return status.receivedMessages
}

/*
SentMessages How many messages have we sent to the connection
*/
func (status *DeviceConnectionStatus) SentMessages() uint64 {
    return status.sentMessages
}

/*
MessageSent A message has been sent to the connection
*/
func (status *DeviceConnectionStatus) MessageSent() {
    status.sentMessages++
    status.lastSentMessageTimeStamp = time.Now().Unix()
}

/*
MessageReceived A message was received from the connection
*/
func (status *DeviceConnectionStatus) MessageReceived() {
    status.receivedMessages++
    status.lastReceivedMessageTimeStamp = time.Now().Unix()
}
