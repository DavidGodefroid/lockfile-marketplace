package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

const (
	plainRTF  = `{\rtf1\ansi\deff0{\fonttbl{\f0 Sample;}}\pard\plain sample\par}`
	objectRTF = `{\rtf1\ansi{\object\objemb{\*\objclass Sample}\objw100\objh100{\*\objdata 0102}}}`
)

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "plain.rtf", []byte(plainRTF))
	fx.Raw(t, "testdata", "header-only.rtf", []byte(`{\rtf`))
	fx.Raw(t, "testdata", "object.rtf", []byte(objectRTF))
	fx.Raw(t, "testdata", "object-linked.rtf", []byte(`{\rtf1{\object\objautlink\rsltpict{\*\objdata 01}}}`))
	fx.Raw(t, "testdata", "object-at-eof.rtf", []byte(`{\rtf1 \object`))
	fx.Raw(t, "testdata", "object-param.rtf", []byte(`{\rtf1 \object1 sample}`))
	fx.Raw(t, "testdata", "object-after-bin.rtf", []byte("{\\rtf1{\\pict\\bin3 \x00\\x}\\object}"))
	fx.Raw(t, "testdata", "escaped-backslash.rtf", []byte(`{\rtf1 sample \\object\par}`))
	fx.Raw(t, "testdata", "longer-word.rtf", []byte(`{\rtf1 \objectx \objemb sample}`))
	fx.Raw(t, "testdata", "in-bin-data.rtf", []byte(`{\rtf1{\pict\bin7 \object}}`))
	fx.Raw(t, "testdata", "bin-truncated.rtf", []byte(`{\rtf1\bin99 \object}`))
	fx.Raw(t, "testdata", "bin-huge.rtf", []byte(`{\rtf1\bin99999999999999999999 \object}`))
}
