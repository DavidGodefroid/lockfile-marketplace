package pdfx

import (
	"errors"
	"strings"
	"unicode"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// Namespaces of the XMP properties the plugins test for.
const (
	PDFUAIDNamespace = "http://www.aiim.org/pdfua/ns/id/"
	PDFXIDNamespace  = "http://www.npes.org/pdfx/ns/id/"
	DCNamespace      = "http://purl.org/dc/elements/1.1/"
)

var (
	// ErrWrongType is returned when a catalog or information entry is
	// present with a type the specification does not allow.
	ErrWrongType = errors.New("pdfx: entry has the wrong type")
	// ErrBadText is returned when a text string cannot be decoded.
	ErrBadText = errors.New("pdfx: text string does not decode")
)

// CatalogDict returns the dereferenced catalog entry key, nil when it is
// absent or null, ErrWrongType when it is not a dictionary.
func CatalogDict(ctx *model.Context, key string) (d types.Dict, err error) {
	defer catch(&err)
	o, err := catalogEntry(ctx, key)
	if err != nil || o == nil {
		return nil, err
	}
	d, ok := o.(types.Dict)
	if !ok {
		return nil, ErrWrongType
	}
	return d, nil
}

// CatalogArray returns the dereferenced catalog entry key, nil when it is
// absent or null, ErrWrongType when it is not an array.
func CatalogArray(ctx *model.Context, key string) (a types.Array, err error) {
	defer catch(&err)
	o, err := catalogEntry(ctx, key)
	if err != nil || o == nil {
		return nil, err
	}
	a, ok := o.(types.Array)
	if !ok {
		return nil, ErrWrongType
	}
	return a, nil
}

// CatalogNonEmptyString reports whether the catalog entry key is a text
// string that is not empty or white space only. The value is not returned.
func CatalogNonEmptyString(ctx *model.Context, key string) (ok bool, err error) {
	defer catch(&err)
	o, err := catalogEntry(ctx, key)
	if err != nil || o == nil {
		return false, err
	}
	return nonEmptyText(o)
}

// HasCatalogMetadata reports whether the catalog has a /Metadata stream,
// without decoding it.
func HasCatalogMetadata(ctx *model.Context) (ok bool, err error) {
	defer catch(&err)
	o, err := catalogEntry(ctx, "Metadata")
	if err != nil || o == nil {
		return false, err
	}
	if _, ok := o.(types.StreamDict); !ok {
		return false, errNotXMP
	}
	return true, nil
}

// InfoHas reports whether the document information dictionary has the
// entry key with a non-null value.
func InfoHas(ctx *model.Context, key string) (ok bool, err error) {
	defer catch(&err)
	o, err := infoEntry(ctx, key)
	return o != nil, err
}

// InfoHasNonEmpty reports whether the document information entry key is a
// text string that is not empty or white space only. The value is not
// returned.
func InfoHasNonEmpty(ctx *model.Context, key string) (ok bool, err error) {
	defer catch(&err)
	o, err := infoEntry(ctx, key)
	if err != nil || o == nil {
		return false, err
	}
	return nonEmptyText(o)
}

func catalogEntry(ctx *model.Context, key string) (types.Object, error) {
	cat, err := ctx.Catalog()
	if err != nil {
		return nil, err
	}
	o, found := cat.Find(key)
	if !found || o == nil {
		return nil, nil
	}
	return ctx.Dereference(o)
}

func infoEntry(ctx *model.Context, key string) (types.Object, error) {
	if ctx.Info == nil {
		return nil, nil
	}
	info, err := ctx.DereferenceDict(*ctx.Info)
	if err != nil {
		return nil, ErrWrongType
	}
	if info == nil {
		return nil, nil
	}
	o, found := info.Find(key)
	if !found || o == nil {
		return nil, nil
	}
	return ctx.Dereference(o)
}

// nonEmptyText decodes a text string (PDFDocEncoding or UTF-16 with a byte
// order mark) and reports whether it has a character other than white
// space. Only that bit leaves the function.
func nonEmptyText(o types.Object) (bool, error) {
	var s string
	var err error
	switch v := o.(type) {
	case types.StringLiteral:
		s, err = types.StringLiteralToString(v)
	case types.HexLiteral:
		s, err = types.HexLiteralToString(v)
	default:
		return false, ErrWrongType
	}
	if err != nil {
		return false, ErrBadText
	}
	return strings.IndexFunc(s, func(r rune) bool { return !unicode.IsSpace(r) && r != byteOrderMark }) >= 0, nil
}

// byteOrderMark may survive the decoding of a UTF-16 string.
const byteOrderMark = 0xFEFF
