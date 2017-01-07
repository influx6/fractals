package fs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/influx6/faux/context"
	"github.com/influx6/fractals"
)

// StripPrefix returns the path which has a prefix if found stripped from its
// string recieved.
func StripPrefix(prefix string) fractals.Handler {
	return fractals.MustWrap(func(path string) string {
		return strings.TrimPrefix(path, prefix)
	})
}

// ReadFile adds a readFile operation whoes contents get passed to the next
// event/Node/Task in the link.
func ReadFile() fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, path string) ([]byte, error) {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}

		defer file.Close()

		var buf bytes.Buffer
		_, err = io.Copy(&buf, file)
		if err != nil && err != io.EOF {
			return nil, err
		}

		return buf.Bytes(), nil
	})
}

// ReadReaderAndClose reads the data pulled from the received reader from the
// pipeline.
func ReadReaderAndClose() fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, r io.ReadCloser) ([]byte, error) {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		if err != nil && err != io.EOF {
			return nil, err
		}

		if err := r.Close(); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	})
}

// ReadReader reads the data pulled from the received reader from the
// pipeline.
func ReadReader() fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, r io.Reader) ([]byte, error) {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		if err != nil && err != io.EOF {
			return nil, err
		}

		return buf.Bytes(), nil
	})
}

// ReplayBytes resends the data provided everytime it is called.
func ReplayBytes(b []byte) fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context) []byte {
		return b
	})
}

// ReplayReader reads the data pulled from the reader everytime, buffers it
// and returns data everytime it gets called.
func ReplayReader(r io.Reader) fractals.Handler {
	var buf bytes.Buffer
	var read bool

	return fractals.MustWrap(func(_ interface{}) interface{} {
		if read {
			return buf.Bytes()
		}

		_, err := io.Copy(&buf, r)
		if err != nil && err != io.EOF {
			return err
		}

		read = true

		return buf.Bytes()
	})
}

// WriteBytes writes the giving bytes to a path it expects to receive when called,
// It appends the provided data to that path continously.
// It passes the data passed in to the its subscribers to
// both allow the chain of events to continue and to allow others to use the data
// as they please.
func WriteBytes(data []byte) fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, w io.Writer) error {
		written, err := w.Write(data)
		if err != nil {
			return err
		}

		if written != len(data) {
			return errors.New("Data written is not matching provided data")
		}

		return nil
	})
}

// WriteWriter expects to recieve []bytes as input and writes the provided
// bytes into the writer it recieves as argument. It returns error if the total
// written does not match the size of the bytes. It passes the incoming data
// down the pipeline.
func WriteWriter(w io.Writer) fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, data []byte) error {
		written, err := w.Write(data)
		if err != nil {
			return err
		}

		if written != len(data) {
			return errors.New("Data written is not matching provided data")
		}

		return nil
	})
}

// Close expects to receive a closer in its pipeline and closest the closer.
func Close() fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, w io.Closer) error {
		if err := w.Close(); err != nil {
			return err
		}

		return nil
	})
}

// OpenFile creates the giving file within the provided directly and
// writes the any recieved data into the file. It sends the file Handler,
// down the piepline.
func OpenFile(path string) fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, _ interface{}) (*os.File, error) {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}

		return file, nil
	})

}

// CreateFile creates the giving file within the provided directly and sends
// out the file handler.
func CreateFile(path string, useRoot bool) fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, root string) (*os.File, error) {
		if useRoot && root != "" {
			path = filepath.Join(root, path)
		}

		file, err := os.Create(path)
		if err != nil {
			return nil, err
		}

		return file, nil
	})
}

// MkFile either creates or opens an existing file for appending. It passes
// the file object for this files down its pipeline. If it gets a string from
// the pipeline, it uses that string if not empty as its root path.
func MkFile(path string, useRoot bool) fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, root string) (*os.File, error) {
		if useRoot && root != "" {
			path = filepath.Join(root, path)
		}

		file, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR, os.ModeAppend)
		if err != nil {
			return nil, err
		}

		return file, nil
	})
}

// ExtendedFileInfo composes a os.FileInfo to provide the fullPath property
// for a giving fileInfo.
type ExtendedFileInfo interface {
	os.FileInfo
	Path() string
	Dir() string
}

// NewExtendedFileInfo returns a structure which implements the ExtendedFileInfo
// interface.
func NewExtendedFileInfo(info os.FileInfo, root string) ExtendedFileInfo {
	ef := extendedFileInfo{
		FileInfo: info,
		path:     filepath.Join(root, info.Name()),
		root:     root,
	}

	return ef
}

type extendedFileInfo struct {
	os.FileInfo
	path string
	root string
}

// Dir returns the directory of the provided file.
func (e extendedFileInfo) Dir() string {
	return e.root
}

// Path returns the path of the provided file.
func (e extendedFileInfo) Path() string {
	return e.path
}

// ReadDir reads the giving path if indeed is a directory, else passing down
// an error down the provided pipeline. It extends the provided os.FileInfo
// with a structure that implements the ExtendedFileInfo interface. It sends the
// individual fileInfo instead of the slice of FileInfos.
func ReadDir(path string) fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, _ interface{}) ([]ExtendedFileInfo, error) {
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE, 0700)
		if err != nil {
			return nil, err
		}

		defer file.Close()

		dirs, err := file.Readdir(-1)
		if err != nil {
			return nil, err
		}

		var edirs []ExtendedFileInfo

		for _, dir := range dirs {
			edirs = append(edirs, NewExtendedFileInfo(dir, path))
		}

		return edirs, nil
	})
}

