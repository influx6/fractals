package netd

import (
	"bytes"
	"crypto/tls"
	"net"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/pborman/uuid"
)

// TCPConn defines a baselevel connection wrapper which provides a flexibile
// tcp request management routine.
type TCPConn struct {
	Stat

	mc     sync.Mutex
	config Config
	sid    string

	infoTCP     BaseInfo
	infoCluster BaseInfo

	tcpClient  net.Listener
	tcpCluster net.Listener

	clients  []Provider
	clusters []Provider

	onConnects    []func(Provider)
	onDisconnects []func(Provider)

	onClusterConnects    []func(Provider)
	onClusterDisconnects []func(Provider)

	runningClient  bool
	runningCluster bool

	closer chan struct{}
	conWG  sync.WaitGroup // waitgroup for incoming connections.
	opWG   sync.WaitGroup // waitgroup for internal servers (client and cluster)
}

// TCP returns a new instance of connection provider.
func TCP(c Config) *TCPConn {
	c.InitLogAndTrace()

	if err := c.ParseTLS(); err != nil {
		c.Log.Error("netd.TCP", "TCP", err, "Error parsing tls arguments")
		panic(err)
	}

	sid := uuid.New()

	var info BaseInfo
	info.Addr = c.Addr
	info.Port = c.Port
	info.Version = VERSION
	info.GoVersion = runtime.Version()
	info.ServerID = sid

	var cinfo BaseInfo
	cinfo.Addr = c.ClustersAddr
	cinfo.Port = c.ClustersPort
	cinfo.Version = VERSION
	cinfo.GoVersion = runtime.Version()
	cinfo.ServerID = sid

	var cn TCPConn
	cn.sid = sid
	cn.infoTCP = info
	cn.infoCluster = cinfo
	cn.config = c

	return &cn
}

// Clients returns the list of available client connections.
func (c *TCPConn) Clients(context interface{}) SearchableInfo {
	var infoList []BaseInfo

	c.mc.Lock()
	for _, client := range c.clients {
		infoList = append(infoList, client.BaseInfo())
	}
	c.mc.Unlock()

	return SearchableInfo(infoList)
}

// OnClientDisonnect adds a function to be called on a client connection disconnect.
func (c *TCPConn) OnClientDisconnect(fn func(Provider)) {
	c.mc.Lock()
	c.onDisconnects = append(c.onDisconnects, fn)
	c.mc.Unlock()
}

// OnClientConnect adds a function to be called on a new client connection.
func (c *TCPConn) OnClientConnect(fn func(Provider)) {
	c.mc.Lock()
	c.onConnects = append(c.onConnects, fn)
	c.mc.Unlock()
}

func (c *TCPConn) callClientConnects(p Provider) {
	c.mc.Lock()
	for _, cnFN := range c.onConnects {
		cnFN(p)
	}
	c.mc.Unlock()
}

func (c *TCPConn) callClientDisconnects(p Provider) {
	c.mc.Lock()
	for _, cnFN := range c.onDisconnects {
		cnFN(p)
	}
	c.mc.Unlock()
}

// Clusters returns a list of available clusters connections.
func (c *TCPConn) Clusters(context interface{}) SearchableInfo {
	var infoList []BaseInfo

	c.mc.Lock()
	for _, cluster := range c.clusters {
		infoList = append(infoList, cluster.BaseInfo())
	}
	c.mc.Unlock()

	return SearchableInfo(infoList)
}

// OnClusterConnect adds a function to be called on a new connection.
func (c *TCPConn) OnClusterConnect(fn func(Provider)) {
	c.mc.Lock()
	c.onClusterConnects = append(c.onClusterConnects, fn)
	c.mc.Unlock()
}

// OnClusterDisonnect adds a function to be called on a connection disconnect.
func (c *TCPConn) OnClusterDisconnect(fn func(Provider)) {
	c.mc.Lock()
	c.onClusterDisconnects = append(c.onClusterDisconnects, fn)
	c.mc.Unlock()
}

func (c *TCPConn) callClusterDisconnects(p Provider) {
	c.mc.Lock()
	for _, cnFN := range c.onClusterDisconnects {
		cnFN(p)
	}
	c.mc.Unlock()
}

