// Package fractals provides a functional reactive fractalssub structure to leverage a
// pure function style reactive behaviour. Originally pulled from fractals.Node.
// NOTE: Any use of "asynchronouse" actually means to "run within a goroutine",
// and inversely, the use of "synchronouse" means to run it within the current
// goroutine, generally referred to as "main", or in other words.
package fractals

import (
	"fmt"
	"reflect"

	"github.com/influx6/faux/context"
	"github.com/influx6/faux/reflection"
	"github.com/influx6/faux/regos"
)

var (
	errorType = reflect.TypeOf((*error)(nil)).Elem()
	boolType  = reflect.TypeOf((*bool)(nil)).Elem()
	ctxType   = reflect.TypeOf((*context.Context)(nil)).Elem()
	uType     = reflect.TypeOf((*interface{})(nil)).Elem()

	dZeroError = reflect.Zero(errorType)
	dZeroBool  = reflect.Zero(boolType)
)

//==============================================================================

// HandlerMap define a map of Handlers.
type HandlerMap map[string]Handler

// Has returns true/false if the tag exists for a built result.
func (r HandlerMap) Has(tag string) bool {
	_, ok := r[tag]
	return ok
}

// Get returns the Result for the specified build tag.
func (r HandlerMap) Get(tag string) Handler {
	return r[tag]
}

//==============================================================================

// Handler defines a function type which processes data and accepts a ReadWriter
// through which it sends its reply.
type Handler func(context.Context, error, interface{}) (interface{}, error)

// MustWrap returns the Handler else panics if it fails to create the Handler
// from the provided function type.
func MustWrap(node interface{}) Handler {
	dh := Wrap(node)
	if dh != nil {
		return dh
	}

	panic("Invalid type provided for Handler")
}

