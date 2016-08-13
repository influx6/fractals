package netd

// ClientAuth defines an authentication interface which returns the needed
// crendentials for authenticating client requests.
type ClientAuth interface {
	Credentials() Credential
}

// Auth defines interface for handling authentication of connection requests
// using the Conn mediator.
type Auth interface {
	Authenticate(ClientAuth) bool
}
