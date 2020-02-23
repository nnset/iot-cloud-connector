package servers

import(
    "net"
    "net/http"
    "bytes"
    "runtime"
    "strconv"
    "context"
    "sync"
    "fmt"
    
    "github.com/nnset/iot-cloud-connector/connections"
    
    "github.com/google/uuid"
    "nhooyr.io/websocket"
    "github.com/sirupsen/logrus"    
)

/*
WebsocketCloudConnectorServer is a ServerInterface implementation using websockets
as communication protocol thanks to nhooyr.io/websocket lib, where you inject 3 methods
in order to: 
    - Handle incoming messages.
    - Connection authentication.
    - Connection lose.
*/
type WebsocketCloudConnectorServer struct {
    id                      string
    address                 string
    port                    string
    network                 string
    certFile                string
    keyFile                 string
    log                     *logrus.Logger 
    
    handleMessage           func (ctx context.Context, c *websocket.Conn) error
    authenticateConnection  func (r *http.Request) error
    handleLostConnection    func (connection *connections.DeviceConnection, err error) error

    httpServer              *http.Server
    zcloudConnector          *CloudConnector
    shutdownChannel         chan bool
    shutdownIsCompleteChannel chan bool
    
    activeConnections      map[string]*connections.DeviceConnection  // Map that controls the established connections
    dataMutex              *sync.Mutex  // Mutex used for modifying this instance's data
}

/*
NewWebsocketServer Returns a new instance of WebsocketCloudConnectorServer
*/
func NewWebsocketServer(
    address, port, network, certFile, keyFile string, 
    log *logrus.Logger,
    handleMessage func (ctx context.Context, c *websocket.Conn) error,
    authenticateConnection func (r *http.Request) error,
    handleLostConnection func (connection *connections.DeviceConnection, err error) error)  *WebsocketCloudConnectorServer {
    
        return &WebsocketCloudConnectorServer {
            id: uuid.New().String(),
            address: address,
            port: port,
            network: network,
            certFile: certFile,
            keyFile: keyFile,
            log: log,
            handleMessage: handleMessage,
            authenticateConnection: authenticateConnection,
            handleLostConnection: handleLostConnection,
        }
}

/*
Name A brief description for this server
*/
func (server *WebsocketCloudConnectorServer) Name() string {
    return "Websocket server using nhooyr.io/websocket"
}

/*
Start Sets the initial values in order to start the server
*/
func (server *WebsocketCloudConnectorServer) Start(cloudConnector *CloudConnector) error {
    server.log.Debug("Server using WebsocketCloudConnectorServer")

    server.cloudConnector = cloudConnector

    portListener, err := net.Listen(server.network, server.address + ":" + server.port)

    if err != nil {
        return err
    }

    defer portListener.Close()

    server.httpServer = &http.Server {
        Handler: http.HandlerFunc(server.handleConnection),
    }

    defer server.httpServer.Close()

    go func() {
        if server.keyFile != "" {
            server.log.Debug("Serving websockets via wss (TLS) at " + server.address + ":" + server.port)
            err = server.httpServer.ServeTLS(portListener, server.certFile, server.keyFile)
        } else {
            server.log.Debug("Serving websockets via ws (TLS OFF) at " + server.address + ":" + server.port)
            err = server.httpServer.Serve(portListener)
        }
        
        if err != http.ErrServerClosed {
            server.log.Error("Server closed ", err)
            return
        }
    }()
    
    <-server.shutdownIsCompleteChannel

    return nil
}