// Wrap returns a new Handler wrapping the provided value as needed if
// it matches its DataHandler, ErrorHandler, Handler or magic function type.
// MagicFunction type is a function which follows this type form:
// func(context.Context, error, <CustomType>).
func Wrap(node interface{}) Handler {
	var hl Handler

	switch node.(type) {
	case func():
		hl = func(ctx context.Context, err error, d interface{}) (interface{}, error) {
			node.(func())()

			if err != nil {
				return nil, err
			}

			return d, err
		}
	case func(context.Context, error, interface{}) (interface{}, error):
		hl = node.(func(context.Context, error, interface{}) (interface{}, error))
	case func(context.Context, interface{}):
		hl = wrapDataWithNoReturn(node.(func(context.Context, interface{})))
	case func(context.Context, interface{}) interface{}:
		hl = wrapDataWithReturn(node.(func(context.Context, interface{}) interface{}))
	case func(context.Context, interface{}) (interface{}, error):
		hl = wrapData(node.(func(context.Context, interface{}) (interface{}, error)))
	case func(context.Context, error) (interface{}, error):
		hl = wrapError(node.(func(context.Context, error) (interface{}, error)))
	case func(context.Context, error):
		hl = func(ctx context.Context, err error, d interface{}) (interface{}, error) {
			if err != nil {
				node.(func(context.Context, error))(ctx, err)
			}

			return d, err
		}
	case func(context.Context, error) error:
		hl = func(ctx context.Context, err error, d interface{}) (interface{}, error) {
			if err != nil {
				(node.(func(context.Context, error)))(ctx, err)
			}

			return d, err
		}
	case func(interface{}) interface{}:
		hl = wrapDataOnly(node.(func(interface{}) interface{}))
	case func(interface{}):
		hl = wrapJustData(node.(func(interface{})))
	case func(error):
		hl = wrapJustError(node.(func(error)))
	case func(error) error:
		hl = wrapErrorReturn(node.(func(error) error))
	case func() interface{}:
		hl = wrapNoData(node.(func() interface{}))
	case func(interface{}) error:
		hl = wrapErrorOnly(node.(func(interface{}) error))
	default:
		if !reflection.IsFuncType(node) {
			return nil
		}

		tm, _ := reflection.FuncValue(node)
		args, _ := reflection.GetFuncArgumentsType(node)

		dLen := len(args)

		var data reflect.Type
		var dZero reflect.Value

		var useContext bool
		var useErr bool
		var useData bool
		var isCustom bool

		// Check if this first item is a context.Context type.
		if dLen < 2 {
			useContext, _ = reflection.CanSetForType(ctxType, args[0])
			useErr, _ = reflection.CanSetForType(errorType, args[0])

			if !useErr {
				data = args[0]
				dZero = reflect.Zero(data)
				useData = true
				isCustom = true
			}
		}

		if dLen == 2 {
			useContext, _ = reflection.CanSetForType(ctxType, args[0])
			useErr, _ = reflection.CanSetForType(errorType, args[1])

			if !useErr {
				data = args[1]
				dZero = reflect.Zero(data)
				useData = true
				isCustom = true
			}
		}

		if dLen > 2 {
			useContext, _ = reflection.CanSetForType(ctxType, args[0])
			useErr, _ = reflection.CanSetForType(errorType, args[1])

			data = args[2]
			dZero = reflect.Zero(data)
			useData = true
            
           if !useContext || !useData || !useErr {
             return nil
           }
		}
        
        if !useData && !useErr {
             return nil
        }
        
		hl = func(ctx context.Context, err error, val interface{}) (interface{}, error) {
			var fnArgs []reflect.Value
			var resArgs []reflect.Value

			var mctx reflect.Value

			me := dZeroError
			md := dZero

			if useContext {
				mctx = reflect.ValueOf(ctx)
			}

			if err != nil {
				me = reflect.ValueOf(err)
			}

			// Flag to skip function if data does not match.
			breakOfData := true

			if val != nil && useData {
				ok, convertData := reflection.CanSetForType(data, reflect.TypeOf(val))
				if ok {
					breakOfData = false
					md = reflect.ValueOf(val)

					if convertData {
						md = md.Convert(data)
					}
				}
			}

			if !useContext && !useData && !useErr {
				resArgs = tm.Call(nil)
			} else {
				// fmt.Printf("%t:%t:%t -> %+s:%+s\n", useContext, useErr, useData, err, data)
				if isCustom && !useErr && err != nil {
					return nil, err
				}

				if !useContext && !useData && useErr && err != nil {
					return nil, err
				}

				if useContext && !useData && useErr && err != nil {
					return nil, err
				}

				// Call the function if it only cares about the error
				if useContext && useErr && me != dZeroError && !useData {
					fnArgs = []reflect.Value{mctx, me}
				}

				// If data does not match then skip this fall.
				if breakOfData && len(fnArgs) < 1 {
					return nil, nil
				}

				if !breakOfData {
					if useContext && useErr && useData {
						fnArgs = []reflect.Value{mctx, me, md}
					}

					if useContext && !useErr && useData {
						fnArgs = []reflect.Value{mctx, md}
					}

					if !useContext && useData && useErr {
						fnArgs = []reflect.Value{me, md}
					}

					if !useContext && useData && !useErr {
						fnArgs = []reflect.Value{md}
					}
				}

				resArgs = tm.Call(fnArgs)
			}

			resLen := len(resArgs)
			if resLen > 0 {

				if resLen < 2 {
					rOnly := resArgs[0]

					if erErr, ok := rOnly.Interface().(error); ok {
						return nil, erErr
					}

					return rOnly.Interface(), nil
				}

				rData := resArgs[0].Interface()
				rErr := resArgs[1].Interface()

				if erErr, ok := rErr.(error); ok {
					return rData, erErr
				}

				return rData, nil
			}

			return dZero, nil
		}
	}

	return hl
}

// DiscardData returns a new Handler which discards it's data and only forwards
// it's errors.
func DiscardData() Handler {
	return func(ctx context.Context, err error, _ interface{}) (interface{}, error) {
		return nil, err
	}
}

// DiscardError returns a new Handler which discards it's errors and only forwards
// its data.
func DiscardError() Handler {
	return func(ctx context.Context, _ error, data interface{}) (interface{}, error) {
		return data, nil
	}
}

// IdentityHandler returns a new Handler which forwards it's errors or data to
// its subscribers.
func IdentityHandler() Handler {
	return func(ctx context.Context, err error, data interface{}) (interface{}, error) {
		if err != nil {
			return nil, err
		}
		return data, nil
	}
}

// dataHandler defines a function type that concentrates on handling only data
// replies alone.
type dataHandler func(context.Context, interface{}) (interface{}, error)

// wrapData returns a Handler which wraps a DataHandler within it, but
// passing forward all errors it receives.
func wrapData(dh dataHandler) Handler {
	return func(m context.Context, err error, data interface{}) (interface{}, error) {
		if err != nil {
			return nil, err
		}

		return dh(m, data)
	}
}

