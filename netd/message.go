package netd

// Message defines a struct that details a specific message piece of a data
// recieved.
type Message struct {
	Command []byte `json:"command"`
	Data    []byte `json:"data"`
}

// MessageParser defines an interface for a message parser which handles
// parsing of a recieved message slice.
type MessageParser interface {
	Parse([]byte) ([]Message, error)
}
