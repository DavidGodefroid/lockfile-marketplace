package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "empty.bin", nil)
	fx.Raw(t, "testdata", "blank-lines.csv", []byte("\n\n\n"))
	fx.Raw(t, "testdata", "sample.txt", fx.Sample)
	fx.Raw(t, "testdata", "single-column.csv", []byte("a\nb\nc\n"))
	fx.Raw(t, "testdata", "table.csv", []byte("id,name,city\n1,Ann,Liège\n2,Bob,Namur\n"))
	fx.Raw(t, "testdata", "crlf-no-final-newline.csv", []byte("a,b\r\n1,2\r\n3,4"))
	fx.Raw(t, "testdata", "quoted.csv", []byte("a,b\n\"x, y\",\"line\nbreak\"\n\"say \"\"hi\"\"\",z\n"))
	fx.Raw(t, "testdata", "ragged.csv", []byte("a,b,c\n1,2,3\n4,5\n"))
	fx.Raw(t, "testdata", "stray-quote.csv", []byte("a,b\n1,x\"y\n"))
	fx.Raw(t, "testdata", "unterminated-quote.csv", []byte("a,b\n1,\"open\n"))
	fx.Raw(t, "testdata", "invalid-utf8.csv", []byte("a,b\n1,\xff\n"))
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
}
