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

// PathName returns the path of the received *Request.
func PathName() fractals.Handler {
	return fractals.MustWrap(func(rw *fhttp.Request) string {
		return rw.Req.URL.Path
	})
}

// FileServer returns a handler capable of serving different files from the provided
// directory but using inputed URL path.
func FileServer(dir string, prefix string) fractals.Handler {
	var stripper Handler

	if prefix != "" {
		stripper = fs.StripPrefix(prefix)
	} else {
		stripper = fractals.IdentityHandler()
	}

	return fractals.SubLiftReplay(true, IdentityMiddleware(), MimeWriter(),
		PathName(), stripper, fs.ResolvePathStringIn(dir), fs.ReadFile())
}

// DirServer returns a fractals.Handler which servers a giving directory
// every single time it receives a request.
func DirServer(dir string) fractals.Handler {
	return fractals.SubLift(func(rw *Request, data []byte) (*Request, error) {
		if _, err := rw.Res.Write(data); err != nil {
			return nil, err
		}

		return rw, nil
	}, IdentityMiddleware(),
		fractals.Replay(dir), fs.ReadDirPath(), fs.SkipStat(fs.IsDir), fs.UnwrapStats(),
		fhttp.JSONEncoder())
}
