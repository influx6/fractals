package netd

import (
	"net"
	"sync"
)

// Provider defines a interface for a connection handler, which ensures
// to manage the request-response cycle of a provided net.Conn.
type Provider interface {
	BaseInfo() BaseInfo
	Close(context interface{}) error
	SendMessage(context interface{}, msg []byte, flush bool) error
	CloseNotify() chan struct{}
}

// Broadcast defines an interface for sending messages to two classes of
// listeners, which are clients and clusters. This allows a flexible system for
// expanding more details from a central controller or within a decentral
// controller.
type Broadcast interface {
	SendToClients(context interface{}, msg []byte, flush bool) error
	SendToClusters(context interface{}, msg []byte, flush bool) error
}

// SearchableInfo defines a BaseInfo slice which allows querying specific data
// from giving info.
type SearchableInfo []BaseInfo

// GetInfosByIP searches if the giving address and port exists within the info list
// returning the info that matches it.
func (s SearchableInfo) GetInfosByIP(ip string) ([]BaseInfo, error) {
	var infos []BaseInfo

	for _, info := range s {
		if info.IP != ip {
			continue
		}

		infos = append(infos, info)
	}

	return infos, nil
}

// GetAddr searches if the giving address and port exists within the info list
// returning the info that matches it.
func (s SearchableInfo) HasAddr(addr string, port int) (BaseInfo, error) {
	var info BaseInfo

	for _, info = range s {
		if info.Addr == addr || info.Port == port {
			break
		}
	}

	return info, nil
}

// HasInfo returns true if the info exists within the lists.
func (s SearchableInfo) HasInfo(target BaseInfo) bool {
	for _, info := range s {
		if info.Addr == target.Addr && info.Port == target.Port {
			return true
		}
	}

	return false
}

// Connections provides a interfae which lists connected clients and clusters.
type Connections interface {
	Clients(context interface{}) SearchableInfo
	OnClientConnect(fn func(Provider))
	OnClientDisconnect(fn func(Provider))

	Clusters(context interface{}) SearchableInfo
	OnClusterConnect(fn func(Provider))
	OnClusterDisconnect(fn func(Provider))
}

// Connection defines a struct which stores the incoming request for a
// connection.
type Connection struct {
	net.Conn
	cw             sync.WaitGroup
	Config         Config
	ServerInfo     BaseInfo
	ConnectionInfo BaseInfo
	Connections    Connections
	BroadCaster    Broadcast
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