func (c *TCPConn) callClusterConnects(p Provider) {
	c.mc.Lock()
	for _, cnFN := range c.onClusterConnects {
		cnFN(p)
	}
	c.mc.Unlock()
}

// SendToClusters sends the provided message to all clusters.
func (c *TCPConn) SendToClusters(context interface{}, msg []byte, flush bool) error {
	c.config.Log.Log(context, "SendToCluster", "Started : Data[%+s]", msg)

	c.mc.Lock()
	defer c.mc.Unlock()

	for _, cluster := range c.clusters {

		var b [][]byte
		b = append(b, []byte("Trace: SendToClients"))
		b = append(b, newLine)
		b = append(b, []byte("Cluster: "))
		b = append(b, []byte(c.infoTCP.String()))
		b = append(b, newLine)
		b = append(b, []byte("ToCluster: "))
		b = append(b, []byte(cluster.BaseInfo().String()))
		b = append(b, newLine)
		b = append(b, []byte("Data: "))
		b = append(b, msg)
		b = append(b, newLine)
		c.config.Trace.Trace(context, bytes.Join(b, emptyString))

		if err := cluster.SendMessage(context, msg, flush); err != nil {
			c.config.Log.Error(context, "SendToCluster", err, "Failed to deliver to cluster : Cluster[%s]", cluster.BaseInfo().String())
		}

		c.config.Trace.Trace(context, endTrace)
	}

	c.config.Log.Log(context, "SendToCluster", "Completed")
	return nil
}

// SendToClusters sends the provided message to all clients.
func (c *TCPConn) SendToClients(context interface{}, msg []byte, flush bool) error {
	c.config.Log.Log(context, "SendToClient", "Started : Data[%+s]", msg)

	c.mc.Lock()
	defer c.mc.Unlock()

	for _, client := range c.clients {

		var b [][]byte
		b = append(b, []byte("Trace: SendToClients"))
		b = append(b, newLine)
		b = append(b, []byte("Cluster: "))
		b = append(b, []byte(c.infoTCP.String()))
		b = append(b, newLine)
		b = append(b, []byte("ToClient: "))
		b = append(b, []byte(client.BaseInfo().String()))
		b = append(b, newLine)
		b = append(b, []byte("Data: "))
		b = append(b, msg)
		b = append(b, newLine)
		c.config.Trace.Trace(context, bytes.Join(b, emptyString))

		if err := client.SendMessage(context, msg, flush); err != nil {
			c.config.Log.Error(context, "SendToClient", err, "Failed to deliver to client : ClientInfo[%s]", client.BaseInfo().String())
		}

		c.config.Trace.Trace(context, endTrace)
	}

	c.config.Log.Log(context, "SendToClient", "Completed")
	return nil
}

// Close ends the tcp connection handler and its internal clusters and clients.
func (c *TCPConn) Close(context interface{}) error {
	if !c.IsRunning() {
		return nil
	}

	c.mc.Lock()
	c.runningClient = false
	c.runningCluster = false
	c.mc.Unlock()

	c.opWG.Wait()

	c.mc.Lock()

	if c.tcpClient != nil {
		if err := c.tcpClient.Close(); err != nil {
			c.config.Log.Error(context, "Close", err, "Completed")
			c.mc.Unlock()
			return err
		}
	}

	if c.tcpCluster != nil {
		if err := c.tcpCluster.Close(); err != nil {
			c.config.Log.Error(context, "Close", err, "Completed")
			c.mc.Unlock()
			return err
		}
	}

	c.mc.Unlock()

	for _, client := range c.clients {
		if err := client.Close("tcp.Close"); err != nil {
			c.config.Log.Error(context, "Close", err, "Failed To Close Client")
		}
	}

	for _, cluster := range c.clusters {
		if err := cluster.Close("tcp.Close"); err != nil {
			c.config.Log.Error(context, "Close", err, "Failed To Close Cluster")
		}
	}

	return nil
}

// IsRunning returns true/false if the connection is up.
func (c *TCPConn) IsRunning() bool {
	var state bool
	c.mc.Lock()
	state = c.runningClient || c.runningCluster
	c.mc.Unlock()
	return state
}

