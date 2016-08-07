package net

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

//LoadTLS loads a tls.Config from a key and cert file path
func LoadTLS(cert string, key string) (*tls.Config, error) {
	var config *tls.Config
	config.Certificates = make([]tls.Certificate, 1)

	c, err := tls.LoadX509KeyPair(cert, key)

	if err != nil {
		return nil, err
	}

	config.Certificates[0] = c
	return config, nil
}

//MakeListener returns a new net.Listener for http.Request
func MakeListener(protocol string, addr string, conf *tls.Config) (net.Listener, error) {
	var l net.Listener
	var err error

	if conf == nil {
		l, err = tls.Listen(protocol, addr, conf)
	} else {
		l, err = net.Listen(protocol, addr)
	}

	if err != nil {
		return nil, err
	}

	return l, nil
}

//NewHTTPServer returns a new http.Server using the provided listener
func NewHTTPServer(l net.Listener, handle http.Handler, c *tls.Config) (*http.Server, net.Listener, error) {
	tl, ok := l.(*net.TCPListener)

	if !ok {
		return nil, nil, fmt.Errorf("Listener is not type *net.TCPListener")
	}

	s := &http.Server{
		Addr:           tl.Addr().String(),
		Handler:        handle,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		TLSConfig:      c,
	}

	s.SetKeepAlivesEnabled(true)

	go s.Serve(tl)

	return s, tl, nil
}

// NewConn returns a tls.Conn object from the provided parameters.
func NewConn(protocol string, addr string) (net.Conn, error) {
	newConn, err := net.Dial(protocol, addr)
	if err != nil {
		return nil, err
	}

	return newConn, nil
}

// TLSConn returns a tls.Conn object from the provided parameters.
func TLSConn(protocol string, addr string, conf *tls.Config) (*tls.Conn, error) {
	newTls, err := tls.Dial(protocol, addr, conf)
	if err != nil {
		return nil, err
	}

	return newTls, nil
}

// TLSFromConn returns a new tls.Conn using the address and the certicates from
// the provided *tls.Conn.
func TLSFromConn(tl *tls.Conn, addr string) (*tls.Conn, error) {
	var conf *tls.Config

	if err := tl.Handshake(); err != nil {
		return nil, err
	}

	state := tl.ConnectionState()
	pool := x509.NewCertPool()

	for _, v := range state.PeerCertificates {
		pool.AddCert(v)
	}

	conf = &tls.Config{RootCAs: pool}
	newTls, err := tls.Dial("tcp", addr, conf)
	if err != nil {
		return nil, err
	}

	return newTls, nil
}

// ProxyHTTPRequest copies a http request from a src net.Conn connection to a
// destination net.Conn.
func ProxyHTTPRequest(src net.Conn, dest net.Conn) error {
	reader := bufio.NewReader(src)
	req, err := http.ReadRequest(reader)
	if err != nil {
		return err
	}

	if req == nil {
		return errors.New("No Request Read")
	}

	if err = req.Write(dest); err != nil {
		return err
	}

	resread := bufio.NewReader(dest)
	res, err := http.ReadResponse(resread, req)
	if err != nil {
		return err
	}

	if res != nil {
		return errors.New("No Response Read")
	}

	if err = res.Write(src); err != nil {
		return err
	}

	return nil
}

// hop headers, These are removed when sent to the backend
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html.
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

// ConnToHTTP proxies a requests from a net.Conn to a destination request, writing
// the response to provided ResponseWriter.
func ConnToHTTP(src net.Conn, destReq *http.Request, destRes http.ResponseWriter) error {
	reader := bufio.NewReader(src)
	req, err := http.ReadRequest(reader)
	if err != nil {
		return err
	}

	destReq.Method = req.Method

	for key, val := range req.Header {
		destReq.Header.Set(key, strings.Join(val, ","))
	}

	for _, v := range hopHeaders {
		destReq.Header.Del(v)
	}

	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return err
	}

	//add us to the proxy list or makeone
	hops, ok := req.Header["X-Forwarded-For"]
	if ok {
		ip = strings.Join(hops, ",") + "," + ip
	}

	destReq.Header.Set("X-Forwarded-For", ip)

	var buf bytes.Buffer
	if req.Body != nil {
		io.Copy(&buf, req.Body)
	}

	if buf.Len() > 0 {
		destReq.Body = ioutil.NopCloser(&buf)
		destReq.ContentLength = int64(buf.Len())
	}

	res, err := http.DefaultClient.Do(destReq)
	if err != nil {
		return err
	}

	for k, v := range res.Header {
		destRes.Header().Add(k, strings.Join(v, ","))
	}

	if err := res.Write(destRes); err != nil {
		return err
	}

	return nil
}

// HTTPToConn proxies a src Request to a net.Con connection and writes back
// the response to the src Response.
func HTTPToConn(srcReq *http.Request, srcRes http.ResponseWriter, dest net.Conn) error {
	if err := srcReq.Write(dest); err != nil {
		return err
	}

	resRead := bufio.NewReader(dest)
	res, err := http.ReadResponse(resRead, srcReq)
	if err != nil {
		return err
	}

	for key, val := range res.Header {
		srcRes.Header().Set(key, strings.Join(val, ","))
	}

	srcRes.WriteHeader(res.StatusCode)

	if err := res.Write(srcRes); err != nil {
		return err
	}

	return nil
}
