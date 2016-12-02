package fhttp

import (
	"errors"
	"io"
	"net/http"

	"github.com/dimfeld/httptreemux"
	"github.com/influx6/faux/context"
	"github.com/influx6/fractals"
)

// WrapFractalHandler returns a new http.HandlerFunc for recieving http request.
func WrapFractalHandler(handler fractals.Handler) func(http.ResponseWriter, *http.Request, map[string]string) {
	return WrapFractalHandlerWith(context.New(), handler)
}

// WrapFractalHandlerWith returns a http.HandlerFunc which accepts an extra parameter and
// passes the request objects to the handler. If no response was sent when
// the handlers are runned and an error came back then we write the error
// as response.
func WrapFractalHandlerWith(ctx context.Context, handler fractals.Handler) func(http.ResponseWriter, *http.Request, map[string]string) {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		rw := &Request{
			Params: Param(params),
			Res:    NewResponseWriter(w),
			Req:    r,
		}

		_, err := handler(ctx, nil, rw)
		if err != nil && !rw.Res.DataWritten() {
			RenderResponseError(err, rw)
		}
	}
}

// WrapRequestFractalHandler returns a function which wraps a fractal.Handler
// passing in the request object it receives.
func WrapRequestFractalHandler(handler fractals.Handler) func(context.Context, *Request) error {
	return func(ctx context.Context, rw *Request) error {
		_, err := handler(ctx, nil, rw)
		return err
	}
}

// DriveMiddleware defines a function type which accepts a context and Request
// returning the request or an error.
type DriveMiddleware func(context.Context, *Request) (*Request, error)

// IdentityMiddleware returns a function which always returns the Request object received.
func IdentityMiddleware() func(context.Context, *Request) (*Request, error) {
	return func(_ context.Context, rw *Request) (*Request, error) {
		return rw, nil
	}
}

// IdentityMiddlewareHandler returns the IdentityMiddleware returned value
// as a fractals.Handler.
func IdentityMiddlewareHandler() fractals.Handler {
	return fractals.MustWrap(IdentityMiddleware())
}

// WrapMW returns a new handler where the first wraps the second with its returned
// values.
func WrapMW(h1, h2 DriveMiddleware) DriveMiddleware {
	return func(ctx context.Context, rw *Request) (*Request, error) {
		m1, e1 := h1(ctx, rw)
		if e1 != nil {
			return nil, e1
		}

		return h2(ctx, m1)
	}
}

// LiftWM wraps a series of DriveMiddleware and returns a DriveMiddleware where
// each feeds its returns as the input of the next.
func LiftWM(mws ...DriveMiddleware) DriveMiddleware {
	var base DriveMiddleware

	for i := len(mws) - 1; i >= 0; i-- {
		if mws[i] == nil {
			continue
		}

		if base == nil {
			base = mws[i]
			continue
		}

		base = WrapMW(mws[i], base)
	}

	return base
}

// WrapForMW takes a giving interface and asserts it into a DriveMiddleware or
// wraps it if needed, returning the middleware.
func WrapForMW(wm interface{}) DriveMiddleware {
	var localWM DriveMiddleware

	switch wm.(type) {
	case func(context.Context, error, interface{}) (interface{}, error):
		localWM = WrapMiddleware(wm.(func(context.Context, error, interface{}) (interface{}, error)))
	case []func(context.Context, error, interface{}) (interface{}, error):
		fm := wm.([]fractals.Handler)
		localWM = WrapMiddleware(fm...)
	case func(interface{}) fractals.Handler:
		localWM = WrapMiddleware(wm.(func(interface{}) fractals.Handler)(fractals.IdentityHandler()))
	case func(context.Context, *Request) error:
		elx := wm.(func(context.Context, *Request) error)
		localWM = func(ctx context.Context, rw *Request) (*Request, error) {
			if err := elx(ctx, rw); err != nil {
				return nil, err
			}

			return rw, nil
		}
	case func(ctx context.Context, rw *Request) (*Request, error):
		localWM = wm.(func(ctx context.Context, rw *Request) (*Request, error))
	default:
		mws := fractals.MustWrap(wm)
		localWM = WrapMiddleware(mws)
	}

	return localWM
}