// dataWithNoReturnHandler defines a function type that concentrates on handling only data
// replies alone.
type dataWithNoReturnHandler func(context.Context, interface{})

// wrapDataWithNoReturn returns a Handler which wraps a DataHandler within it, but
// passing forward all errors it receives.
func wrapDataWithNoReturn(dh dataWithNoReturnHandler) Handler {
	return func(m context.Context, err error, data interface{}) (interface{}, error) {
		if err != nil {
			return nil, err
		}

		dh(m, data)
		return data, nil
	}
}

type dataWithReturnHandler func(context.Context, interface{}) interface{}

// wrapDataWithReturn returns a Handler which wraps a DataHandler within it, but
// passing forward all errors it receives.
func wrapDataWithReturn(dh dataWithReturnHandler) Handler {
	return func(m context.Context, err error, data interface{}) (interface{}, error) {
		if err != nil {
			return nil, err
		}

		return dh(m, data), nil
	}
}

// NoDataHandler defines an Handler which allows a return value when called
// but has no data passed in.
type noDataHandler func() interface{}

// wrapNoData returns a Handler which wraps a NoDataHandler within it, but
// forwards all errors it receives. It calls its internal function
// with no arguments taking the response and sending that out.
func wrapNoData(dh noDataHandler) Handler {
	return func(m context.Context, err error, data interface{}) (interface{}, error) {
		if err != nil {
			return nil, err
		}

		res := dh()
		if erx, ok := res.(error); ok {
			return nil, erx
		}

		return res, nil
	}
}

// DataOnlyHandler defines a function type that concentrates on handling only data
// replies alone.
type dataOnlyHandler func(interface{}) interface{}

// wrapDataOnly returns a Handler which wraps a DataOnlyHandler within it, but
// passing forward all errors it receives.
func wrapDataOnly(dh dataOnlyHandler) Handler {
	return func(m context.Context, err error, data interface{}) (interface{}, error) {
		if err != nil {
			return nil, err
		}

		res := dh(data)
		if erx, ok := res.(error); ok {
			return nil, erx
		}

		return res, nil
	}
}

// JustDataHandler defines a function type which expects one argument.
type justDataHandler func(interface{})

// wrapJustData wraps a JustDataHandler and returns it as a Handler.
func wrapJustData(dh justDataHandler) Handler {
	return func(ctx context.Context, err error, d interface{}) (interface{}, error) {
		if err != nil {
			return nil, err
		}

		dh(d)
		return d, nil
	}
}

type justErrorHandler func(error)

// wrapJustError returns a Handler which wraps a DataOnlyHandler within it, but
// passing forward all errors it receives.
func wrapJustError(dh justErrorHandler) Handler {
	return func(ctx context.Context, err error, d interface{}) (interface{}, error) {
		if err != nil {
			dh(err)
			return nil, err
		}

		return d, nil
	}
}

type errorReturnHandler func(error) error

// wrapErrorReturn returns a Handler which wraps a DataOnlyHandler within it, but
// passing forward all errors it receives.
func wrapErrorReturn(dh errorReturnHandler) Handler {
	return func(ctx context.Context, err error, d interface{}) (interface{}, error) {
		if err != nil {
			return nil, dh(err)
		}

		return d, nil
	}
}

type errorHandler func(context.Context, error) (interface{}, error)

// wrapError returns a Handler which wraps a DataOnlyHandler within it, but
// passing forward all errors it receives.
func wrapError(dh errorHandler) Handler {
	return func(m context.Context, err error, data interface{}) (interface{}, error) {
		if err != nil {
			return dh(m, err)
		}

		return data, nil
	}
}

// errorOnlyHandler defines a function type that concentrates on handling only error
// replies alone.
type errorOnlyHandler func(interface{}) error

// wrapErrorOnly returns a Handler which wraps a ErrorOnlyHandler within it, but
// passing forward all errors it receives.
func wrapErrorOnly(dh errorOnlyHandler) Handler {
	return func(m context.Context, err error, data interface{}) (interface{}, error) {
		if err != nil {
			return nil, dh(data)
		}

		return data, nil
	}
}

//==============================================================================

// WrapHandlers returns a new handler where the first wraps the second with its returned
// values.
func WrapHandlers(h1 Handler, h2 Handler) Handler {
	return func(ctx context.Context, err error, data interface{}) (interface{}, error) {
		m1, e1 := h1(ctx, err, data)
		return h2(ctx, e1, m1)
	}
}

