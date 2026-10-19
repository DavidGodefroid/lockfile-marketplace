package pdfx

import (
	"errors"
	"sort"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// MaxDepth bounds the nesting of inline dicts and arrays EachObject
// descends into.
const MaxDepth = 64

// ErrTooDeep is returned for an object nested deeper than MaxDepth.
var ErrTooDeep = errors.New("pdfx: object nesting too deep")

// EachObject calls fn on every object of the cross-reference table, those
// inside object streams included, each dereferenced once, and on every dict
// and array nested inline in them (a stream's dict too; stream data is never
// decoded). Indirect references are not followed: each indirect object is
// visited from the table, so reference cycles cannot loop. It stops when fn
// returns true.
func EachObject(ctx *model.Context, fn func(o types.Object) (stop bool)) (err error) {
	defer catch(&err)
	nums := make([]int, 0, len(ctx.Table))
	for n := range ctx.Table {
		nums = append(nums, n)
	}
	sort.Ints(nums)
	for _, n := range nums {
		e := ctx.Table[n]
		if n == 0 || e == nil || e.Free {
			continue
		}
		gen := 0
		if e.Generation != nil {
			gen = *e.Generation
		}
		o, err := ctx.Dereference(*types.NewIndirectRef(n, gen))
		if err != nil {
			return err
		}
		stop, err := walk(o, 1, fn)
		if err != nil || stop {
			return err
		}
	}
	return nil
}

func walk(o types.Object, depth int, fn func(types.Object) bool) (bool, error) {
	if depth > MaxDepth {
		return false, ErrTooDeep
	}
	if o == nil {
		return false, nil
	}
	if fn(o) {
		return true, nil
	}
	switch v := o.(type) {
	case types.Dict:
		for _, x := range v {
			if stop, err := walk(x, depth+1, fn); err != nil || stop {
				return stop, err
			}
		}
	case types.StreamDict:
		return walk(v.Dict, depth+1, fn)
	case types.Array:
		for _, x := range v {
			if stop, err := walk(x, depth+1, fn); err != nil || stop {
				return stop, err
			}
		}
	}
	return false, nil
}
