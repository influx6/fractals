package main

import (
  "os"
  "flag"
  "path/filepath"

  "net/http"
  "github.com/influx6/faux/context"
  "github.com/influx6/fractals/fhttp"
)

func main(){

  var (
    addrs string
    hasIndexFile bool
    basePath string
    assetPath string
  )

  pwd, err := os.Getwd()
  if err != nil {
    panic(err)
  }

  defaultAssets := filepath.Join(pwd, "assets")


  flag.StringVar(&addrs,"addrs",":4050", "addrs: The address and port to use for the http server.")
  flag.StringVar(&basePath,"base",pwd, "base: This values sets the path to be loaded as the base path.\n\t")
  flag.StringVar(&assetPath,"assets",defaultAssets, "assets: sets the absolute path to use for assets.\n\t")
  flag.BoolVar(&hasIndexFile,"withIndex",true, "withIndex: Indicates whether we should serve index.html as root path.")
  flag.Parse()

  basePath = filepath.Clean(basePath)
  assetPath = filepath.Clean(assetPath)


  app_http := fhttp.NewHTTP([]fhttp.DriveMiddleware{
    fhttp.WrapMiddleware(fhttp.Logger()),
  }, nil)

  app_router := fhttp.Route(app_http)

  app_router(fhttp.Endpoint{
    Path: "/assets/*",
    Method: "GET",
    Action: func(ctx context.Context, rw *fhttp.Request) error {return nil },
    LocalMW: fhttp.FileServer(assetPath, "/assets/"),
  })

  if hasIndexFile {
    app_router(fhttp.Endpoint{
      Path: "/",
      Method: "GET",
      Action: func(ctx context.Context, rw *fhttp.Request) error {return nil },
      LocalMW: fhttp.IndexServer(basePath, "index.html",""),
    })
  }

  http.ListenAndServe(addrs, app_http)
}