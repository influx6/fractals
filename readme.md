# Fractals
Fractal evolved out of an experiment in using middleware like patterns to create
functional programming behavior in Go. It provides functions as its expressive
approach to create operations that both are pure and combinative. In essence
follows the idea of a stack of pure functions combined for reuse.

## Install

```
go get -u -v github.com/influx6/fractals
```

## Example

- Handler Examples

```go
// MimeWriter tries to extract the mime type from the possible extension in
// the URL path name and applies that to the request.
func MimeWriter() fractals.Handler {
	return fractals.MustWrap(func(rw *Request) *Request {
		ctn := mimes.GetByExtensionName(filepath.Ext(rw.Req.URL.Path))
		rw.Res.Header().Add("Content-Type", ctn)
		return rw
	})
}

// IndexServer returns a handler capable of serving a specific file from the provided
// directores which it recieves but using combining the filename with the giving
// path from the reequest.
func IndexServer(dir string, index string, prefix string) fractals.Handler {
	var stripper fractals.Handler

	if prefix != "" {
		stripper = fs.StripPrefix(prefix)
	} else {
		stripper = fractals.IdentityHandler()
	}

	return fractals.SubLift(func(rw *Request, data []byte) (*Request, error) {
		if _, err := rw.Res.Write(data); err != nil {
			return nil, err
		}

		return rw, nil
	}, IdentityMiddlewareHandler(), MimeWriterFor(index),
		JoinPathName(index), stripper, fs.ResolvePathStringIn(dir), fs.ReadFile())
}
```


- Observable Examples

```go
  var wg sync.WaitGroup

	ob := fractals.NewObservable(fractals.NewBehaviour(func(name string) string {
		return "Mr." + name
	}, nil, nil), false)

	ob2 := fractals.DebounceWithObserver(ob, 10*time.Millisecond)

	ob2.Subscribe(fractals.NewObservable(fractals.NewBehaviour(func(name string) {
		fmt.Printf("Debounce: %s\n", name)
		wg.Done()
	}, nil, nil), false))

	// These items wont be seen.
	ob.Next(context.New(), "Thunder")
	ob.Next(context.New(), "Thunder2")
	ob.Next(context.New(), "Thunder3")
	ob.Next(context.New(), "Thunder4")

	<-time.After(11 * time.Millisecond)
	ob.Next(context.New(), "Lightening")

  ob.Done()
  ob.End()

  ob2.Done()
  ob2.End()
```
