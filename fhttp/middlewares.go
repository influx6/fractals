package fhttp

import (
	"mime"
	"path/filepath"

	"github.com/influx6/fractals"
	"github.com/influx6/fractals/fhttp"
	"github.com/influx6/fractals/fs"
)

// MimeWriter tries to extract the mime type from the possible extension in
// the URL path name and applies that to the request.
func MimeWriter() fractals.Handler {
	return fractals.MustWrap(func(rw *fhttp.Request) *fhttp.Request {
		ctn := mime.TypeByExtension(filepath.Ext(rw.Req.URL.Path))

		if ctn == "" {
			ctn = "text/plain"
		}

		rw.Res.Header().Add("Content-Type", ctn)
		return rw
	})
}

//
func PathName() fractals.Handler {
	return fractals.MustWrap(func(rw *fhttp.Request) string {
		return rw.Req.URL.Path
	})
}

// FileServer returns a server capable of serving different files from the provided
// directory but using inputed URL path.
func FileServer(dir string) fractals.Handler {
	files := fractals.Lift(fs.ResolvePathStringIn(dir), fs.ReadFile())

}

func DirServer(dir string) fractals.Handler {
	dirs := fractals.Lift(fs.ReadDirPath(), fs.SkipStat(fs.IsDir), fs.UnwrapStats(), fhttp.JSONEncoder())

}
