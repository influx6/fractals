package fhttp

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
)

// LaunchHTTP lunches a http server, setting up the signal handler needed.
func LaunchHTTP(addr string, mux http.Handler) {
	go func() {
		fmt.Printf("HTTP Server starting... {Addr: %q}", addr)
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
		fmt.Printf("HTTPS Server starting... {Addr: %q}", addr)
		http.ListenAndServeTLS(addr, tlsCert, tlsKey, mux)
	}()

	// Listen for an interrupt signal from the OS.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
}