//==============================================================================

// LiftHandler defines a type that takes a interface which must be a function
// and wraps it inside a Handler, returning that Handler.
type LiftHandler func(interface{}) Handler

// Lift takes a series of handlers which handlers which it combines serially,
// it returns a function that takes a final function to receive results. Passing
// results from the previous to the next function to be called.
// If the value of the argument is not a function, then it panics.
func Lift(lifts ...Handler) LiftHandler {

	// We will stack the handlers where one outputs becomes the input of the next.
	return func(handle interface{}) Handler {
		mh := Wrap(handle)
		if mh == nil {
			panic("Expected handle passed into be a function")
		}

		base := mh

		for i := len(lifts) - 1; i >= 0; i-- {
			if lifts[i] == nil {
				continue
			}

			if base == nil {
				base = lifts[i]
				continue
			}

			base = WrapHandlers(lifts[i], base)
		}

		return func(ctx context.Context, err error, data interface{}) (interface{}, error) {
			if base != nil {
				return base(ctx, err, data)
			}

			return data, err
		}
	}
}

// RLiftHandler defines the type the Node function returns which allows providers
// to assign handlers to use for
type RLiftHandler func(...Handler) Handler

// RLift takes a handler as the last call and returns the a function that allows
// you to supply handlers which it combines serially. Lifting from the right.
// Passing results from the previous to the next function to be called.
// If the value of the argument is not a function, then it panics.
func RLift(handle interface{}) RLiftHandler {
	mh := Wrap(handle)
	if mh == nil {
		panic("Expected handle passed into be a function")
	}

	// We will stack the handlers where one outputs becomes the input of the next.
	return func(lifts ...Handler) Handler {
		base := mh

		for i := len(lifts) - 1; i >= 0; i-- {
			if lifts[i] == nil {
				continue
			}

			if base == nil {
				base = lifts[i]
				continue
			}

			base = WrapHandlers(lifts[i], base)
		}

		return func(ctx context.Context, err error, data interface{}) (interface{}, error) {
			if base != nil {
				return base(ctx, err, data)
			}

			return data, err
		}
	}
}

// Distribute takes the output from the provided handle and distribute
// it's returned values to the provided Handlers.
func Distribute(lifts ...Handler) LiftHandler {

	// We will stack the handlers where one outputs becomes the input of the next.
	return func(handle interface{}) Handler {
		mh := Wrap(handle)
		if mh == nil {
			panic("Expected handle passed into be a function")
		}

		return func(ctx context.Context, err error, data interface{}) (interface{}, error) {
			m1, e1 := mh(ctx, err, data)

			for _, lh := range lifts {
				lh(ctx, e1, m1)
			}

			return m1, e1
		}
	}
}

// RDistribute takes the output from the provided handle and distribute
// it's returned values to the provided Handlers.
func RDistribute(handle interface{}) RLiftHandler {
	mh := Wrap(handle)
	if mh == nil {
		panic("Expected handle passed into be a function")
	}

	// We will stack the handlers where one outputs becomes the input of the next.
	return func(lifts ...Handler) Handler {
		return func(ctx context.Context, err error, data interface{}) (interface{}, error) {
			m1, e1 := mh(ctx, err, data)

			for _, lh := range lifts {
				lh(ctx, e1, m1)
			}

			return m1, e1
		}
	}
}

// Response defines a struct for collecting the response from the Handlers.
type Response struct {
	Err   error
	Value interface{}
}

// DistributeButPack takes the output from the provided handle and distribute
// it's returned values to the provided Handlers and packs their responses in a
// slice []Response and returns that as the final response.
func DistributeButPack(lifts ...Handler) LiftHandler {

	// We will stack the handlers where one outputs becomes the input of the next.
	return func(handle interface{}) Handler {
		mh := Wrap(handle)
		if mh == nil {
			panic("Expected handle passed into be a function")
		}

		return func(ctx context.Context, err error, data interface{}) (interface{}, error) {
			var pack []Response

			m1, e1 := mh(ctx, err, data)

			for _, lh := range lifts {
				ld, le := lh(ctx, e1, m1)
				pack = append(pack, Response{
					Err:   le,
					Value: ld,
				})
			}

			return pack, nil
		}
	}
}

