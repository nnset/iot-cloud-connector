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
	deviceName                   string
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
func NewDeviceConnection(connectionID, deviceID, deviceName, deviceType, userAgent, remoteAddress string) *DeviceConnection {
	return &DeviceConnection{
		connectionID:  connectionID,
		deviceID:      deviceID,
		deviceName:    deviceID,
		deviceType:    deviceType,
		userAgent:     userAgent,
		remoteAddress: remoteAddress,
		createdAt:     time.Now().Unix(),
	}
}

// ConnectionID Connection's UUID
func (c *DeviceConnection) ConnectionID() string {
	return c.connectionID
}

// DeviceID Connection's Device UUID
func (c *DeviceConnection) DeviceID() string {
	return c.deviceID
}

// DeviceName Connection's Device name
func (c *DeviceConnection) DeviceName() string {
	return c.deviceName
}

// DeviceType Connection's Device type
func (c *DeviceConnection) DeviceType() string {
	return c.deviceType
}

// UserAgent Connection's user agent
func (c *DeviceConnection) UserAgent() string {
	return c.userAgent
}

// RemoteAddress Connection's remote address
func (c *DeviceConnection) RemoteAddress() string {
	return c.remoteAddress
}

// CreatedAt Connection's creation timestamp
func (c *DeviceConnection) CreatedAt() int64 {
	return c.createdAt
}

// Uptime how many seconds the connection has been active
func (c *DeviceConnection) Uptime() (int64, error) {
	if c.createdAt == 0 {
		return -1, errors.New("Connection has not been established")
	}

	return time.Now().Unix() - c.createdAt, nil
}

// LastReceivedMessageTimeStamp When was the last time when a message was received from the connected IoT device (unix time)
func (c *DeviceConnection) LastReceivedMessageTimeStamp() int64 {
	return c.lastReceivedMessageTimeStamp
}

// LastSentMessageTimeStamp When was the last time when a message was sent to the connected IoT device (unix time)
func (c *DeviceConnection) LastSentMessageTimeStamp() int64 {
	return c.lastSentMessageTimeStamp
}

// ReceivedMessages How many messages have were received from the connected IoT device
func (c *DeviceConnection) ReceivedMessages() uint {
	return c.receivedMessages
}

// SentMessages How many messages were we sent to the connected IoT device
func (c *DeviceConnection) SentMessages() uint {
	return c.sentMessages
}

// MessageSent A message has been sent to the connected IoT device
func (c *DeviceConnection) MessageSent() {
	c.sentMessages++
	c.lastSentMessageTimeStamp = time.Now().Unix()
}

// MessageReceived A message was received from the connected IoT device
func (c *DeviceConnection) MessageReceived() {
	c.receivedMessages++
	c.lastReceivedMessageTimeStamp = time.Now().Unix()
}
