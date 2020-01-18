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
Close Closes the websocket connection
*/
func (wsCon *WebsocketConnection) Close(statusCode ConnectionStatusCode, reason string) error {
    return wsCon.Conn.Close(websocket.StatusCode(statusCode), reason)
}