// RDistributeButPack takes the output from the provided handle and distribute
// it's returned values to the provided Handlers and packs their responses in a
// slice []Response and returns that as the final response.
func RDistributeButPack(handle interface{}) RLiftHandler {
	mh := Wrap(handle)
	if mh == nil {
		panic("Expected handle passed into be a function")
	}

	// We will stack the handlers where one outputs becomes the input of the next.
	return func(lifts ...Handler) Handler {
		return func(ctx context.Context, err error, data interface{}) (interface{}, error) {
			var pack []Response

			m1, e1 := mh(ctx, err, data)

			for _, lh := range lifts {
				ld, le := lh(ctx, e1, m1)
				pack = append(pack, Response{
					Err:   le,
					Value: ld,
				})
			}

			return pack, nil
		}
	}
}

// Collect takes all the returned values by passing the recieved arguments and
// applying them to the handle. Where the responses of the handle is packed into
// an array of type []Collected and then returned as the response of the function.
func Collect(lifts ...Handler) LiftHandler {

	// We will stack the handlers where one outputs becomes the input of the next.
	return func(handle interface{}) Handler {
		mh := Wrap(handle)
		if mh == nil {
			panic("Expected handle passed into be a function")
		}

		return func(ctx context.Context, err error, data interface{}) (interface{}, error) {
			var pack []Response

			for _, lh := range lifts {
				m1, e1 := lh(ctx, err, data)
				d1, de := mh(ctx, e1, m1)
				pack = append(pack, Response{
					Err:   de,
					Value: d1,
				})
			}

			return pack, nil
		}
	}
}

// RCollect takes all the returned values by passing the recieved arguments and
// applying them to the handle. Where the responses of the handle is packed into
// an array of type []Collected and then returned as the response of the function.
func RCollect(handle interface{}) RLiftHandler {
	mh := Wrap(handle)
	if mh == nil {
		panic("Expected handle passed into be a function")
	}

	// We will stack the handlers where one outputs becomes the input of the next.
	return func(lifts ...Handler) Handler {
		return func(ctx context.Context, err error, data interface{}) (interface{}, error) {
			var pack []Response

			for _, lh := range lifts {
				m1, e1 := lh(ctx, err, data)
				d1, de := mh(ctx, e1, m1)
				pack = append(pack, Response{
					Err:   de,
					Value: d1,
				})
			}

			return pack, nil
		}
	}
}

//==============================================================================

// StreamHandler defines a function type which requires a context, data and
// a bool value to indicate if the data is the last item in the stream, meaning
// the stream has ended.
type StreamHandler func(context.Context, interface{}, bool) interface{}

// WrapStreamHandler wraps a handler returning a StreamHandler.
func WrapStreamHandler(h interface{}) StreamHandler {
	switch h.(type) {
	case func(context.Context, error, interface{}) (interface{}, error):
		return func(ctx context.Context, data interface{}, end bool) interface{} {
			if ed, ok := data.(error); ok {
				res, err := (h.(func(context.Context, error, interface{}) (interface{}, error)))(ctx, ed, nil)
				if err != nil {
					return err
				}

				return res
			}

			res, err := (h.(func(context.Context, error, interface{}) (interface{}, error)))(ctx, nil, data)
			if err != nil {
				return err
			}

			return res
		}

	case func(context.Context, interface{}, bool) interface{}:
		return h.(func(context.Context, interface{}, bool) interface{})

	default:
		if !reflection.IsFuncType(h) {
			return nil
		}

		tm, _ := reflection.FuncValue(h)
		args, _ := reflection.GetFuncArgumentsType(h)

		dLen := len(args)

		var data reflect.Type
		var dZero reflect.Value

		var useContext bool
		var useBool bool
		var useData bool
		// var isCustom bool

		// Check if this first item is a context.Context type.
		if dLen < 2 {
			useContext, _ = reflection.CanSetForType(ctxType, args[0])
			useBool, _ = reflection.CanSetForType(boolType, args[0])

			if !useBool {
				data = args[0]
				dZero = reflect.Zero(data)
				useData = true
				// isCustom = true
			}
		}

		if dLen == 2 {
			useContext, _ = reflection.CanSetForType(ctxType, args[0])
			useBool, _ = reflection.CanSetForType(boolType, args[1])

			if !useBool {
				data = args[1]
				dZero = reflect.Zero(data)
				useData = true
				// isCustom = true
			}
		}

		if dLen > 2 {
			useContext, _ = reflection.CanSetForType(ctxType, args[0])
			useBool, _ = reflection.CanSetForType(boolType, args[2])

			data = args[1]
			dZero = reflect.Zero(data)
			useData = true
		}

		return func(ctx context.Context, val interface{}, done bool) interface{} {
			var fnArgs []reflect.Value
			var resArgs []reflect.Value

			var mctx reflect.Value

			me := reflect.ValueOf(done)
			md := dZero

			if useContext {
				mctx = reflect.ValueOf(ctx)
			}

			// Flag to skip function if data does not match.
			breakOfData := true

			if val != nil && useData {
				ok, convertData := reflection.CanSetForType(data, reflect.TypeOf(val))
				if ok {
					breakOfData = false
					md = reflect.ValueOf(val)

					if convertData {
						md = md.Convert(data)
					}
				}
			}

			if !useContext && !useData && !useBool {
				resArgs = tm.Call(nil)
			} else {
                
				// If data does not match then skip this fall.
				if breakOfData && len(fnArgs) < 1 {
					return nil
				}

				if !breakOfData {
					if useContext && useBool && useData {
						fnArgs = []reflect.Value{mctx, md,me}
					}

					if useContext && !useBool && useData {
						fnArgs = []reflect.Value{mctx, md}
					}

					if !useContext && useData && useBool {
						fnArgs = []reflect.Value{md, me}
					}

					if !useContext && useData && !useBool {
						fnArgs = []reflect.Value{md}
					}
				}

				resArgs = tm.Call(fnArgs)
			}

			resLen := len(resArgs)
			if resLen < 1 {
				return nil
			}

			return resArgs[0].Interface()
		}
	}
}

