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
	fx.Raw(t, "testdata", "safe.zip", archive("a.txt", "dir/b.txt", "dir/..c/d..txt", "ab:c.txt"))
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
	fx.Raw(t, "testdata", "dotdot-slash.zip", archive("a.txt", "dir/../../outside.txt"))
	fx.Raw(t, "testdata", "dotdot-backslash.zip", archive(`dir\..\..\outside.txt`))
	fx.Raw(t, "testdata", "dotdot-only.zip", archive(".."))
	fx.Raw(t, "testdata", "absolute-slash.zip", archive("/etc/passwd"))
	fx.Raw(t, "testdata", "absolute-backslash.zip", archive(`\Windows\win.ini`))
	fx.Raw(t, "testdata", "drive-letter.zip", archive("C:evil.txt"))
	fx.Raw(t, "testdata", "docx-traversal.docx", fx.Docx(fx.Paragraph, fx.File("../evil.xml", body)))
}
