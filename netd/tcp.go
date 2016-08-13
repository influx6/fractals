package netd

import (
	"net"
	"sync"
	"sync/atomic"
)

// Stat defines a struct for storing statistics data recieved from a provided Conn.
type Stat struct {
	InMsg        int64
	OutMsg       int64
	OutBytes     int64
	InBytes      int64
	Requests     int64
	TotalClients int64
}

// IncrementWrites increments the InByte counter.
func (stat Stat) IncrementWrites(size int) {
	atomic.AddInt64(&stat.OutBytes, int64(size))
}

// IncrementOutMsg increments the OutMsg counter.
func (stat Stat) IncrementRead(size int) {
	atomic.AddInt64(&stat.InBytes, int64(size))
}

// IncrementInMsg increments the InMsg counter.
func (stat Stat) IncrementInMsg() {
	atomic.AddInt64(&stat.InMsg, 1)
}

// IncrementOutMsg increments the OutMsg counter.
func (stat Stat) IncrementOutMsg() {
	atomic.AddInt64(&stat.OutMsg, 1)
}

// IncrementRequest increments the Requests counter.
func (stat Stat) IncrementRequest() {
	atomic.AddInt64(&stat.Requests, 1)
}

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
