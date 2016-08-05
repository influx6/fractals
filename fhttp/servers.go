package fhttp

import (
	"errors"
	"io"
	"net/http"

	"github.com/dimfeld/httptreemux"
	"github.com/influx6/faux/context"
	"github.com/influx6/fractals"
)

// DriveMiddleware defines a function type which accepts a context and Request
// returning the request or an error.
type DriveMiddleware func(context.Context, *Request) (*Request, error)

// WrapMiddleware wraps a Handler or lifted Handler from the slices of Handlers
// which when runned must return either a (*Request, io.WriteTo, io.Reader,[]byte),
// if it matches others except a *Request, then their contents will be written to
// the response and the *Request passed in will be returned, this allows the
// flexibility of multiple writers based on some operations.
// if the returned value matches non of these types, then an error is returned.
// When running the Handler if it returns an error then that error is returned
// as well.
func WrapMiddleware(handlers ...fractals.Handler) DriveMiddleware {
	var handle fractals.Handler

	if len(handlers) > 1 {
		handle = fractals.Lift(handlers...)(fractals.IdentityHandler())
	} else {
		handle = handlers[0]
	}

	return func(ctx context.Context, rw *Request) (*Request, error) {
		res, err := handle(ctx, nil, rw)
		if err != nil {
			return nil, err
		}

		// If the response is nil, then forward the request object.
		if res == nil {
			return rw, nil
		}

		// If its not nil, then check if it matches a series of types else
		// return an error.
		switch res.(type) {
		case []byte:
			if _, err := rw.Res.Write(res.([]byte)); err != nil {
				return nil, err
			}

			return rw, nil
		case io.Reader:
			rd := res.(io.Reader)
			if _, err := io.Copy(rw.Res, rd); err != nil {
				return nil, err
			}

			return rw, nil

		case io.WriterTo:
			wt := res.(io.WriterTo)

			if _, err := wt.WriteTo(rw.Res); err != nil {
				return nil, err
			}

			return rw, nil
		case *Request:
			return res.(*Request), nil
		default:
			return nil, errors.New("Invalid Type, Require *Request type")
		}
	}
}

// HTTPDrive defines a structure for providing a global httprouter by using the
// httptreemux.Tree underneath.
type HTTPDrive struct {
	*httptreemux.TreeMux
	globalMW DriveMiddleware // global middleware.
}

// Serve lunches the drive with a http server.
func (hd *HTTPDrive) Serve(addr string) {
	LaunchHTTP(addr, hd)
}

// ServeTLS lunches the drive with a http server.
func (hd *HTTPDrive) ServeTLS(addr string, certFile string, keyFile string) {
	LaunchHTTPS(addr, certFile, keyFile, hd)
}

// NewHTTPDrive returns a new instance of the HTTPDrive struct.
func NewHTTPDrive(handlers ...fractals.Handler) *HTTPDrive {
	var drive HTTPDrive
	drive.TreeMux = httptreemux.New()
	drive.globalMW = WrapMiddleware(handlers...)
	return &drive
}

// Endpoint defines a struct for registering router paths with the HTTPDrive router.
type Endpoint struct {
	Path    string
	Method  string
	Action  interface{}
	LocalMW interface{}
}

func (e Endpoint) handlerFunc(globalWM DriveMiddleware) func(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var action func(context.Context, *Request) error

	switch e.Action.(type) {
	case func(w http.ResponseWriter, r *http.Request, params map[string]interface{}):
		return e.Action.(func(w http.ResponseWriter, r *http.Request, params map[string]string))

	case func(context.Context, *Request) error:
		action = e.Action.(func(context.Context, *Request) error)

	case func(context.Context, error, interface{}) (interface{}, error):
		handler := e.Action.(func(context.Context, error, interface{}) (interface{}, error))
		action = func(ctx context.Context, r *Request) error {
			if _, err := handler(ctx, nil, r); err != nil {
				return err
			}

			return nil
		}

	default:
		return nil
	}

	var localWM DriveMiddleware

	switch e.LocalMW.(type) {
	case func(context.Context, error, interface{}) (interface{}, error):
		localWM = WrapMiddleware(e.LocalMW.(func(context.Context, error, interface{}) (interface{}, error)))
	case []func(context.Context, error, interface{}) (interface{}, error):
		fm := e.LocalMW.([]fractals.Handler)
		localWM = WrapMiddleware(fm...)
	case func(interface{}) fractals.Handler:
		localWM = WrapMiddleware(e.LocalMW.(func(interface{}) fractals.Handler)(fractals.IdentityHandler()))
	case func(context.Context, *Request) error:
		elx := e.LocalMW.(func(context.Context, *Request) error)
		localWM = func(ctx context.Context, rw *Request) (*Request, error) {
			if err := elx(ctx, rw); err != nil {
				return nil, err
			}

			return rw, nil
		}
	case func(ctx context.Context, rw *Request) (*Request, error):
		localWM = e.LocalMW.(func(ctx context.Context, rw *Request) (*Request, error))
	}

	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		ctx := context.New()
		rw := &Request{
			Params: Param(params),
			Res:    NewResponseWriter(w),
			Req:    r,
		}

		// Run the global middleware first and recieve its returned values.
		var err error
		if globalWM != nil {
			rw, err = globalWM(ctx, rw)
		}

		if err != nil && !rw.Res.DataWritten() {
			RenderResponseError(err, rw)
		}

		// Run local middleware second and receive its return values.
		if localWM != nil {
			rw, err = localWM(ctx, rw)
		}

		if err != nil && !rw.Res.DataWritten() {
			RenderResponseError(err, rw)
		}

		if err := action(ctx, rw); err != nil && !rw.Res.DataWritten() {
			RenderResponseError(err, rw)
		}
	}
}

// Route returns a functional register, which uses the same drive for registring
// http endpoints.
func Route(drive *HTTPDrive) func(Endpoint) error {
	return func(end Endpoint) error {
		drive.Handle(end.Method, end.Path, end.handlerFunc(drive.globalMW))
		return nil
	}
}

// RouteBy provides a more direct function that lets you specify the drive and
// endpoint directly.
func RouteBy(drive *HTTPDrive, end Endpoint) error {
	drive.Handle(end.Method, end.Path, end.handlerFunc(drive.globalMW))
	return nil
}
