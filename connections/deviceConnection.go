package connections

import (
	"errors"
	"time"
)

// DeviceConnection Device's connections information used to keep track on how
// connections are behaving
type DeviceConnection struct {
	connectionID                 string
	deviceID                     string
	deviceType                   string
	userAgent                    string
	remoteAddress                string
	createdAt                    int64
	lastReceivedMessageTimeStamp int64
	lastSentMessageTimeStamp     int64
	receivedMessages             uint
	sentMessages                 uint
}

// NewDeviceConnection Creates a new instance of DeviceConnection
func NewDeviceConnection(connectionID, deviceID, deviceType, userAgent, remoteAddress string) *DeviceConnection {
	return &DeviceConnection{
		connectionID:  connectionID,
		deviceID:      deviceID,
		deviceType:    deviceType,
		userAgent:     userAgent,
		remoteAddress: remoteAddress,
		createdAt:     time.Now().Unix(),
	}
}

// ConnectionID Connection's UUID
func (status *DeviceConnection) ConnectionID() string {
	return status.connectionID
}

// DeviceID Connection's Device UUID
func (status *DeviceConnection) DeviceID() string {
	return status.deviceID
}

// DeviceType Connection's Device type
func (status *DeviceConnection) DeviceType() string {
	return status.deviceType
}

// UserAgent Connection's user agent
func (status *DeviceConnection) UserAgent() string {
	return status.userAgent
}

// RemoteAddress Connection's remote address
func (status *DeviceConnection) RemoteAddress() string {
	return status.remoteAddress
}

// Uptime how many seconds the connection has been active
func (status *DeviceConnection) Uptime() (int64, error) {
	if status.createdAt == 0 {
		return -1, errors.New("Connection has not been established")
	}

	return time.Now().Unix() - status.createdAt, nil
}

// LastReceivedMessageTimeStamp When was the last time when a message was received from the connected IoT device (unix time)
func (status *DeviceConnection) LastReceivedMessageTimeStamp() int64 {
	return status.lastReceivedMessageTimeStamp
}

// LastSentMessageTimeStamp When was the last time when a message was sent to the connected IoT device (unix time)
func (status *DeviceConnection) LastSentMessageTimeStamp() int64 {
	return status.lastSentMessageTimeStamp
}

// ReceivedMessages How many messages have were received from the connected IoT device
func (status *DeviceConnection) ReceivedMessages() uint {
	return status.receivedMessages
}

// SentMessages How many messages were we sent to the connected IoT device
func (status *DeviceConnection) SentMessages() uint {
	return status.sentMessages
}

// MessageSent A message has been sent to the connected IoT device
func (status *DeviceConnection) MessageSent() {
	status.sentMessages++
	status.lastSentMessageTimeStamp = time.Now().Unix()
}

// MessageReceived A message was received from the connected IoT device
func (status *DeviceConnection) MessageReceived() {
	status.receivedMessages++
	status.lastReceivedMessageTimeStamp = time.Now().Unix()
}
