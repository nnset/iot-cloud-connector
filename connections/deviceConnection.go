package connections

import(
    "time"
    "errors"
    
    "github.com/google/uuid"
)

/*
DeviceConnection represents a permanent and bidirectional (if you need it to be) connection between
a server and an IoT device, such as an edge server, a sensor or an actuator.
*/
type DeviceConnection struct {
    id             string
    deviceID       string
    deviceType     string
    userAgent      string
    remoteAddress  string
    createdAt      int64
    stats          deviceConnectionStats
    connection     NetworkConnection
}

/*
NetworkConnection defines the methods required for any kind of network connection
with a device
*/
type NetworkConnection interface {
    Close(statusCode ConnectionStatusCode, reason string) error
}

type deviceConnectionStats struct {
    lastIncomingMessageTimeStamp int64
    lastOutgoingMessageTimeStamp int64
    incomingMessages     uint64
    outgoingMessages     uint64
}

/*
ConnectionStatusCode Network connection statuses
*/
type ConnectionStatusCode int

// https://tools.ietf.org/html/rfc6455#section-7.4
const (
    StatusNormalClosure   ConnectionStatusCode = 1000
    StatusGoingAway       ConnectionStatusCode = 1001
    StatusProtocolError   ConnectionStatusCode = 1002
    StatusUnsupportedData ConnectionStatusCode = 1003

    StatusAbnormalClosure         ConnectionStatusCode = 1006
    StatusInvalidFramePayloadData ConnectionStatusCode = 1007
    StatusPolicyViolation         ConnectionStatusCode = 1008
    StatusMessageTooBig           ConnectionStatusCode = 1009
    StatusInternalError           ConnectionStatusCode = 1011
    StatusServiceRestart          ConnectionStatusCode = 1012
    StatusTryAgainLater           ConnectionStatusCode = 1013
    StatusBadGateway              ConnectionStatusCode = 1014
)

/*
Init Sets all required attributes
*/
func (deviceConn *DeviceConnection) Init(netConn NetworkConnection, deviceID string, remoteAddress string, userAgent string) error {
    if deviceConn.id == "" {
        deviceConn.id = uuid.New().String()
        deviceConn.createdAt = time.Now().Unix()
                
        deviceConn.remoteAddress = remoteAddress
        deviceConn.userAgent = userAgent
        deviceConn.connection = netConn
        deviceConn.deviceID = deviceID
    }

    return nil
}

/*
MessageReceived A message was received from the connected device
*/
func (deviceConn *DeviceConnection) MessageReceived() {
    deviceConn.stats.incomingMessages++
    deviceConn.stats.lastIncomingMessageTimeStamp = time.Now().Unix()
}

/*
MessageSent A message was sent to the connected device
*/
func (deviceConn *DeviceConnection) MessageSent() {
    deviceConn.stats.outgoingMessages++
    deviceConn.stats.lastOutgoingMessageTimeStamp = time.Now().Unix()
}

/*
ID Connection's UUID
*/
func (deviceConn *DeviceConnection) ID() string {
    return deviceConn.id
}

/*
DeviceID Connected device's ID
*/
func (deviceConn *DeviceConnection) DeviceID() string {
    return deviceConn.deviceID
}

/*
UserAgent Connected device's user agent
*/
func (deviceConn *DeviceConnection) UserAgent() string {
    return deviceConn.userAgent
}

/*
RemoteAddress Connected device's remote network address
*/
func (deviceConn *DeviceConnection) RemoteAddress() string {
    return deviceConn.remoteAddress
}

/*
IncomingMessages How many messages were received from the connected device
*/
func (deviceConn *DeviceConnection) IncomingMessages() uint64 {
    return deviceConn.stats.incomingMessages
}

/*
OutgoingMessages How many messages were sent to the connected device
*/
func (deviceConn *DeviceConnection) OutgoingMessages() uint64 {
    return deviceConn.stats.outgoingMessages
}

/*
LatestIncomingMessageTimeStamp When was the last time a message was received from 
the connected device 
*/
func (deviceConn *DeviceConnection) LatestIncomingMessageTimeStamp() int64 {
    return deviceConn.stats.lastOutgoingMessageTimeStamp
}

/*
LatestOutgoingMessageTimeStamp When eas the last time a message was sent to
the connected device 
*/
func (deviceConn *DeviceConnection) LatestOutgoingMessageTimeStamp() int64 {
    return deviceConn.stats.lastOutgoingMessageTimeStamp
}

/*
Close Closes the connection
*/
func (deviceConn *DeviceConnection) Close(statusCode ConnectionStatusCode, reason string) error {
    // TODO perhaps add a timeout here?
    return deviceConn.connection.Close(statusCode, reason)
}

/*
Duration How many seconds the connection has been active
*/
func (deviceConn *DeviceConnection) Duration() (int64, error) {
    if deviceConn.createdAt == 0{
        return -1, errors.New("Connection has not been initiated")
    }

    return time.Now().Unix() - deviceConn.createdAt, nil
}
