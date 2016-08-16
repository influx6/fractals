package netd

import (
	"bufio"
	"fmt"
	"sync"
	"time"
)

// BaseProvider creates a base provider structure for use in writing handlers
// for connections.
type BaseProvider struct {
	*Connection
	running        bool
	Closer         chan struct{}
	ProviderLock   sync.Mutex
	ProviderWriter *bufio.Writer
}

// NewBaseProvider returns a new instance of a BaseProvider.
func NewBaseProvider(conn *Connection) *BaseProvider {
	var bp BaseProvider
	bp.Connection = conn
	return &bp
}

// Init initializes the base provider and its internal management system.
func (bp *BaseProvider) Init(context interface{}) {
	bp.Closer = make(chan struct{}, 0)
	bp.running = true

	bp.ProviderLock.Lock()
	bp.ProviderWriter = bufio.NewWriterSize(bp.Conn, MIN_DATA_WRITE_SIZE)
	bp.ProviderLock.Unlock()
}

// IsRunning returns true/false if the base provider is still running.
func (bp *BaseProvider) IsRunning() bool {
	var done bool

	bp.ProviderLock.Lock()
	done = bp.running
	bp.ProviderLock.Unlock()

	return done
}

// SendMessage sends a message into the provider connection. This exists for
// the outside which wishes to call a write into the connection.
func (bp *BaseProvider) SendMessage(context interface{}, msg []byte, doFlush bool) error {
	if len(msg) > MAX_PAYLOAD_SIZE {
		return fmt.Errorf("Data is above allowed payload size of %d", MAX_PAYLOAD_SIZE)
	}

	var err error
	if bp.ProviderWriter != nil && bp.Connection != nil && bp.Connection.Conn != nil {
		var deadlineSet bool

		if bp.ProviderWriter.Available() < len(msg) {
			bp.Conn.SetWriteDeadline(time.Now().Add(DEFAULT_FLUSH_DEADLINE))
			deadlineSet = true
		}

		_, err = bp.ProviderWriter.Write(msg)
		if err == nil && doFlush {
			err = bp.ProviderWriter.Flush()
		}

		if deadlineSet {
			bp.Conn.SetWriteDeadline(time.Time{})
		}
	}

	return err
}

// BaseInfo returns a BaseInfo struct which contains information on the
// connection.
func (bp *BaseProvider) BaseInfo() BaseInfo {
	var info BaseInfo

	bp.ProviderLock.Lock()
	info = bp.Connection.ConnectionInfo
	bp.ProviderLock.Unlock()

	return info
}

// CloseNoify returns a chan which allows notification of a close state of
// the base provider.
func (bp *BaseProvider) CloseNotify() chan struct{} {
	return bp.Closer
}

// Close ends the loop cycle for the baseProvider.
func (bp *BaseProvider) Close(context interface{}) error {
	bp.ProviderLock.Lock()
	bp.running = false
	bp.ProviderLock.Unlock()
	return nil
}

// ReadLoop provides a means of intersecting with the looping mechanism
// for a BaseProvider, its an optional mechanism to provide a callback
// like state of behaviour for the way the loop works.
func (bp *BaseProvider) ReadLoop(context interface{}, loopFn func(*BaseProvider)) {
	{
		for bp.running {
			loopFn(bp)
		}
	}

	bp.ProviderLock.Lock()
	close(bp.Closer)
	bp.ProviderLock.Unlock()
}
