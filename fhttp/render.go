package fhttp

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
)

// Field defines a struct for collating fields errors that occur.
type Field struct {
	Name     string      `json:"field_name"`
	Value    string      `json:"field_value"`
	Error    string      `json:"field_error"`
	Expected interface{} `json:"expected_value"`
}

// JSONError defines a json error response struct
type JSONError struct {
	Error  string                 `json:"error"`
	Fields []Field                `json:"fields,omitempty"`
	Extras map[string]interface{} `json:"extras,omitempty"`
}

// Param defines the map of values to be handled by the provider.
type Param map[string]string

// Get returns the given value for a specific key and returns a bool to indicate
// if it was found.
func (p Param) Get(key string) (val string, found bool) {
	val, found = p[key]
	return
}

// GetBool returns a bool value if possible from the value of a giving key if it
// exits.
func (p Param) GetBool(key string) (item bool, err error) {
	val, ok := p[key]
	if !ok {
		err = errors.New("Not Found")
		return
	}

	item, err = strconv.ParseBool(val)
	return
}

// GetFloat returns a float value if possible from the value of a giving key if
// it exits.
func (p Param) GetFloat(key string) (item float64, err error) {
	val, ok := p[key]
	if !ok {
		err = errors.New("Not Found")
		return
	}

	item, err = strconv.ParseFloat(val, 64)
	return
}

// GetInt returns a int value if possible from the value of a giving key if it
// exits.
func (p Param) GetInt(key string) (item int, err error) {
	val, ok := p[key]
	if !ok {
		err = errors.New("Not Found")
		return
	}

	item, err = strconv.Atoi(val)
	return
}

// Request defines a response object which holds the request  object
// associated with it and allows you write out the behaviour.
type Request struct {
	Params Param
	Req    *http.Request
	Res    ResponseWriter
}

// Respond renders out a JSON response and status code giving using the Render
// function.
func (r *Request) Respond(code int, data interface{}) {
	Render(code, r.Req, r.Res, data)
}

// RespondAny renders out a JSON response and status code giving using the Render
// function.
func (r *Request) RespondAny(code int, content string, data []byte) {
	RenderAny(code, r.Req, r.Res, content, data)
}

// RespondError renders out a error response into the request object.
func (r *Request) RespondError(code int, err error) {
	RenderErrorWithStatus(code, err, r.Req, r.Res)
}

// RenderAny writes the giving data into the response as JSON.
func RenderAny(code int, r *http.Request, w http.ResponseWriter, content string, data []byte) {
	if code == http.StatusNoContent {
		w.WriteHeader(code)
		return
	}

	w.Header().Set("Content-Type", content)
	w.WriteHeader(code)

	if cb := r.URL.Query().Get("callback"); cb != "" {
		io.WriteString(w, cb+"("+string(data)+")")
		return
	}

	w.Write(data)
}

// Render writes the giving data into the response as JSON.
func Render(code int, r *http.Request, w http.ResponseWriter, data interface{}) {
	if code == http.StatusNoContent {
		w.WriteHeader(code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	jsd, err := json.Marshal(data)
	if err != nil {
		jsd = []byte("{}")
	}

	if cb := r.URL.Query().Get("callback"); cb != "" {
		io.WriteString(w, cb+"("+string(jsd)+")")
		return
	}

	io.WriteString(w, string(jsd))
}

// RenderResponse writes the giving data into the response as JSON to the passed
// ResponseRequest.
func RenderResponse(code int, r *Request, data interface{}) {
	Render(code, r.Req, r.Res, data)
}

// RenderResponseErrorWithStatus renders the giving error as a json response to
// the ResponseRequest.
func RenderResponseErrorWithStatus(status int, err error, r *Request) {
	Render(status, r.Req, r.Res, JSONError{Error: err.Error()})
}

// RenderErrorWithStatus renders the giving error as a json response.
func RenderErrorWithStatus(status int, err error, r *http.Request, w http.ResponseWriter) {
	Render(status, r, w, JSONError{Error: err.Error()})
}

// RenderError renders the giving error as a json response.
func RenderError(err error, r *http.Request, w http.ResponseWriter) {
	Render(http.StatusBadRequest, r, w, JSONError{Error: err.Error()})
}

// RenderResponseError renders the giving error as a json response to the
// passed ResponseRequest object.
func RenderResponseError(err error, r *Request) {
	Render(http.StatusBadRequest, r.Req, r.Res, JSONError{Error: err.Error()})
}

// ResponseWriter is a wrapper around http.ResponseWriter that provides extra information about
// the response. It is recommended that middleware handlers use this construct to wrap a responsewriter
// if the functionality calls for it.
type ResponseWriter interface {
	http.ResponseWriter
	http.Flusher
	// Status returns the status code of the response or 0 if the response has not been written.
	Status() int

	// Written returns whether or not the ResponseWriter has been written.
	StatusWritten() bool

	// DataWritten returns true/false if the response.Write method had been called.
	DataWritten() bool

	// Size returns the size of the response body.
	Size() int
}

// NewResponseWriter creates a ResponseWriter that wraps an http.ResponseWriter
func NewResponseWriter(rw http.ResponseWriter) ResponseWriter {
	return &responseWriter{rw, 0, 0, 0}
}

type responseWriter struct {
	http.ResponseWriter
	status    int
	size      int
	datawrite int64
}

func (rw *responseWriter) WriteHeader(s int) {
	rw.status = s
	rw.ResponseWriter.WriteHeader(s)
}

func (rw *responseWriter) Write(b []byte) (int, error) {

	// The status will be StatusOK if WriteHeader has not been called yet
	if !rw.StatusWritten() {
		rw.WriteHeader(http.StatusOK)
	}

	atomic.StoreInt64(&rw.datawrite, 1)

	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) Size() int {
	return rw.size
}

func (rw *responseWriter) DataWritten() bool {
	return atomic.LoadInt64(&rw.datawrite) > 0
}

func (rw *responseWriter) StatusWritten() bool {
	return rw.status != 0
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (rw *responseWriter) CloseNotify() <-chan bool {
	return rw.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (rw *responseWriter) Flush() {
	flusher, ok := rw.ResponseWriter.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}