// Stream defines a interface which exposes a stream that allows continous
// data to be send down the pipeline.
type Stream interface {
	Emit(context.Context, interface{}, bool) interface{}
	Stream(interface{}) Stream
}

// MustSteram returns a new Stream using the handler it receives.
func MustStream(handler interface{}) Stream {
	hs := WrapStreamHandler(handler)
	if hs == nil {
		panic("Argument is not a StreamHandler")
	}

	var sm stream
	sm.main = hs
	return &sm
}

type stream struct {
	main StreamHandler
	next Stream
}

// Emit calls the next handler in the stream connection.
func (s *stream) Emit(ctx context.Context, data interface{}, end bool) interface{} {
	res := s.main(ctx, data, end)
	if s.next != nil {
		return s.next.Emit(ctx, res, end)
	}

	if edata, ok := res.(error); ok {
		return edata
	}

	return res
}

// Stream returns a new Stream with the provided Handler.
func (s *stream) Stream(h interface{}) Stream {
	var sm Stream

	switch h.(type) {
	case Stream:
		sm = h.(Stream)
	default:
		sm = MustStream(h)
	}

	if s.next != nil {
		return s.next.Stream(sm)
	}

	return sm
}

//==============================================================================

var hl = regos.New()

// Register adds the provided Handle maker into the internal handler maker
// registery.
func Register(name string, desc string, handlerMaker interface{}) {
	hl.Register(regos.Meta{
		Name:   name,
		Desc:   desc,
		Inject: handlerMaker,
	})
}

// Make returns a function that collects list of Handlers make maps which
// details the handler makers to call to create a map of Handlers keyed by
// the provided tags.
func Make() func(...map[string]interface{}) (HandlerMap, error) {
	var items []regos.Do

	hlMap := make(map[string]Handler)

	return func(tasks ...map[string]interface{}) (HandlerMap, error) {

		// If we are told no task then build
		if len(tasks) < 1 {
			if err := makeDo(hlMap, items); err != nil {
				return nil, err
			}

			return hlMap, nil
		}

		for _, task := range tasks {
			items = append(items, regos.Do{
				Name: task["name"].(string),
				Tag:  task["tag"].(string),
				Use:  task["use"],
			})
		}

		return hlMap, nil
	}
}

func makeDo(res HandlerMap, items []regos.Do) error {
	for _, do := range items {
		if res.Has(do.Tag) {
			return fmt.Errorf("Build Instruction for %s using reserved tag %s", do.Name, do.Tag)
		}

		func() {
			defer func() {
				if ex := recover(); ex != nil {
					fmt.Printf("Panic: failed to build Pub[%s] with Tag[%s]: [%s]\n", do.Name, do.Tag, ex)
				}
			}()

			pb := hl.NewBuild(do.Name, do.Use).(Handler)
			res[do.Tag] = pb
		}()
	}

	return nil
}
