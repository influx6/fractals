package fhttp

import (
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/influx6/fractals"
	"github.com/influx6/fractals/fhttp/mimes"
	"github.com/influx6/fractals/fs"
)

// MimeWriter tries to extract the mime type from the possible extension in
// the URL path name and applies that to the request.
func MimeWriter() fractals.Handler {
	return fractals.MustWrap(func(rw *Request) *Request {
		ctn := mimes.GetByExtensionName(filepath.Ext(rw.Req.URL.Path))
		rw.Res.Header().Add("Content-Type", ctn)
		return rw
	})
}

// MimeWriterFor writes the mimetype for the provided file depending on the
// extension of the file.
func MimeWriterFor(file string) fractals.Handler {
	return fractals.MustWrap(func(rw *Request) *Request {
		ctn := mimes.GetByExtensionName(filepath.Ext(file))
		rw.Res.Header().Add("Content-Type", ctn)
		return rw
	})
}

// LogWith returns a Handler which logs to the provided Writer details of the
// http request.
func LogWith(w io.Writer, behave func(io.Writer, *Request)) fractals.Handler {
	return fractals.MustWrap(func(rw *Request) *Request {
		behave(w, rw)
		return rw
	})
}

// ResponseLogger provides a logger which logs the initial request of the
// incoming request.
func ResponseLogger(w io.Writer) fractals.Handler {
	return LogWith(w, func(ws io.Writer, rw *Request) {
		now := time.Now().UTC()
		content := rw.Res.Header().Get("Content-Type")
		fmt.Fprintf(ws, "HTTP : %q : Content{%s} : Status{%d} : URI{%s} : DataSize{%d}\n", now, content, rw.Res.Status(), rw.Req.URL, rw.Res.Size())
	})
}

// RequestLogger provides a logger which logs the initial request of the
// incoming request.
func RequestLogger(w io.Writer) fractals.Handler {
	return LogWith(w, func(ws io.Writer, rw *Request) {
		now := time.Now().UTC()
		content := rw.Req.Header.Get("Accept")
		if !rw.Res.StatusWritten() {
			fmt.Fprintf(ws, "HTTP : %q : Content{%s} : Method{%s} : URI{%s}\n", now, content, rw.Req.Method, rw.Req.URL)
		} else {
			fmt.Fprintf(ws, "HTTP : %q : Status{%d} : Content{%s} : Method{%s} : URI{%s}\n", now, rw.Res.Status(), rw.Res.Header().Get("Content-Type"), rw.Req.Method, rw.Req.URL)
		}
	})
}

// PathName returns the path of the received *Request.
func PathName() fractals.Handler {
	return fractals.MustWrap(func(rw *Request) string {
		return rw.Req.URL.Path
	})
}

// JoinPathName returns the path of the received *Request.
func JoinPathName(file string) fractals.Handler {
	return fractals.MustWrap(func(rw *Request) string {
		return filepath.Join(rw.Req.URL.Path, file)
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

// FileServer returns a handler capable of serving different files from the provided
// directory but using inputed URL path.
func FileServer(file string) fractals.Handler {
	return fractals.SubLift(func(rw *Request, data []byte) (*Request, error) {
		if _, err := rw.Res.Write(data); err != nil {
			return nil, err
		}

		return rw, nil
	}, IdentityMiddlewareHandler(), MimeWriterFor(file), fractals.Replay(file),
		fs.ReadFile())
}

// DirFileServer returns a handler capable of serving different files from the provided
// directory but using inputed URL path.
func DirFileServer(dir string, prefix string) fractals.Handler {
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
	}, IdentityMiddlewareHandler(), MimeWriter(),
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
	}, IdentityMiddlewareHandler(),
		fractals.Replay(dir), fs.ReadDirPath(), fs.SkipStat(fs.IsDir), fs.UnwrapStats(),
		JSONEncoder())
}
