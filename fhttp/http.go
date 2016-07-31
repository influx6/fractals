package fhttp

import (
	"net/http"
	"os"
	"os/signal"

	"github.com/influx6/faux/context"
	"github.com/influx6/fractals"
)

// CORS setup a generic CORS hader within the response for recieved request response.
func CORS() fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, wm *Request) {
		wm.Res.Header().Set("Access-Control-Allow-Origin", "*")
		wm.Res.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		wm.Res.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		wm.Res.Header().Set("Access-Control-Max-Age", "86400")
	})
}

// Headers returns a fractals.Handler which hads the provided values into the
// response headers.
func Headers(h map[string]string) fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, wm *Request) {
		for key, value := range h {
			wm.Res.Header().Set(key, value)
		}
	})
}

// WrapMW returns a new http.HandlerFunc for recieving http request.
func WrapMW(handler fractals.Handler) func(http.ResponseWriter, *http.Request, map[string]string) {
	return WrapMWWith(context.New(), handler)
}

// WrapMWWith returns a http.HandlerFunc which accepts an extra parameter and
// passes the request objects to the handler.
func WrapMWWith(ctx context.Context, handler fractals.Handler) func(http.ResponseWriter, *http.Request, map[string]string) {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		handler(ctx, nil, &Request{
			Params: Param(params),
			Res:    NewResponseWriter(w),
			Req:    r,
		})
	}
}

// LaunchHTTP lunches a http server, setting up the signal handler needed.
func LaunchHTTP(addr string, mux http.Handler) {
	go func() {
		http.ListenAndServe(addr, mux)
	}()

	// Listen for an interrupt signal from the OS.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
}

// LaunchHTTPS lunches a http server, setting up the signal handler needed.
func LaunchHTTPS(addr string, tlsKey string, tlsCert string, mux http.Handler) {
	go func() {
		http.ListenAndServeTLS(addr, tlsCert, tlsKey, mux)
	}()

	// Listen for an interrupt signal from the OS.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
}