// WrapForAction returns a action which is typed asserts or morph into a
// fractal http handler.
func WrapForAction(action interface{}) func(context.Context, *Request) error {
	switch action.(type) {
	case func(w http.ResponseWriter, r *http.Request, params map[string]interface{}):
		handler := action.(func(w http.ResponseWriter, r *http.Request, params map[string]string))
		return func(ctx context.Context, rw *Request) error {
			handler(rw.Res, rw.Req, map[string]string{})
			return nil
		}

	case func(context.Context, *Request) error:
		return action.(func(context.Context, *Request) error)

	case func(context.Context, error, interface{}) (interface{}, error):
		handler := action.(func(context.Context, error, interface{}) (interface{}, error))
		return func(ctx context.Context, r *Request) error {
			if _, err := handler(ctx, nil, r); err != nil {
				return err
			}

			return nil
		}

	case fractals.Handler:
		handler := action.(fractals.Handler)
		return func(ctx context.Context, r *Request) error {
			if _, err := handler(ctx, nil, r); err != nil {
				return err
			}

			return nil
		}

	default:
		mw := fractals.MustWrap(action)
		return WrapRequestFractalHandler(mw)
	}
}

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
	globalMW      DriveMiddleware // global middleware.
	globalMWAfter DriveMiddleware // global middleware.
}

// Serve lunches the drive with a http server.
func (hd *HTTPDrive) Serve(addr string) {
	LaunchHTTP(addr, hd)
}

// ServeTLS lunches the drive with a http server.
func (hd *HTTPDrive) ServeTLS(addr string, certFile string, keyFile string) {
	LaunchHTTPS(addr, certFile, keyFile, hd)
}

// MW returns the giving lists of passed in middleware, it is provided as
// as a convenience function.
func MW(md ...DriveMiddleware) []DriveMiddleware {
	return md
}

// NewHTTP returns a new instance of the HTTPDrive struct.
func NewHTTP(before []DriveMiddleware, after []DriveMiddleware) *HTTPDrive {
	var drive HTTPDrive
	drive.TreeMux = httptreemux.New()
	drive.globalMW = LiftWM(before...)
	drive.globalMWAfter = LiftWM(after...)
	return &drive
}

// Endpoint defines a struct for registering router paths with the HTTPDrive router.
type Endpoint struct {
	Path    string
	Method  string
	Action  interface{}
	LocalMW interface{}
	AfterWM interface{}
}

func (e Endpoint) handlerFunc(globalBeforeWM, globalAfterWM DriveMiddleware) func(w http.ResponseWriter, r *http.Request, params map[string]string) {
	action := WrapForAction(e.Action)

	var localWM DriveMiddleware
	var afterWM DriveMiddleware

	if e.LocalMW != nil {
		localWM = WrapForMW(e.LocalMW)
	}

	if e.AfterWM != nil {
		afterWM = WrapForMW(e.AfterWM)
	}

	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		ctx := context.New()
		rw := &Request{
			Params: Param(params),
			Res:    NewResponseWriter(w),
			Req:    r,
		}

		// Run the global middleware first and recieve its returned values.
		if globalBeforeWM != nil {
			_, err := globalBeforeWM(ctx, rw)
			if err != nil && !rw.Res.DataWritten() {
				RenderResponseError(err, rw)
				return
			}
		}

		// Run local middleware second and receive its return values.
		if localWM != nil {
			_, err := localWM(ctx, rw)
			if err != nil && !rw.Res.DataWritten() {
				RenderResponseError(err, rw)
				return
			}
		}

		if err := action(ctx, rw); err != nil && !rw.Res.DataWritten() {
			RenderResponseError(err, rw)
			return
		}

		if afterWM != nil {
			_, err := afterWM(ctx, rw)
			if err != nil && !rw.Res.DataWritten() {
				RenderResponseError(err, rw)
				return
			}
		}

		if globalAfterWM != nil {
			_, err := globalAfterWM(ctx, rw)
			if err != nil && !rw.Res.DataWritten() {
				RenderResponseError(err, rw)
				return
			}
		}

	}
}

// Route returns a functional register, which uses the same drive for registring
// http endpoints.
func Route(drive *HTTPDrive) func(Endpoint) error {
	return func(end Endpoint) error {
		drive.Handle(end.Method, end.Path, end.handlerFunc(drive.globalMW, drive.globalMWAfter))
		return nil
	}
}

// RouteBy provides a more direct function that lets you specify the drive and
// endpoint directly.
func RouteBy(drive *HTTPDrive, end Endpoint) error {
	drive.Handle(end.Method, end.Path, end.handlerFunc(drive.globalMW, drive.globalMWAfter))
	return nil
}
