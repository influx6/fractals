package netd

import "time"

const (
	// VERSION is the current version for the server.
	VERSION = "0.0.1"

	// DEFAULT_PORT is the deault port for client connections.
	DEFAULT_PORT = 3508

	// RANDOM_PORT is the value for port that, when supplied, will cause the
	// server to listen on a randomly-chosen available port. The resolved port
	// is available via the Addr() method.
	RANDOM_PORT = -1

	// MIN_DATA_WRITE_SIZE defines the minimum buffer writer size to be recieved by
	// the connection readers.
	MIN_DATA_WRITE_SIZE = 512

	// MAX_Data_WRITE_SIZE defines the maximum buffer writer size and data size to be
	// allowed on the connection
	MAX_DATA_WRITE_SIZE = 6048

	// DEFAULT_FLUSH_DEADLINE is the write/flush deadlines.
	DEFAULT_FLUSH_DEADLINE = 2 * time.Second

	// ACCEPT_MIN_SLEEP is the minimum acceptable sleep times on temporary errors.
	ACCEPT_MIN_SLEEP = 10 * time.Millisecond

	// MAX_CONTROL_LINE_SIZE is the maximum allowed protocol control line size.
	// 1k should be plenty since payloads sans connect string are separate
	MAX_CONTROL_LINE_SIZE = 1024

	// ACCEPT_MAX_SLEEP is the maximum acceptable sleep times on temporary errors
	ACCEPT_MAX_SLEEP = 1 * time.Second

	// MAX_PAYLOAD_SIZE is the maximum allowed payload size. Should be using
	// something different if > 1MB payloads are needed.
	MAX_PAYLOAD_SIZE = (1024 * 1024)

	// MAX_PENDING_SIZE is the maximum outbound size (in bytes) per client.
	MAX_PENDING_SIZE = (10 * 1024 * 1024)

	// DEFAULT_MAX_CONNECTIONS is the default maximum connections allowed.
	DEFAULT_MAX_CONNECTIONS = (64 * 1024)

	// TLS_TIMEOUT is the TLS wait time.
	TLS_TIMEOUT = float64(500*time.Millisecond) / float64(time.Second)

	// AUTH_TIMEOUT is the authorization wait time.
	AUTH_TIMEOUT = float64(2*TLS_TIMEOUT) / float64(time.Second)

	// DEFAULT_PING_INTERVAL is how often pings are sent to clients and routes.
	DEFAULT_PING_INTERVAL = 2 * time.Minute

	// DEFAULT_PING_MAX_OUT is maximum allowed pings outstanding before disconnect.
	DEFAULT_PING_MAX_OUT = 2
)

var (
	emptyString = []byte("")
	newLine     = []byte("\n")
	endTrace    = []byte("End Trace")
)
