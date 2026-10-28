package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

func writeFixtures(t *testing.T) {
	body := []byte("sample body\n")
	archive := func(names ...string) []byte {
		var es []fx.Entry
		for _, n := range names {
			es = append(es, fx.File(n, body))
		}
		return fx.Zip(es...)
	}
	fx.Raw(t, "testdata", "empty-archive.zip", fx.Zip())
	fx.Raw(t, "testdata", "unique.zip", archive("a.txt", "dir/a.txt", "dir/"))
	fx.Raw(t, "testdata", "case-differs.zip", archive("a.txt", "A.txt"))
	fx.Raw(t, "testdata", "path-differs.zip", archive("dir/a.txt", "dir//a.txt", "./dir/a.txt"))
	fx.Raw(t, "testdata", "duplicate.zip", archive("a.txt", "b.txt", "a.txt"))
	fx.Raw(t, "testdata", "duplicate-dir.zip", archive("dir/", "dir/"))
	fx.Raw(t, "testdata", "duplicate-stored.zip", fx.Zip(
		fx.File("a.txt", body),
		fx.Entry{Name: "a.txt", Body: body},
	))
}
