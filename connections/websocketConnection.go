package connections

import(
    "nhooyr.io/websocket"
)

/*
WebsocketConnection is a NetworkConnection interface implementation using nhooyr.io/websocket
websockets protocol implementation
*/
type WebsocketConnection struct {
    Conn  *websocket.Conn
}

/*
NewWebsocketConnection Creates a new WebsocketConnection
*/
func NewWebsocketConnection(conn *websocket.Conn) *WebsocketConnection {
    return &WebsocketConnection { Conn: conn }
}

/*
Close Closes the websocket connection
*/
func (wsCon *WebsocketConnection) Close(statusCode ConnectionStatusCode, reason string) error {
    return wsCon.Conn.Close(websocket.StatusCode(statusCode), reason)
}