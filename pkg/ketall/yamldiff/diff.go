package yamldiff

import (
	"fmt"
	"reflect"
)

const (
	nodeObject = iota
	nodeLeaf
	nodeList
	nodeRoot
	nodeMixed
)

type visit struct {
	v   uintptr
	typ reflect.Type
}

type diffTree struct {
	route       string
	depth       int
	fields      []*diffTree
	nodeType    int
	seen        map[visit]int
	left, right reflect.Value
}

func makeDiffTree(a, b interface{}) *diffTree {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	dt := &diffTree{nodeType: nodeRoot, seen: make(map[visit]int)}
	fields := dt.diff(av, bv)
	if fields == nil {
		return nil
	}
	dt.fields = fields
	return dt
}

func (t *diffTree) diff(av, bv reflect.Value) []*diffTree {
	switch {
	case !av.IsValid() && !bv.IsValid():
		return nil
	case av.IsValid() != bv.IsValid():
		return t.makeLeaf(av, bv)
	}

	at := av.Type()
	bt := bv.Type()
	if at != bt {
		return t.makeLeaf(av, bv)
	}

	switch kind := at.Kind(); kind {
	case reflect.Bool:
		if a, b := av.Bool(), bv.Bool(); a != b {
			return t.makeLeaf(av, bv)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if a, b := av.Int(), bv.Int(); a != b {
			return t.makeLeaf(av, bv)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if a, b := av.Uint(), bv.Uint(); a != b {
			return t.makeLeaf(av, bv)
		}
	case reflect.Float32, reflect.Float64:
		if a, b := av.Float(), bv.Float(); a != b {
			return t.makeLeaf(av, bv)
		}
	case reflect.String:
		if a, b := av.String(), bv.String(); a != b {
			return t.makeLeaf(av, bv)
		}
	case reflect.Complex64, reflect.Complex128:
		if a, b := av.Complex(), bv.Complex(); a != b {
			return t.makeLeaf(av, bv)
		}
	case reflect.Array:
		n := av.Len()
		var tt []*diffTree
		for i := 0; i < n; i++ {
			if item := t.makeListItem(i, av.Index(i), bv.Index(i)); item != nil {
				tt = append(tt, item)
			}
		}
		return tt
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		// todo ignore these for now
		return nil
		/*if a, b := av.Pointer(), bv.Pointer(); a != b {
			return t.makeLeaf(reflect.ValueOf(at.Name()), reflect.ValueOf(bt.Name()))
		}*/
	case reflect.Interface:
		return t.diff(av.Elem(), bv.Elem())
	case reflect.Map:
		ak, both, bk := keyDiff(av.MapKeys(), bv.MapKeys())
		var tt []*diffTree
		for _, k := range ak {
			item := &diffTree{
				route:    fmt.Sprintf("%v", k),
				nodeType: nodeLeaf,
				depth:    t.depth + 1,
				left:     av.MapIndex(k),
				right:    reflect.Zero(at),
			}
			tt = append(tt, item)
		}
		for _, k := range both {
			if item := t.makeObjItem(fmt.Sprintf("%v", k), av.MapIndex(k), bv.MapIndex(k)); item != nil {
				tt = append(tt, item)
			}
		}
		for _, k := range bk {
			item := &diffTree{
				route:    fmt.Sprintf("%v", k),
				nodeType: nodeLeaf,
				depth:    t.depth + 1,
				left:     bv.MapIndex(k),
				right:    reflect.Zero(bt),
			}
			tt = append(tt, item)
		}
		return tt
	case reflect.Ptr:
		if av.IsNil() != bv.IsNil() {
			return t.makeLeaf(av, bv)
		}
		return t.diff(av.Elem(), bv.Elem())
	case reflect.Slice:
		lenA := av.Len()
		lenB := bv.Len()
		if lenA != lenB {
			return t.makeLeaf(av, bv)
		}
		var tt []*diffTree
		for i := 0; i < lenA; i++ {
			if item := t.makeListItem(i, av.Index(i), bv.Index(i)); item != nil {
				tt = append(tt, item)
			}
		}
		return tt
	case reflect.Struct:
		var tt []*diffTree
		if t.visited(av) || t.visited(bv) {
			return nil
		}
		for i := 0; i < av.NumField(); i++ {
			if item := t.makeObjItem(at.Field(i).Name, av.Field(i), bv.Field(i)); item != nil {
				tt = append(tt, item)
			}
		}
		return tt
	default:
		panic("unknown reflect Kind: " + kind.String())
	}

	return nil
}

func (t *diffTree) makeLeaf(av reflect.Value, bv reflect.Value) []*diffTree {
	return []*diffTree{{
		nodeType: nodeLeaf,
		depth:    t.depth + 1,
		left:     av,
		right:    bv,
	}}
}

func (t *diffTree) makeListItem(i int, av reflect.Value, bv reflect.Value) *diffTree {
	li := &diffTree{
		nodeType: nodeList,
		seen:     t.seen,
		depth:    t.depth,
	}
	if fields := li.diff(av, bv); len(fields) > 0 {
		li.fields = fields
		return li
	}
	return nil
}

func (t *diffTree) makeObjItem(route string, av reflect.Value, bv reflect.Value) *diffTree {
	o := &diffTree{
		route:    route,
		nodeType: nodeObject,
		seen:     t.seen,
		depth:    t.depth + 1,
	}

	if fields := o.diff(av, bv); len(fields) > 0 {
		o.fields = fields
		return o
	}
	return nil
}

func (t *diffTree) visited(v reflect.Value) bool {
	if v.CanAddr() {
		addr := v.UnsafeAddr()
		vis := visit{addr, v.Type()}
		if vd, ok := t.seen[vis]; ok && vd < t.depth {
			return true
		}
		t.seen[vis] = t.depth
	}
	return false
}

func keyDiff(a, b []reflect.Value) (ak, both, bk []reflect.Value) {
	for _, av := range a {
		inBoth := false
		for _, bv := range b {
			if keyEqual(av, bv) {
				inBoth = true
				both = append(both, av)
				break
			}
		}
		if !inBoth {
			ak = append(ak, av)
		}
	}
	for _, bv := range b {
		inBoth := false
		for _, av := range a {
			if keyEqual(av, bv) {
				inBoth = true
				break
			}
		}
		if !inBoth {
			bk = append(bk, bv)
		}
	}
	return
}

// keyEqual compares a and b for equality.
// Both a and b must be valid map keys.
func keyEqual(av, bv reflect.Value) bool {
	if !av.IsValid() && !bv.IsValid() {
		return true
	}
	if !av.IsValid() || !bv.IsValid() || av.Type() != bv.Type() {
		return false
	}
	switch kind := av.Kind(); kind {
	case reflect.Bool:
		a, b := av.Bool(), bv.Bool()
		return a == b
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		a, b := av.Int(), bv.Int()
		return a == b
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		a, b := av.Uint(), bv.Uint()
		return a == b
	case reflect.Float32, reflect.Float64:
		a, b := av.Float(), bv.Float()
		return a == b
	case reflect.Complex64, reflect.Complex128:
		a, b := av.Complex(), bv.Complex()
		return a == b
	case reflect.Array:
		for i := 0; i < av.Len(); i++ {
			if !keyEqual(av.Index(i), bv.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Chan, reflect.UnsafePointer, reflect.Ptr:
		a, b := av.Pointer(), bv.Pointer()
		return a == b
	case reflect.Interface:
		return keyEqual(av.Elem(), bv.Elem())
	case reflect.String:
		a, b := av.String(), bv.String()
		return a == b
	case reflect.Struct:
		for i := 0; i < av.NumField(); i++ {
			if !keyEqual(av.Field(i), bv.Field(i)) {
				return false
			}
		}
		return true
	default:
		panic("invalid map key type " + av.Type().String())
	}
}
