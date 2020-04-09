package connections

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

/*
DeviceConnection represents a permanent and bidirectional (if you need it to be) connection between
a server and an IoT device, such as an edge server, a sensor or an actuator.
*/
type DeviceConnection struct {
	id            string
	deviceID      string
	deviceType    string
	userAgent     string
	remoteAddress string
	createdAt     int64
	connection    NetworkConnection
}

/*
NetworkConnection defines the methods required for any kind of network connection
with a device
*/
type NetworkConnection interface {
	Close(statusCode ConnectionStatusCode, reason string) error
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
NewDeviceConnection Creates a new instance of DeviceConnection
*/
func NewDeviceConnection(netConn NetworkConnection, deviceID, remoteAddress, userAgent string) *DeviceConnection {
	return &DeviceConnection{
		id:            uuid.New().String(),
		createdAt:     time.Now().Unix(),
		remoteAddress: remoteAddress,
		userAgent:     userAgent,
		connection:    netConn,
		deviceID:      deviceID,
	}
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
DeviceType Connected device's ID
*/
func (deviceConn *DeviceConnection) DeviceType() string {
	return deviceConn.deviceType
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
	if deviceConn.createdAt == 0 {
		return -1, errors.New("Connection has not been initiated")
	}

	return time.Now().Unix() - deviceConn.createdAt, nil
}
