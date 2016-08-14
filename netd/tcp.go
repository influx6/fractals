package netd

import (
	"net"
	"sync"
	"sync/atomic"
)

// TCPConn defines a baselevel connection wrapper which provides a flexibile
// tcp request management routine.
type TCPConn struct {
	Stat

	config         Config
	mc             sync.Mutex
	tcp            net.Listener
	http           net.Listener
	runningClient  bool
	runningCluster bool

	closer chan struct{}
	conWG  sync.WaitGroup // waitgroup for incoming connections.
	opWG   sync.WaitGroup // waitgroup for internal servers (client and cluster)
}

// TCP returns a new instance of connection provider.
func TCP(c Config) *TCPConn {
	if err := c.ParseTLS(); err != nil {
		panic(err)
	}

	var cn TCPConn
	cn.config = c

	return &cn
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
	addr := net.JoinHostPort(c.config.ClustersAddr, c.config.ClustersPort)

	var err error
	c.mc.Lock()

	if c.runningCluster {
		c.mc.Unlock()
		return nil
	}

	c.tcp, err = net.Listen("tcp", addr)
	if err != nil {
		c.mc.Unlock()
		return err
	}

	c.mc.Unlock()

	go c.clusterLoop(h)

	return nil
}

// ServeClients runs to create the listener for listening to client based
// requests for the tcp connection.
func (c *TCPConn) ServeClients(context interface{}, h Handler) error {
	addr := net.JoinHostPort(c.config.Addr, c.config.Port)

	var err error
	c.mc.Lock()

	if c.runningClient {
		c.mc.Unlock()
		return nil
	}

	c.tcp, err = net.Listen("tcp", addr)
	if err != nil {
		c.mc.Unlock()
		return err
	}

	c.mc.Unlock()

	go c.clientLoop(h)
	return nil
}

func (c *TCPConn) clientLoop(context interface{}, h Handler) {

	// Collect needed state and flag variables.
	c.mc.Lock()
	useTLS := c.config.UseTLS
	useAuth := c.config.Authenticate
	c.mc.Unlock()

	c.mc.Lock()
	c.opWG.Add(1)
	c.mc.Unlock()

runLoop:
	{
		for c.IsRunning() {

			conn, err := c.tcp.Accept()
			if err != nil {

			}

		}
	}
}

func (c *TCPConn) clusterLoop(context interface{}, h Handler) {
	c.mc.Lock()
	defer c.mc.Unlock()

	c.mc.Lock()
	useTLS := c.config.UseTLS
	useAuth := c.config.Authenticate
	c.mc.Unlock()

	c.mc.Lock()
	c.opWG.Add(1)
	c.mc.Unlock()
runLoop:
	{
		for c.IsRunning() {

		}
	}
}