// ServeClusters runs to create the listener for listening to cluster based
// requests for the tcp connection.
func (c *TCPConn) ServeClusters(context interface{}, h Handler) error {
	c.config.Log.Log(context, "tcp.ServeCluster", "Started : Initializing cluster service : Addr[%s] : Port[%d]", c.config.ClustersAddr, c.config.ClustersPort)
	addr := net.JoinHostPort(c.config.ClustersAddr, strconv.Itoa(c.config.ClustersPort))

	var err error
	c.mc.Lock()

	if c.runningCluster {
		c.config.Log.Log(context, "tcp.ServeCluster", "Completed")
		c.mc.Unlock()
		return nil
	}

	c.tcpCluster, err = net.Listen("tcp", addr)
	if err != nil {
		c.config.Log.Error(context, "tcp.ServeCluster", err, "Completed")
		c.mc.Unlock()
		return err
	}

	ip, port, _ := net.SplitHostPort(c.tcpCluster.Addr().String())
	iport, _ := strconv.Atoi(port)

	var info BaseInfo
	info.IP = ip
	info.Port = iport
	info.Version = VERSION
	info.MaxPayload = MAX_PAYLOAD_SIZE
	info.GoVersion = runtime.Version()
	info.ServerID = c.sid

	c.mc.Unlock()

	go c.clusterLoop(context, h, info)

	c.config.Log.Log(context, "tcp.ServeCluster", "Completed")
	return nil
}

// ServeClients runs to create the listener for listening to client based
// requests for the tcp connection.
func (c *TCPConn) ServeClients(context interface{}, h Handler) error {
	c.config.Log.Log(context, "tcp.ServeClients", "Started : Initializing client service : Addr[%s] : Port[%d]", c.config.Addr, c.config.Port)
	addr := net.JoinHostPort(c.config.Addr, strconv.Itoa(c.config.Port))

	var err error
	c.mc.Lock()

	if c.runningClient {
		c.config.Log.Log(context, "tcp.ServeClients", "Completed")
		c.mc.Unlock()
		return nil
	}

	c.tcpClient, err = net.Listen("tcp", addr)
	if err != nil {
		c.config.Log.Error(context, "tcp.ServeClients", err, "Completed")
		c.mc.Unlock()
		return err
	}

	ip, port, _ := net.SplitHostPort(c.tcpClient.Addr().String())
	iport, _ := strconv.Atoi(port)

	var info BaseInfo
	info.IP = ip
	info.Port = iport
	info.Version = VERSION
	info.MaxPayload = MAX_PAYLOAD_SIZE
	info.GoVersion = runtime.Version()
	info.ServerID = c.sid

	c.mc.Unlock()

	go c.clientLoop(context, h, info)

	c.config.Log.Log(context, "tcp.ServeClients", "Completed")
	return nil
}

