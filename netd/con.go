package netd

import (
	"net"
	"sync"
)

// Provider defines a interface for a connection handler, which ensures
// to manage the request-response cycle of a provided net.Conn.
type Provider interface {
	Close(context interface{}) error
	SendMessage(context interface{}, msg []byte) error
	CloseNotify() chan struct{}
}

// Broadcast defines an interface for sending messages to two classes of
// listeners, which are clients and clusters. This allows a flexible system for
// expanding more details from a central controller or within a decentral
// controller.
type Broadcast interface {
	SendToClients(context interface{}, msg []byte) error
	SendToClusters(context interface{}, msg []byte) error
}

// ConnectionInfo provides a interfae which lists connected clients and clusters.
type ConnectionInfo interface {
	Clusters() []BaseInfo
	Clients() []BaseInfo
}

// SearchableInfo defines a BaseInfo slice which allows querying specific data
// from giving info.
type SearchableInfo []BaseInfo

// GetInfosByIP searches if the giving address and port exists within the info list
// returning the info that matches it.
func (SearchableInfo) GetInfosByIP(ip string) ([]BaseInfo, error) {
	var infos []BaseInfo

	return infos, nil
}

// GetAddr searches if the giving address and port exists within the info list
// returning the info that matches it.
func (SearchableInfo) HasAddr(addr string, port int) (BaseInfo, error) {
	var info BaseInfo

	return info, nil
}

// HasInfo returns true if the info exists within the lists.
func (SearchableInfo) HasInfo(info BaseInfo) bool {

	return true
}

// Connection defines a struct which stores the incoming request for a
// connection.
type Connection struct {
	net.Conn
	cw             sync.WaitGroup
	Config         Config
	ServerInfo     BaseInfo
	ConnectionInfo BaseInfo
	Stat           StatProvider
}

// Handler defines a function handler which returns a new Provider from a
// Connection.
type Handler func(context interface{}, c *Connection) (Provider, error)

// Conn defines an interface which manages the connection creation and accept
// lifecycle and using the provided ConnHandler produces connections for
// both clusters and and clients.
type Conn interface {
	Broadcast
	ServeClient(context interface{}, h Handler) error
	ServeCluster(context interface{}, h Handler) error
}