func (server *WebsocketCloudConnectorServer) handleConnection(w http.ResponseWriter, r *http.Request) {
    server.log.Debugf("Incoming connection, go routine id %d", server.goRoutineID())

    c, err := websocket.Accept(w, r, &websocket.AcceptOptions{})

    if err != nil {
        server.log.Error("Websocket rejected", err)
        return
    }

    server.log.Debugf("Websocket from %s accepted", r.RemoteAddr)

    defer c.Close(websocket.StatusInternalError, "Websocket KO")

    err = server.authenticateConnection(r)

    if err != nil {
        server.log.Debug("Connection rejected. Invalid credentials.")
        return
    }

    deviceID := r.Header.Get("device_id")

    var wsConnection = connections.NewWebsocketConnection(c)
    var deviceConnection = connections.NewDeviceConnection(wsConnection, deviceID, r.RemoteAddr, r.UserAgent())

    server.addConnection(deviceConnection)
    
    for {
        err = server.handleMessage(r.Context(), c)

        server.cloudConnector.MessageReceived(deviceConnection.ID())

        if websocket.CloseStatus(err) == websocket.StatusNormalClosure {            
            server.log.Debug("Handling message returned StatusNormalClosure", r.RemoteAddr, err)

            server.handleLostConnection(deviceConnection, err)
            server.closeConnection(deviceConnection, connections.StatusNormalClosure, "Connection closed from client")
    
            return
        }

        if err != nil {
            server.log.Error("Failed while handling message ", r.RemoteAddr, err)

            server.handleLostConnection(deviceConnection, err)
            server.closeConnection(deviceConnection, connections.StatusNormalClosure, err.Error())

            return
        }
    }
}

func (server *WebsocketCloudConnectorServer) addConnection(connection *connections.DeviceConnection) error {
    server.dataMutex.Lock()
    defer server.dataMutex.Unlock()

    _, alreadyConnected := server.activeConnections[connection.ID()]

    if alreadyConnected {
        return fmt.Errorf(fmt.Sprintf("Connection rejected. Connection #%s with device #%s was already established.", connection.ID(), connection.DeviceID()))
    }

    server.activeConnections[connection.ID()] = connection
    err := server.cloudConnector.ConnectionEstablished(connection)

    if err != nil {
        server.log.Error(err)
        // TODO 
        // if cloud connector rejected the connection but server did accept it, should we cancel the connection?
    }

    return nil
}

func (server *WebsocketCloudConnectorServer) closeConnection(connection *connections.DeviceConnection, statusCode connections.ConnectionStatusCode, reason string) error {
    server.dataMutex.Lock()
    defer server.dataMutex.Unlock()

    connectionID := connection.ID()

    _, exists := server.activeConnections[connection.ID()]

    if !exists {
        return fmt.Errorf(fmt.Sprintf("Connection #%s with device #%s already closed.", connection.ID(), connection.DeviceID()))
    }

    // TODO timeout ?
    err := connection.Close(statusCode, reason)

    if (err != nil) {
        delete(server.activeConnections, connectionID)
        server.cloudConnector.ConnectionClosed(connectionID, statusCode, reason)
    }
    
    return err
}

/*
Shutdown Shutsdown the server and notifies shutdownConfirmation channel once finished
*/
func (server *WebsocketCloudConnectorServer) Shutdown(shutdownConfirmation *chan bool) error {
    server.log.Debug("WebsocketCloudConnectorServer is shutting down. Closing active connections.")

    for _, connection := range server.activeConnections {
        remoteAddress := connection.RemoteAddress()
        userAgent := connection.UserAgent()
        deviceID := connection.DeviceID()
        connectionID := connection.ID()

        err := server.closeConnection(connection, connections.StatusNormalClosure, "Server shutdown")

        if err != nil {
            server.log.Debugf("Unable to close connection #%s with device #%s from %s (%s) : %s", connectionID, deviceID, remoteAddress, userAgent, err)
        }
    }

    server.shutdownIsCompleteChannel <- true
    *shutdownConfirmation <- true

    return nil
}

func (server *WebsocketCloudConnectorServer) goRoutineID() uint64 {
    b := make([]byte, 64)
    b = b[:runtime.Stack(b, false)]
    b = bytes.TrimPrefix(b, []byte("goroutine "))
    b = b[:bytes.IndexByte(b, ' ')]
    n, _ := strconv.ParseUint(string(b), 10, 64)

    return n
}