func (c *TCPConn) clusterLoop(context interface{}, h Handler, info BaseInfo) {
	c.config.Log.Log(context, "tcp.clusterLoop", "Started")

	var stat StatProvider

	// Collect needed state and flag variables.
	c.mc.Lock()
	stat = c.Stat
	config := c.config
	useTLS := c.config.UseTLS
	c.mc.Unlock()

	c.mc.Lock()
	c.opWG.Add(1)
	defer c.opWG.Done()
	c.mc.Unlock()

	sleepTime := ACCEPT_MIN_SLEEP

	{
		for c.IsRunning() {

			conn, err := c.tcpCluster.Accept()
			if err != nil {
				config.Log.Error(context, "tcp.clusterLoop", err, "Accept Error")
				if tmpError, ok := err.(net.Error); ok && tmpError.Temporary() {
					config.Log.Log(context, "tcp.clusterLoop", "Temporary error recieved, sleeping for %dms", sleepTime/time.Millisecond)
					time.Sleep(sleepTime)
					sleepTime *= 2
					if sleepTime > ACCEPT_MAX_SLEEP {
						sleepTime = ACCEPT_MIN_SLEEP
					}
				}

				continue
			}

			sleepTime = ACCEPT_MIN_SLEEP
			config.Log.Log(context, "tcp.clusterLoop", " New Connection : Addr[%a]", conn.RemoteAddr().String())

			var connection Connection

			addr, port, _ := net.SplitHostPort(conn.RemoteAddr().String())
			iport, _ := strconv.Atoi(port)

			var connInfo BaseInfo
			connInfo.Addr = addr
			connInfo.Port = iport
			connInfo.GoVersion = runtime.Version()
			connInfo.MaxPayload = MAX_PAYLOAD_SIZE
			connInfo.ServerID = uuid.New()
			connInfo.Version = VERSION

			// Check if we are required to be using TLS then try to wrap net.Conn
			// to tls.Conn.
			if useTLS {

				tlsConn := tls.Server(conn, config.TLSConfig)
				ttl := secondsToDuration(TLS_TIMEOUT * float64(time.Second))

				var tlsPassed bool

				time.AfterFunc(ttl, func() {
					config.Log.Log(context, "tcp.clusterLoop", "Connection TLS Handshake Timeout : Status[%s] : Addr[%a]", tlsPassed, conn.RemoteAddr().String())

					// Once the time has elapsed, close the connection and nil out.
					if !tlsPassed {
						tlsConn.SetReadDeadline(time.Time{})
						tlsConn.Close()
					}
				})

				tlsConn.SetReadDeadline(time.Now().Add(ttl))

				if err := tlsConn.Handshake(); err != nil {
					config.Log.Error(context, "tcp.clusterLoop", err, " New Connection : Addr[%a] : Failed Handshake", conn.RemoteAddr().String())
					tlsConn.SetReadDeadline(time.Time{})
					tlsConn.Close()
					continue
				}

				connection = Connection{
					Conn:           tlsConn,
					Config:         config,
					ServerInfo:     info,
					ConnectionInfo: connInfo,
					BroadCaster:    c,
					Connections:    c,
					Stat:           stat,
				}

			} else {

				connection = Connection{
					Conn:           conn,
					Config:         config,
					ServerInfo:     info,
					ConnectionInfo: connInfo,
					BroadCaster:    c,
					Connections:    c,
					Stat:           stat,
				}

			}

			provider, err := h(context, &connection)
			if err != nil {
				config.Log.Error(context, "tcp.clusterLoop", err, " New Connection : Addr[%a] : Failed Provider Creation", conn.RemoteAddr().String())
				connection.SetReadDeadline(time.Time{})
				connection.Close()
			}

			// Check authentication of provider and certify if we are authorized.
			if config.Authenticate {
				providerAuth, ok := provider.(ClientAuth)
				if !ok && c.config.MustAuthenticate {
					config.Log.Error(context, "tcp.clusterLoop", err, " New Connection : Addr[%a] : Provider does not match ClientAuth interface", conn.RemoteAddr().String())
					provider.SendMessage(context, []byte("Error: Provider has no authentication. Authentication needed"), true)
					provider.Close(context)
					continue
				}

				if !config.ClusterAuth.Authenticate(providerAuth) {
					if config.MatchClusterCredentials(providerAuth.Credentials()) {
						c.mc.Lock()
						c.clients = append(c.clients, provider)
						c.mc.Unlock()
						continue
					}

					config.Log.Error(context, "tcp.clusterLoop", err, " New Connection : Addr[%a] : Provider does not match ClientAuth interface", conn.RemoteAddr().String())
					provider.SendMessage(context, []byte("Error: Authentication failed"), true)
					provider.Close(context)
					continue
				}
			}

			// Listen for the end signal and descrease connection wait group.
			go func() {
				<-provider.CloseNotify()
				c.conWG.Done()
				c.callClusterDisconnects(provider)
			}()

			c.mc.Lock()
			c.clusters = append(c.clusters, provider)
			c.mc.Unlock()

			c.callClusterConnects(provider)

			continue
		}

	}

	c.config.Log.Log(context, "tcp.clusterLoop", "Completed")
}

