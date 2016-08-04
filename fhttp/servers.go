package fhttp

type HTTPDrive struct{
  router *httptreemux.Tree
  handlers []fractals.Handler // slice of global handlers.
}

// Serve lunches the drive with a http server.
func (hd *HTTPDrive) Serve(addr string) {
    LaunchHTTP(addr, hd.router)
}

// ServeTLS lunches the drive with a http server.
func (hd *HTTPDrive) ServeTLS(addr string, certFile string, keyFile string) {
    LaunchHTTPS(addr,certFile, keyFile, hd.router)
}

// NewHTTPDrive returns a new instance of the HTTPDrive struct.
func NewHTTPDrive() *HTTPDrive{
    var drive HTTPDrive
    drive.router = httptreemux.New()
    return &drive
}

// Endpoint defines a struct for registering router paths with the HTTPDrive router.
type Endpoint struct{
  Path string
  Method string
  Action interface{} 
}

func (e Endpoint) HandlerFunc() func(w http.ResponseWriter,r *http.Request, params map[string]interface{}) {
    var action func(context.Context, *Request) error
    
    switch e.Action.(type){
        case func(w http.ResponseWriter,r *http.Request, params map[string]interface{}):
            return e.Action.(func(w http.ResponseWriter,r *http.Request, params map[string]interface{}))
        case func(context.Context, *Request) error:
            action = e.Action.(func(context.Context, *Request) error)
        case fractals.Handler:
        handler
        return func(context.Context, *Request) error{
            
        }
        
    }
}

// Route returns a function which registers all 
func Route(drive *HTTPDrive) func(Endpoint) error {
    return func(end Endpoint) error {
        drive.Handle(end.Method, end.Path, func())
    }
} 