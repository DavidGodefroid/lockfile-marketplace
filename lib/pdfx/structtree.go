package pdfx

import (
	"errors"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

var (
	// ErrStructTreeTooDeep is returned for a structure tree deeper than
	// maxTreeDepth.
	ErrStructTreeTooDeep = errors.New("pdfx: structure tree too deep")
	// ErrStructTreeCycle is returned when a structure element is its own
	// descendant.
	ErrStructTreeCycle = errors.New("pdfx: structure tree has a cycle")
	// ErrRoleMapCycle is returned when following the /RoleMap from a
	// structure type comes back to it.
	ErrRoleMapCycle = errors.New("pdfx: role map has a cycle")
)

// EachStructElem calls fn on every structure element reachable from the
// /K of the catalog's /StructTreeRoot, depth first, until fn returns true.
// role is the element's /S mapped through the root's /RoleMap until it
// names a type the map does not list or maps to itself ("" when /S is
// absent); the element's
// own /S is left for fn to read. Marked-content and object references, and
// MCIDs, are not elements and are skipped. An element reached twice is
// visited once. An untagged file visits nothing. A tree deeper than
// maxTreeDepth, a cycle, a role map cycle or an /S that is not a name is an
// error.
func EachStructElem(ctx *model.Context, fn func(elem types.Dict, role string) (stop bool)) (err error) {
	defer catch(&err)
	root, err := CatalogDict(ctx, "StructTreeRoot")
	if err != nil || root == nil {
		return err
	}
	w := structWalker{ctx: ctx, fn: fn, seen: map[int]bool{}, path: map[int]bool{}}
	if o, ok := root.Find("RoleMap"); ok {
		if w.roles, err = ctx.DereferenceDict(o); err != nil {
			return err
		}
	}
	k, ok := root.Find("K")
	if !ok {
		return nil
	}
	_, err = w.kids(k, 1)
	return err
}

type structWalker struct {
	ctx   *model.Context
	fn    func(types.Dict, string) bool
	roles types.Dict
	seen  map[int]bool // elements already visited
	path  map[int]bool // elements on the current branch
}

// kids walks a /K value: one kid or an array of kids.
func (w *structWalker) kids(k types.Object, depth int) (bool, error) {
	o, err := w.ctx.Dereference(k)
	if err != nil {
		return false, err
	}
	a, ok := o.(types.Array)
	if !ok {
		return w.elem(k, depth)
	}
	for _, kid := range a {
		if stop, err := w.elem(kid, depth); err != nil || stop {
			return stop, err
		}
	}
	return false, nil
}

// elem visits one kid if it is a structure element, then its own kids.
func (w *structWalker) elem(kid types.Object, depth int) (bool, error) {
	if depth > maxTreeDepth {
		return false, ErrStructTreeTooDeep
	}
	if ref, ok := kid.(types.IndirectRef); ok {
		num := ref.ObjectNumber.Value()
		if w.path[num] {
			return false, ErrStructTreeCycle
		}
		if w.seen[num] {
			return false, nil
		}
		w.seen[num] = true
		w.path[num] = true
		defer delete(w.path, num)
	}
	o, err := w.ctx.Dereference(kid)
	if err != nil {
		return false, err
	}
	d, ok := o.(types.Dict)
	if !ok {
		return false, nil
	}
	if t := d.NameEntry("Type"); t != nil && *t != "StructElem" {
		return false, nil
	}
	role, err := w.role(d)
	if err != nil {
		return false, err
	}
	if w.fn(d, role) {
		return true, nil
	}
	k, ok := d.Find("K")
	if !ok {
		return false, nil
	}
	return w.kids(k, depth+1)
}

// role returns the element's /S mapped through the role map.
func (w *structWalker) role(d types.Dict) (string, error) {
	o, ok := d.Find("S")
	if !ok {
		return "", nil
	}
	o, err := w.ctx.Dereference(o)
	if err != nil {
		return "", err
	}
	if o == nil {
		return "", nil
	}
	s, ok := o.(types.Name)
	if !ok {
		return "", ErrWrongType
	}
	role := s.Value()
	for range len(w.roles) + 1 {
		m, ok := w.roles.Find(role)
		if !ok {
			return role, nil
		}
		m, err := w.ctx.Dereference(m)
		if err != nil {
			return "", err
		}
		n, ok := m.(types.Name)
		if !ok {
			return "", ErrWrongType
		}
		if n.Value() == role {
			return role, nil
		}
		role = n.Value()
	}
	return "", ErrRoleMapCycle
}
