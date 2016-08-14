package netd

// StatProvider provides a interfce which allows access to operations on
// stats items.
type StatProvider interface {
  IncrementInMsg()
  IncrementOutMsg()
  IncrementRequest()
  IncrementReads(size int)
  IncrementWrites(size int)
}

// Stat defines a struct for storing statistics data recieved from a provided 
// Conn.
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