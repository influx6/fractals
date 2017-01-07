package fs_test

import (
	"testing"

	"github.com/influx6/faux/context"
	"github.com/influx6/fractals"
	"github.com/influx6/fractals/fs"
)

// succeedMark is the Unicode codepoint for a check mark.
const succeedMark = "\u2713"

// failedMark is the Unicode codepoint for an X mark.
const failedMark = "\u2717"

func TestReadDirPath(t *testing.T) {
	var lists []string
	items := fractals.RLift(func(ctx context.Context, list []string) {
		lists = list
	})(fs.ReadDirPath(), fs.SkipStat(fs.IsDir), fs.UnwrapStats(), fs.ResolvePath())

	items(context.New(), nil, "../../..")

	if len(lists) < 1 {
		t.Fatalf("%s Expected a list of directories", failedMark)
	}

	t.Logf("%s Expected a list of directories", succeedMark)
}

func TestReadDir(t *testing.T) {

	var lists []string

	items := fractals.RLift(func(ctx context.Context, list []string) {
		lists = list
	})(fs.ReadDir("../../.."), fs.SkipStat(fs.IsDir), fs.UnwrapStats(), fs.ResolvePath())

	items(context.New(), nil, "")

	if len(lists) < 1 {
		t.Fatalf("%s Expected a list of directories", failedMark)
	}

	t.Logf("%s Expected a list of directories", succeedMark)
}

func BenchmarkFileList(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	ctx := context.New()

	for i := 0; i > b.N; i++ {
		fractals.RLift(fractals.IdentityHandler())(fs.ReadDir("../../.."), fs.SkipStat(fs.IsDir), fs.UnwrapStats(), fs.ResolvePath())(ctx, nil, "")
	}
}

func BenchmarkReadDirPath(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	ctx := context.New()

	for i := 0; i > b.N; i++ {
		fractals.RLift(fractals.IdentityHandler())(fs.ReadDirPath(), fs.SkipStat(fs.IsDir), fs.UnwrapStats(), fs.ResolvePath())(ctx, nil, "../../..")
	}
}