// ReadDirPath reads the giving path if indeed is a directory, else passing down
// an error down the provided pipeline. It extends the provided os.FileInfo
// with a structure that implements the ExtendedFileInfo interface. It sends the
// individual fileInfo instead of the slice of FileInfos.
func ReadDirPath() fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, path string) ([]ExtendedFileInfo, error) {
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE, 0700)
		if err != nil {
			return nil, err
		}

		defer file.Close()

		dirs, err := file.Readdir(-1)
		if err != nil {
			return nil, err
		}

		var edirs []ExtendedFileInfo

		for _, dir := range dirs {
			edirs = append(edirs, NewExtendedFileInfo(dir, path))
		}

		return edirs, nil
	})
}

// WalkDir walks the giving path if indeed is a directory, else passing down
// an error down the provided pipeline. It extends the provided os.FileInfo
// with a structure that implements the ExtendedFileInfo interface. It sends the
// individual fileInfo instead of the slice of FileInfos.
func WalkDir(path string) fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, _ interface{}) ([]ExtendedFileInfo, error) {
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE, 0700)
		if err != nil {
			return nil, err
		}

		defer file.Close()

		fdirs, err := file.Readdir(-1)
		if err != nil {
			return nil, err
		}

		var dirs []ExtendedFileInfo

		for _, dir := range fdirs {
			dirInfo := NewExtendedFileInfo(dir, path)

			// If this is a sysmbol link, then continue we won't read through it.
			if _, err := os.Readlink(dirInfo.Dir()); err == nil {
				continue
			}

			dirs = append(dirs, dirInfo)
		}

		return dirs, nil
	})

}

// Mkdir creates a directly returning the path down the pipeline. If the chain
// flag is on, then mkdir when it's pipeline receives a non-empty string as
// an argument, will join the string recieved with the path provided.
// This allows chaining mkdir paths down the pipeline.
func Mkdir(path string, chain bool) fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, root string) error {
		if chain && root != "" {
			path = filepath.Join(root, path)
		}

		if err := os.MkdirAll(path, 0700); err != nil {
			return err
		}

		return nil
	})
}

// ResolvePathIn returns an ExtendedFileInfo for paths recieved if they match
// a specific root directory once resolved using the root directory.
func ResolvePathIn(rootDir string) fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, path string) (ExtendedFileInfo, error) {
		absRoot, err := filepath.Abs(rootDir)
		if err != nil {
			return nil, err
		}

		rootPath := filepath.Clean(absRoot)
		finalPath := filepath.Clean(filepath.Join(rootDir, path))

		if strings.Contains(finalPath, rootPath) {
			return nil, fmt.Errorf("%q not within %q root", path, rootDir)
		}

		file, err := os.Open(finalPath)
		if err != nil {
			return nil, err
		}

		stat, err := file.Stat()
		if err != nil {
			return nil, err
		}

		return NewExtendedFileInfo(stat, filepath.Base(finalPath)), nil
	})
}

// ResolvePathStringIn returns the full valid path for paths recieved if they match
// a specific root directory once resolved using the root directory.
func ResolvePathStringIn(rootDir string) fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, path string) (string, error) {
		absRoot, err := filepath.Abs(rootDir)
		if err != nil {
			return "", err
		}

		rootPath := filepath.Clean(absRoot)
		finalPath := filepath.Clean(filepath.Join(rootDir, path))

		if strings.Contains(finalPath, rootPath) {
			return "", fmt.Errorf("%q not within %q root", path, rootDir)
		}

		return finalPath, nil
	})
}

// ResolvePath resolves a giving path or sets of paths into their  absolute
// form.
func ResolvePath() fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, path interface{}) (interface{}, error) {
		switch path.(type) {
		case string:
			return filepath.Abs(path.(string))
		case []string:
			var resolved []string

			for _, p := range path.([]string) {
				res, err := filepath.Abs(p)
				if err != nil {
					return nil, err
				}

				resolved = append(resolved, res)
			}

			return resolved, nil
		}

		return nil, errors.New("Invalid Type expected")
	})
}

// Remove deletes the giving path and passes the path down
// the pipeline.
func Remove(path string) fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, _ interface{}) error {
		if err := os.Remove(path); err != nil {
			return err
		}

		return nil
	})
}

// RemoveAll deletes the giving path and its subpaths if its a directory
// and passes the path down
// the pipeline.
func RemoveAll(path string) fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, _ interface{}) error {
		if err := os.RemoveAll(path); err != nil {
			return err
		}

		return nil
	})
}

// SkipStat takes a function to filter out the FileInfo that are running through
// its pipeline. This allows you to define specific file paths you wish to treat.
// If the filter function returns true, then any FileInfo/ExtendedFileInfo that
// match its criteria are sent down its pipeline.
func SkipStat(filter func(ExtendedFileInfo) bool) fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, info []ExtendedFileInfo) []ExtendedFileInfo {
		var filtered []ExtendedFileInfo

		for _, dir := range info {
			if !filter(dir) {
				continue
			}
			filtered = append(filtered, dir)
		}

		return filtered
	})
}

// UnwrapStats takes the provided ExtendedFileInfo and unwraps them into their
// full path, allows you to retrieve the strings path.
func UnwrapStats() fractals.Handler {
	return fractals.MustWrap(func(ctx context.Context, info []ExtendedFileInfo) []string {
		var dirs []string

		for _, dir := range info {
			dirs = append(dirs, dir.Path())
		}

		return dirs
	})
}

// IsDir defines a function which returns true/false if the FileInfo
// is a directory.
func IsDir(ex ExtendedFileInfo) bool {
	if ex.IsDir() {
		return true
	}

	return false
}

// IsFile defines a function which returns true/false if the FileInfo
// is a file not a directory.
func IsFile(ex ExtendedFileInfo) bool {
	if !ex.IsDir() {
		return true
	}
	return false
}
