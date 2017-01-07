package fhttp

import (
	"bytes"
	"encoding/json"
	"io"
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

// JSONDecoder decodes the data it recieves into an map type and returns the values.
func JSONDecoder() fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, data []byte) (map[string]interface{}, error) {
		ms := make(map[string]interface{})

		var b bytes.Buffer
		b.Write(data)

		if err := json.NewDecoder(&b).Decode(&ms); err != nil {
			return nil, err
		}

		return ms, nil
	})
}

// JSONWrite encodes the data it recieves into JSON and returns the values.
func JSONWrite(data interface{}) fractals.Handler {
	var bu bytes.Buffer
	var done bool

	return fractals.MustWrap(func(ctx context.Context, w io.Writer) error {
		if !done {
			if err := json.NewEncoder(&bu).Encode(data); err != nil {
				return err
			}

			done = true
		}

		if _, err := w.Write(bu.Bytes()); err != nil {
			return err
		}

		return nil
	})
}

// JSONEncoder encodes the data it recieves into JSON and returns the values.
func JSONEncoder() fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, data interface{}) ([]byte, error) {
		var d bytes.Buffer

		if err := json.NewEncoder(&d).Encode(data); err != nil {
			return nil, err
		}

		return d.Bytes(), nil
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
