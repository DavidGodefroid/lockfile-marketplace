package pdfx

import (
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// HasAction reports whether some object of the file (EachObject: orphans
// and actions chained by /Next included) is an action dictionary whose /S
// is one of kinds. A dict with a /Type other than /Action is not an
// action (/S also names transition styles, page label styles, structure
// types…).
func HasAction(ctx *model.Context, kinds ...string) (bool, error) {
	found := false
	err := EachObject(ctx, func(o types.Object) bool {
		found = isAction(o, kinds)
		return found
	})
	return found, err
}

func isAction(o types.Object, kinds []string) bool {
	d, ok := o.(types.Dict)
	if !ok {
		return false
	}
	if t := d.NameEntry("Type"); t != nil && *t != "Action" {
		return false
	}
	s := d.NameEntry("S")
	if s == nil {
		return false
	}
	for _, k := range kinds {
		if *s == k {
			return true
		}
	}
	return false
}
