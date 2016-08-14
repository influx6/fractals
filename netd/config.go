package netd

import (
	"crypto/tls"
	"time"
)

// Trace defines an interface which receives data trace data logs.
type Trace interface {
	Trace(context interface{}, msg []byte)
}

// Log defines an interface which receives logs events/messages.
type Log interface {
	Log(context interface{}, targetFunc string, message string, data ...interface{})
	Error(context interface{}, targetFunc string, err error, message string, data ...interface{})
}

// Crendential defines a struct for storing user authentication crendentials.
type Credential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Config provides a configuration struct which defines specific settings for
// the connection handler.
type Config struct {
	Trace Trace `json:"-"`
	Debug Log   `json:"-"`

	ClientCrendentails []Credential `json:"-"`

	Port int    `json:"port"`
	Addr string `json:"addr"`

	HTTPPort  int    `json:"http_port"`
	HTTPAddr  string `json:"http_addr"`
	HTTPSPort int    `json:"https_port"`
	HTTPSAddr string `json:"https_addr"`

	ClusterCredentials    []Credential `json:"-"`
	ClustersPort          int          `json:"clusters_port"`
	ClustersAddr          string       `json:"clusters_addr"`
	MaxClusterAuthTimeout float64      `json:"max_cluster_auth_timeout"`

	MaxPayload      int64         `json:"max_payload_size"`
	MaxPending      int64         `json:"max_pending_size"`
	MaxConnections  int           `json:"max_connections"`
	MaxPingInterval time.Duration `json:"max_ping_timeout"`
	MaxPingTimeout  float64       `json:"max_ping_timeout"`

	Authenticate bool `json:"authenticate"`

	ClientAuth  Auth `json:"-"`
	ClusterAuth Auth `json:"-"`

	UseTLS        bool        `json:"use_tls"`
	MaxTLSTimeout float64     `json:"max_tls_timeout"`
	TLSKeyFile    string      `json:"-"`
	TLSCertFile   string      `json:"-"`
	TLSCaCertFile string      `json:"-"`
	TLSVerify     bool        `json:"TLSVerify"`
	TLSConfig     *tls.Config `json:"-"`
}

// MatchClientCredentials matches the provided crendential against the
// provided static users crendential, this is useful for testing as it
// allows a predefined set of crendentails to allow.
func (c Config) MatchClientCredentials(cd Credential) bool {
	for _, user := range c.ClientCrendentails {
		if cd.Username == user.Username && cd.Password == user.Password {
			return true
		}
	}

	return false
}

// MatchClusterCredentials matches the provided crendential against the
// provided static cluster users crendential, this is useful for testing as it
// allows a predefined set of crendentails to allow.
func (c Config) MatchClusterCredentials(cd Credential) bool {
	for _, user := range c.ClusterCredentials {
		if cd.Username == user.Username && cd.Password == user.Password {
			return true
		}
	}

	return false
}

// ParseTLS parses the tls configuration variables assigning the value to the
// TLSConfig if not already assigned to.
func (c Config) ParseTLS() error {
	if c.TLSConfig != nil || !c.UseTLS {
		return nil
	}

	var err error
	c.TLSConfig, err = LoadTLS(c.TLSCertFile, c.TLSKeyFile, c.TLSCaCertFile)
	if err != nil {
		return err
	}

	return nil
}