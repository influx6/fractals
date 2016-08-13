package netd

import "net"

// Provider defines a interface for a connection handler, which ensures
// to manage the request-response cycle of a provided net.Conn.
type Provider interface {
	Close(context interface{}) error
	SendMessage(context interface{},msg []byte) error
	CloseNotify() chan struct{}
}

// Broadcast defines an interface for sending messages to two classes of
// listeners, which are clients and clusters. This allows a flexible system for
// expanding more details from a central controller or within a decentral
// controller.
type Broadcast interface {
	SendToClients(context interface{}, msg[]byte) error
	SendToClusters(context interface{},msg []byte) error
}

// StatProvider provides a interfce which allows access to operations on
// stats items.
type StatProvider interface {
  IncrementInMsg()
  IncrementOutMsg()
  IncrementRequest()
  IncrementReads(size int)
  IncrementWrites(size int)
}

// Handler defines a function handler which
type Handler func(context interface{}, net.Conn, Config) (Provider, error)

// Conn defines an interface which manages the connection creation and accept
// lifecycle and using the provided ConnHandler produces connections for
// both clusters and and clients.
type Conn interface {
	Broadcast
	ServeClient(context interface{}, Handler) error
	ServeCluster(context interface{}, Handler) error
}