func (c *TCPConn) clientLoop(context interface{}, h Handler, info BaseInfo) {
	c.config.Log.Log(context, "tcp.clientLoop", "Started")

	var stat StatProvider

	c.mc.Lock()
	stat = c.Stat
	config := c.config
	useTLS := c.config.UseTLS
	c.mc.Unlock()

	c.mc.Lock()
	c.opWG.Add(1)
	defer c.opWG.Done()
	c.mc.Unlock()

	sleepTime := ACCEPT_MIN_SLEEP

	{
		for c.IsRunning() {

			conn, err := c.tcpClient.Accept()
			if err != nil {
				config.Log.Error(context, "tcp.clientLoop", err, "Accept Error")
				if tmpError, ok := err.(net.Error); ok && tmpError.Temporary() {
					config.Log.Log(context, "clientLoop", "Temporary error recieved, sleeping for %dms", sleepTime/time.Millisecond)
					time.Sleep(sleepTime)
					sleepTime *= 2
					if sleepTime > ACCEPT_MAX_SLEEP {
						sleepTime = ACCEPT_MIN_SLEEP
					}
				}

				continue
			}

			sleepTime = ACCEPT_MIN_SLEEP
			config.Log.Log(context, "tcp.clientLoop", " New Connection : Addr[%a]", conn.RemoteAddr().String())

			var connection Connection

			addr, port, _ := net.SplitHostPort(conn.RemoteAddr().String())
			iport, _ := strconv.Atoi(port)

			var connInfo BaseInfo
			connInfo.Addr = addr
			connInfo.Port = iport
			connInfo.GoVersion = runtime.Version()
			connInfo.MaxPayload = MAX_PAYLOAD_SIZE
			connInfo.ServerID = uuid.New()
			connInfo.Version = VERSION

			// Check if we are required to be using TLS then try to wrap net.Conn
			// to tls.Conn.
			if useTLS {

				tlsConn := tls.Server(conn, config.TLSConfig)
				ttl := secondsToDuration(TLS_TIMEOUT * float64(time.Second))

				var tlsPassed bool

				time.AfterFunc(ttl, func() {
					config.Log.Log(context, "tcp.clientLoop", "Connection TLS Handshake Timeout : Status[%s] : Addr[%a]", tlsPassed, conn.RemoteAddr().String())

					// Once the time has elapsed, close the connection and nil out.
					if !tlsPassed {
						tlsConn.SetReadDeadline(time.Time{})
						tlsConn.Close()
					}
				})

				tlsConn.SetReadDeadline(time.Now().Add(ttl))

				if err := tlsConn.Handshake(); err != nil {
					config.Log.Error(context, "tcp.clientLoop", err, " New Connection : Addr[%a] : Failed Handshake", conn.RemoteAddr().String())
					tlsConn.SetReadDeadline(time.Time{})
					tlsConn.Close()
					continue
				}

				connection = Connection{
					Conn:           tlsConn,
					Config:         config,
					ServerInfo:     info,
					ConnectionInfo: connInfo,
					BroadCaster:    c,
					Connections:    c,
					Stat:           stat,
				}

			} else {

				connection = Connection{
					Conn:           conn,
					Config:         config,
					ServerInfo:     info,
					ConnectionInfo: connInfo,
					BroadCaster:    c,
					Connections:    c,
					Stat:           stat,
				}

			}

			provider, err := h(context, &connection)
			if err != nil {
				config.Log.Error(context, "tcp.clientLoop", err, " New Connection : Addr[%a] : Failed Provider Creation", conn.RemoteAddr().String())
				connection.SetReadDeadline(time.Time{})
				connection.Close()
			}

			// Check authentication of provider and certify if we are authorized.
			if config.Authenticate {
				providerAuth, ok := provider.(ClientAuth)
				if !ok && c.config.MustAuthenticate {
					config.Log.Error(context, "tcp.clientLoop", err, " New Connection : Addr[%a] : Provider does not match ClientAuth interface", conn.RemoteAddr().String())
					provider.SendMessage(context, []byte("Error: Provider has no authentication. Authentication needed"), true)
					provider.Close(context)
					continue
				}

				if !config.ClientAuth.Authenticate(providerAuth) {
					if config.MatchClientCredentials(providerAuth.Credentials()) {
						c.mc.Lock()
						c.clients = append(c.clients, provider)
						c.mc.Unlock()
						continue
					}

					config.Log.Error(context, "tcp.clientLoop", err, " New Connection : Addr[%a] : Provider does not match ClientAuth interface", conn.RemoteAddr().String())
					provider.SendMessage(context, []byte("Error: Authentication failed"), true)
					provider.Close(context)
					continue
				}
			}

			// Listen for the end signal and descrease connection wait group.
			go func() {
				<-provider.CloseNotify()
				c.conWG.Done()
				c.callClientDisconnects(provider)
			}()

			c.mc.Lock()
			c.clients = append(c.clients, provider)
			c.mc.Unlock()

			c.callClientConnects(provider)

			continue
		}
	}

	c.config.Log.Log(context, "tcp.clusterLoop", "Completed")
}
