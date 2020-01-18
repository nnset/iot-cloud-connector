package servers

import(
    "net"
    "net/http"
    "bytes"
    "runtime"
    "strconv"
    "context"
    
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
    port                    string
    address                 string
    network                 string
    certFile                string
    keyFile                 string
    log                     *logrus.Logger 
    handleMessage           func (ctx context.Context, c *websocket.Conn) error
    authenticateConnection  func (r *http.Request) error
    handleLostConnection    func (connection *connections.DeviceConnection, err error) error

    httpServer              *http.Server
    cloudConnector          *CloudConnector
    shutdownChannel         chan bool
    shutdownIsCompleteChannel chan bool
}

/*
WebsocketCloudConnectorServerSettings Attributes required to start the server
*/
type WebsocketCloudConnectorServerSettings struct {
    Address                 string
    Port                    string
    Network                 string
    CertFile                string
    KeyFile                 string
    Log                     *logrus.Logger 
    HandleMessage           func (ctx context.Context, c *websocket.Conn) error
    AuthenticateConnection  func (r *http.Request) error
    HandleLostConnection    func (connection *connections.DeviceConnection, err error) error
}

/*
Name A brief description for this server
*/
func Name() string {
    return "Websocket server using nhooyr.io/websocket"
}

/*
Init Sets the initial attributes values in order to start the server
*/
func (server *WebsocketCloudConnectorServer) Init(settings WebsocketCloudConnectorServerSettings) {
    if server.id == "" {
        server.id = uuid.New().String()        
        server.port = settings.Port
        server.address = settings.Address
        server.network = settings.Network
        server.certFile = settings.CertFile
        server.keyFile = settings.KeyFile    
        server.log = settings.Log
        server.handleMessage = settings.HandleMessage
        server.authenticateConnection = settings.AuthenticateConnection
        server.handleLostConnection = settings.HandleLostConnection
        server.shutdownIsCompleteChannel = make(chan bool, 1)
    }
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

    wsConnection := connections.WebsocketConnection {c}
    var connection connections.DeviceConnection    
    err = connection.Init(&wsConnection, deviceID, r.RemoteAddr, r.UserAgent())
    
    if err != nil {    
        return
    }

    connectionID, err := server.cloudConnector.AddConnection(&connection)
    
    if err != nil {
        server.log.Error(err)
        return
    }

    for {
        err = server.handleMessage(r.Context(), c)

        server.cloudConnector.MessageReceived(connectionID)

        if websocket.CloseStatus(err) == websocket.StatusNormalClosure {            
            server.log.Debug("Handling message returned StatusNormalClosure", r.RemoteAddr, err)

            server.handleLostConnection(&connection, err)
            server.cloudConnector.CloseConnection(connectionID, connections.StatusNormalClosure, "Connection closed from client")
    
            return
        }

        if err != nil {
            server.log.Error("Failed while handling message ", r.RemoteAddr, err)

            server.handleLostConnection(&connection, err)
            server.cloudConnector.CloseConnection(connectionID, connections.StatusNormalClosure, err.Error())

            return
        }

        connection.MessageReceived()
    }
}

func (server *WebsocketCloudConnectorServer) goRoutineID() uint64 {
    b := make([]byte, 64)
    b = b[:runtime.Stack(b, false)]
    b = bytes.TrimPrefix(b, []byte("goroutine "))
    b = b[:bytes.IndexByte(b, ' ')]
    n, _ := strconv.ParseUint(string(b), 10, 64)

    return n
}

/*
Shutdown Shutsdown the server and notifies shutdownConfirmation channel once finished
*/
func (server *WebsocketCloudConnectorServer) Shutdown(shutdownConfirmation *chan bool) error {
    server.log.Debug("WebsocketCloudConnectorServer is shutting down. Closing active connections.")

    server.cloudConnector.CloseAllConnections("Server shutdown")
    
    server.shutdownIsCompleteChannel <- true
    *shutdownConfirmation <- true

    return nil
}