package fractals_test

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/influx6/faux/context"
	"github.com/influx6/fractals"
)

// succeedMark is the Unicode codepoint for a check mark.
const succeedMark = "\u2713"

// failedMark is the Unicode codepoint for an X mark.
const failedMark = "\u2717"

// TestBasicFn  validates the use of reflection with giving types to test use of
// the form in fractals.
func TestBasicFn(t *testing.T) {
	var count int64

	ctx := context.New()
	pos := fractals.RLift(func(r context.Context, err error, number int) int {
		atomic.AddInt64(&count, 1)
		return number * 2
	})()

	err := errors.New("Ful")
	digit, e := pos(ctx, err, 0)
	if e == err {
		fatalFailed(t, "Should have not recieved err %s but got %s", err, e)
	}
	logPassed(t, "Should have not recieved err ")

	if digit != 0 {
		fatalFailed(t, "Should have recieved zero as number %d but got %d", 0, digit)
	}
	logPassed(t, "Should have recieved zero %d", digit)

	res, _ := pos(ctx, nil, 30)
	if res != 60 {
		fatalFailed(t, "Should have returned %d given %d", 60, 30)
	}
	logPassed(t, "Should have returned %d given %d", 60, 30)

	pos(ctx, nil, "Word") // -> This would not be seen. Has it does not match int type.

	res, _ = pos(ctx, nil, 20)
	if res != 40 {
		fatalFailed(t, "Should have returned %d given %d", 40, 20)
	}
	logPassed(t, "Should have returned %d given %d", 40, 20)

	if atomic.LoadInt64(&count) != 3 {
		fatalFailed(t, "Total processed values is not equal, expected %d but got %d", 2, count)
	}
	logPassed(t, "Total processed values was with count %d", count)
}

func TestBasicStream(t *testing.T) {
	sm := fractals.MustStream(func(ctx context.Context, number int, done bool) int {
		if done {
			return number * 400
		}

		return number * 200
	})

	dl := sm.Emit(context.New(), 4, false)
	if dll, ok := dl.(int); !ok || dll != 800 {
		fatalFailed(t, "Should have recieved 800 but got %d", dll)
	}
	logPassed(t, "Should have recieved 800")

	dl = sm.Emit(context.New(), 4, true)
	if dll, ok := dl.(int); !ok || dll != 1600 {
		fatalFailed(t, "Should have recieved 1600 but got %d", dll)
	}
	logPassed(t, "Should have recieved 1600")
}

// TestAutoFn  validates the use of reflection with giving types to test use of
// the form in fractals.
func TestAutoFn(t *testing.T) {
	var count int64

	ctx := context.New()
	pos := fractals.RLift(func(r context.Context, number int) int {
		atomic.AddInt64(&count, 1)
		return number * 2
	})()

	err := errors.New("Ful")
	_, e := pos(ctx, err, nil)
	if e != err {
		fatalFailed(t, "Should have recieved err %s but got %s", err, e)
	}
	logPassed(t, "Should have recieved err %s", err)

	res, _ := pos(ctx, nil, 30)
	if res != 60 {
		fatalFailed(t, "Should have returned %d given %d", 60, 30)
	}
	logPassed(t, "Should have returned %d given %d", 60, 30)

	pos(ctx, nil, "Word") // -> This would not be seen. Has it does not match int type.

	res, _ = pos(ctx, nil, 20)
	if res != 40 {
		fatalFailed(t, "Should have returned %d given %d", 40, 20)
	}
	logPassed(t, "Should have returned %d given %d", 40, 20)

	if atomic.LoadInt64(&count) != 2 {
		fatalFailed(t, "Total processed values is not equal, expected %d but got %d", 2, count)
	}
	logPassed(t, "Total processed values was with count %d", count)
}

// BenchmarkNodes benches the performance of using the Node api.
func BenchmarkNodes(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	ctx := context.New()
	read := fractals.RLift(func(r context.Context, number int) int {
		return number * 2
	})()

	for i := 0; i < b.N; i++ {
		read(ctx, nil, i)
	}
}

// BenchmarkNoReflect benches the performance of using the Node api.
func BenchmarkNoReflect(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	ctx := context.New()
	read := fractals.RLift(func(r context.Context, number interface{}) interface{} {
		return number.(int) * 2
	})()

	for i := 0; i < b.N; i++ {
		read(ctx, nil, i)
	}
}

func logPassed(t *testing.T, msg string, data ...interface{}) {
	t.Logf("%s %s", fmt.Sprintf(msg, data...), succeedMark)
}

func fatalFailed(t *testing.T, msg string, data ...interface{}) {
	t.Fatalf("%s %s", fmt.Sprintf(msg, data...), failedMark)
}
