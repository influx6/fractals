package net

import (
	"crypto/tls"
	"net"
	"sync"
	"time"
)

const (
	// VERSION is the current version for the server.
	VERSION = "0.1.2"

	// DEFAULT_PORT is the deault port for client connections.
	DEFAULT_PORT = 4222

	// RANDOM_PORT is the value for port that, when supplied, will cause the
	// server to listen on a randomly-chosen available port. The resolved port
	// is available via the Addr() method.
	RANDOM_PORT = -1

	// MIN_DATA_SIZE defines the minimum buffer writer size to be recieved by
	// the connection readers.
	MIN_DATA_SIZE = 512

	// MAX_DATA_SIZE defines the maximum buffer writer size and data size to be
	// allowed on the connection
	MAX_DATA_SIZE = 6048

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

// UserCrendential defines a struct for storing user authentication crendentials.
type UserCredentail struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ClientAuth defines an authentication interface which returns the needed
// crendentials for authenticating client requests.
type ClientAuth interface {
	Credentials() interface{}
}

// Auth defines interface for handling authentication of connection requests
// using the Conn mediator.
type Auth interface {
	Authenticate(ClientAuth) bool
}

// Stat defines a struct for storing statistics data recieved from a provided Conn.
type Stat struct {
	InMsg        int64
	OutMsg       int64
	OutBytes     int64
	InBytes      int64
	Requests     int64
	TotalClients int64
}

// Config provides a configuration struct which defines specific settings for
// the connection handler.
type Config struct {
	Trace                 bool             `json:"-"`
	Debug                 bool             `json:"-"`
	Port                  int              `json:"port"`
	Addr                  string           `json:"addr"`
	Users                 []UserCredentail `json:"-"`
	HTTPPort              int              `json:"http_port"`
	HTTPAddr              string           `json:"http_addr"`
	HTTPSPort             int              `json:"https_port"`
	HTTPSAddr             string           `json:"https_addr"`
	ClusterCredentail     UserCredentail   `json:"-"`
	ClustersPort          int              `json:"clusters_port"`
	ClustersAddr          string           `json:"clusters_addr"`
	MaxClusterAuthTimeout float64          `json:"max_cluster_auth_timeout"`
	MaxPayload            int64            `json:"max_payload_size"`
	MaxPending            int64            `json:"max_pending_size"`
	MaxConnections        int              `json:"max_connections"`
	MaxPingInterval       time.Duration    `json:"max_ping_timeout"`
	MaxPingTimeout        float64          `json:"max_ping_timeout"`
	Authenticate          bool             `json:"authenticate"`
	ClientAuth            Auth             `json:"-"`
	RouterAuth            Auth             `json:"-"`
	UseTLS                bool             `json:"use_tls"`
	MaxTLSTimeout         float64          `json:"max_tls_timeout"`
	TLSKeyFile            string           `json:"-"`
	TLSCertFile           string           `json:"-"`
	TLSCaCertFile         string           `json:"-"`
	TLSVerify             bool             `json:"TLSVerify"`
	TLSConfig             *tls.Config      `json:"-"`
}

// TLSConfig holds the parsed tls config information,
// used with flag parsing
type TLSConfig struct {
	CertFile string
	KeyFile  string
	CaFile   string
	Verify   bool
	Timeout  float64
	Ciphers  []uint16
}

// Conn defines a baselevel connection wrapper which provides a flexibile
// tcp request management routine.
type Conn struct {
	Stat

	mc   sync.Mutex
	tcp  net.Listener
	http net.Listener

	waiters sync.WaitGroup
}
