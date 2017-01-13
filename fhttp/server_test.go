package fhttp_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influx6/faux/context"
	"github.com/influx6/fractals/fhttp"
)

func TestHTTPDrive(t *testing.T) {

	drive := fhttp.Drive(func(ctx context.Context, rw *fhttp.Request) (*fhttp.Request, error) {
		ctx.Set("names", []string{"fall-out", "reckless"})
		return rw, nil
	})(nil)

	router := fhttp.Route(drive)

	router(fhttp.Endpoint{
		Path:   "/names",
		Method: "GET",
		Action: func(ctx context.Context, rw *fhttp.Request) error {
			games, notFailed := ctx.Get("games")
			if !notFailed {
				return errors.New("Failed to retrieve games list")
			}

			names, notFailed := ctx.Get("names")
			if !notFailed {
				return errors.New("Failed to retrieve names list")
			}

			rw.Respond(http.StatusOK, map[string][]string{
				"games": games.([]string),
				"names": names.([]string),
			})

			return nil
		},
		LocalMW: func(ctx context.Context, rw *fhttp.Request) *fhttp.Request {
			ctx.Set("games", []string{"final-fantasy"})
			return rw
		},
	})

	record := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/names", nil)
	if err != nil {
		fatalFailed(t, "Should have created requests for '/names': %s", err)
	}
	logPassed(t, "Should have created requests for '/names'")

	drive.ServeHTTP(record, request)

	if record.Code != http.StatusOK {
		t.Logf("Expected Status: %d", http.StatusOK)
		t.Logf("Received Status: %d", record.Code)
		t.Logf("Received Body: %+q", record.Body.Bytes())
		fatalFailed(t, "Should have received success response status")
	}
	logPassed(t, "Should have received success response status")
	t.Logf("Received Body: %+q", record.Body.Bytes())

}

func TestHTTPDriveWithFractalHandlers(t *testing.T) {

	drive := fhttp.Drive(func(ctx context.Context, rw *fhttp.Request) (*fhttp.Request, error) {
		ctx.Set("names", []string{"fall-out", "reckless"})
		return rw, nil
	})(nil)

	router := fhttp.Route(drive)

	router(fhttp.Endpoint{
		Path:   "/names",
		Method: "GET",
		Action: func(ctx context.Context, rw *fhttp.Request) error {
			games, notFailed := ctx.Get("games")
			if !notFailed {
				return errors.New("Failed to retrieve games list")
			}

			names, notFailed := ctx.Get("names")
			if !notFailed {
				return errors.New("Failed to retrieve names list")
			}

			rw.Respond(http.StatusOK, map[string][]string{
				"games": games.([]string),
				"names": names.([]string),
			})

			return nil
		},
		LocalMW: func(ctx context.Context, err error, rw *fhttp.Request) *fhttp.Request {
			ctx.Set("games", []string{"final-fantasy", "dragon ball Z"})
			return rw
		},
	})

	record := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/names", nil)
	if err != nil {
		fatalFailed(t, "Should have created requests for '/names': %s", err)
	}
	logPassed(t, "Should have created requests for '/names'")

	drive.ServeHTTP(record, request)

	if record.Code != http.StatusOK {
		t.Logf("Expected Status: %d", http.StatusOK)
		t.Logf("Received Status: %d", record.Code)
		t.Logf("Received Body: %+q", record.Body.Bytes())
		fatalFailed(t, "Should have received success response status")
	}
	logPassed(t, "Should have received success response status")
	t.Logf("Received Body: %+q", record.Body.Bytes())

}

const succeedMark = "\u2713"
const failedMark = "\u2717"

func logPassed(t *testing.T, msg string, data ...interface{}) {
	t.Logf("%s %s", fmt.Sprintf(msg, data...), succeedMark)
}

func fatalFailed(t *testing.T, msg string, data ...interface{}) {
	t.Fatalf("%s %s", fmt.Sprintf(msg, data...), failedMark)
}
