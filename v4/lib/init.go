// Copyright 2022 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// ~/src/modernc.org/ccorpus2/

package ccgo // import "modernc.org/ccgo/v4/lib"

import (
	"fmt"
	"math/big"
	"sort"

	"modernc.org/cc/v4"
)

type initPatch struct {
	d   *cc.Declarator
	off int64
	b   *buf
}

func (c *ctx) initializerOuter(w writer, n *cc.Initializer, t cc.Type) (r *buf) {
	a := c.initalizerFlatten(n, nil)
	// dumpInitializer(a, "")
	return c.initializer(w, n, a, t, 0, false)
}

func (c *ctx) initalizerFlatten(n *cc.Initializer, a []*cc.Initializer) (r []*cc.Initializer) {
	r = a
	switch n.Case {
	case cc.InitializerExpr: // AssignmentExpression
		return append(r, n)
	case cc.InitializerInitList: // '{' InitializerList ',' '}'
		for l := n.InitializerList; l != nil; l = l.InitializerList {
			r = append(r, c.initalizerFlatten(l.Initializer, nil)...)
		}
	default:
		c.err(errorf("internal error %T %v", n, n.Case))
	}
	return r
}

func (c *ctx) initializer(w writer, n cc.Node, a []*cc.Initializer, t cc.Type, off0 int64, arrayElem bool) (r *buf) {
	// p := n.Position()
	// if len(a) != 0 {
	// 	p = a[0].Position()
	// }
	// trc("==== (init A) typ %s off0 %#0x (%v:) (from %v: %v: %v:)", t, off0, p, origin(4), origin(3), origin(2))
	// dumpInitializer(a, "")
	// defer trc("---- (init Z) typ %s off0 %#0x (%v:)", t, off0, p)
	if cc.IsScalarType(t) {
		if len(a) == 0 {
			c.err(errorf("TODO"))
			return nil
		}

		if a[0].Offset()-off0 != 0 && a[0].Len() == 1 {
			c.err(errorf("TODO"))
			return nil
		}

		r = c.expr(w, a[0].AssignmentExpression, t, exprDefault)
		if t.Kind() == cc.Ptr && t.(*cc.PointerType).Elem().Kind() == cc.Function && c.initPatch != nil {
			c.initPatch(off0, r)
			var b buf
			b.w("(%suintptr(0))", tag(preserve))
			return &b
		}

		return r
	}

	switch x := t.(type) {
	case *cc.ArrayType:
		if len(a) == 1 && a[0].Type().Kind() == cc.Array && a[0].Value() != cc.Unknown {
			return c.expr(w, a[0].AssignmentExpression, t, exprDefault)
		}

		return c.initializerArray(w, n, a, x, off0)
	case *cc.StructType:
		if len(a) == 1 && a[0].Type().Kind() == cc.Struct {
			return c.expr(w, a[0].AssignmentExpression, t, exprDefault)
		}

		return c.initializerStruct(w, n, a, x, off0)
	case *cc.UnionType:
		if len(a) == 1 && a[0].Type().Kind() == cc.Union && a[0].Type().Size() == x.Size() {
			return c.expr(w, a[0].AssignmentExpression, t, exprDefault)
		}

		return c.initializerUnion(w, n, a, x, off0, arrayElem)
	default:
		// trc("%v: in type %v, in expr type %v, t %v", a[0].Position(), a[0].Type(), a[0].AssignmentExpression.Type(), t)
		c.err(errorf("TODO %T", x))
		return nil
	}
}

func (c *ctx) isZeroInitializerSlice(s []*cc.Initializer) bool {
	for _, v := range s {
		if !c.isZero(v.AssignmentExpression.Value()) {
			return false
		}
	}

	return true
}

func (c *ctx) isZero(v cc.Value) bool {
	switch x := v.(type) {
	case cc.Int64Value:
		return x == 0
	case cc.UInt64Value:
		return x == 0
	case cc.Float64Value:
		return x == 0
	case *cc.ZeroValue:
		return true
	case cc.Complex128Value:
		return x == 0
	case cc.Complex64Value:
		return x == 0
	case *cc.ComplexLongDoubleValue:
		return c.isZero(x.Re) && c.isZero(x.Im)
	case *cc.LongDoubleValue:
		return !(*big.Float)(x).IsInf() && (*big.Float)(x).Sign() == 0
	default:
		return false
	}
}

func (c *ctx) initializerArray(w writer, n cc.Node, a []*cc.Initializer, t *cc.ArrayType, off0 int64) (r *buf) {
	// trc("==== (array A, size %v) %s off0 %#0x (%v:)", t.Size(), t, off0, pos(n))
	// dumpInitializer(a, "")
	// trc("---- (array Z)")
	var b buf
	b.w("%s{", c.typ(n, t))
	if c.isZeroInitializerSlice(a) {
		b.w("}")
		return &b
	}

	et := t.Elem()
	esz := et.Size()
	s := sortInitializers(a, func(n int64) int64 { n -= off0; return n - n%esz })
	ranged := false
	for _, v := range s {
		if v[0].Len() != 1 {
			ranged = true
			break
		}
	}
	switch {
	case ranged:
		type expanded struct {
			s   *cc.Initializer
			off int64
		}
		m := map[int64]*expanded{}
		for _, vs := range s {
			for _, v := range vs {
				off := v.Offset() - off0
				off -= off % esz
				x := off / esz
				switch ln := v.Len(); {
				case ln != 1:
					for i := int64(0); i < ln; i++ {
						if ex, ok := m[x]; !ok || ex.s.Order() < v.Order() {
							m[x] = &expanded{v, off0 + off + i*esz}
						}
						x++
					}
				default:
					if ex, ok := m[x]; !ok || ex.s.Order() < v.Order() {
						m[x] = &expanded{v, off0 + off}
					}
				}
			}
		}
		var a []int64
		for k := range m {
			a = append(a, k)
		}
		sort.Slice(a, func(i, j int) bool { return a[i] < a[j] })
		for _, k := range a {
			v := m[k]
			b.w("\n%d: %s, ", k, c.initializer(w, n, []*cc.Initializer{v.s}, et, v.off, true))
		}
	default:
		for _, v := range s {
			v0 := v[0]
			off := v0.Offset() - off0
			off -= off % esz
			switch ln := v0.Len(); {
			case ln != 1:
				for i := int64(0); i < ln; i++ {
					b.w("\n%d: %s, ", off/esz+i, c.initializer(w, n, v, et, off0+off+i*esz, true))
				}
			default:
				off := v[0].Offset() - off0
				off -= off % esz
				b.w("\n%d: %s, ", off/esz, c.initializer(w, n, v, et, off0+off, true))
			}
		}
	}
	b.w("}")
	return &b
}

func (c *ctx) initializerStruct(w writer, n cc.Node, a []*cc.Initializer, t *cc.StructType, off0 int64) (r *buf) {
	// trc("==== %v: (struct A, size %v) %s off0 %#0x", n.Position(), t.Size(), t, off0)
	// dumpInitializer(a, "")
	// defer trc("---- %v: (struct Z, size %v) %s off0 %#0x", n.Position(), t.Size(), t, off0)
	var b buf
	b.w("%s{", c.initTyp(n, t))
	if c.isZeroInitializerSlice(a) {
		b.w("}")
		return &b
	}

	var flds []*cc.Field
	for i := 0; ; i++ {
		if f := t.FieldByIndex(i); f != nil {
			if f.Type().Size() <= 0 {
				switch x := f.Type().(type) {
				case *cc.StructType:
					if x.NumFields() != 0 {
						c.err(errorf("TODO %T", x))
						return nil
					}
				case *cc.UnionType:
					if x.NumFields() != 0 {
						c.err(errorf("TODO %T", x))
						return nil
					}
				case *cc.ArrayType:
					if x.Len() > 0 {
						c.err(errorf("TODO %T", x))
						return nil
					}
				default:
					c.err(errorf("TODO %T", x))
					return nil
				}
				continue
			}

			if f.IsBitfield() && f.ValueBits() == 0 {
				continue
			}

			flds = append(flds, f)
			// trc("appended: flds[%d] %q %s off %#0x ogo %#0x sz %#0x", len(flds)-1, f.Name(), f.Type(), f.Offset(), f.OuterGroupOffset(), f.Type().Size())
			continue
		}

		break
	}
	s := sortInitializers(a, func(off int64) int64 {
		off -= off0
		i := sort.Search(len(flds), func(i int) bool {
			return flds[i].OuterGroupOffset() >= off
		})
		if i < len(flds) && flds[i].OuterGroupOffset() == off {
			return off
		}

		return flds[i-1].OuterGroupOffset()
	})
	// trc("==== initializers (A)")
	// for i, v := range s {
	// 	for j, w := range v {
	// 		if w.Field() == nil {
	// 			trc("%d.%d: %q off %v, %s", i, j, "", w.Offset(), cc.NodeSource(w.AssignmentExpression))
	// 			continue
	// 		}

	// 		trc("%d.%d: %q off %v foff %v, fogo %v, %s", i, j, w.Field().Name(), w.Offset(), w.Field().Offset(), w.Field().OuterGroupOffset(), cc.NodeSource(w.AssignmentExpression))
	// 	}
	// }
	// trc("==== initializers (Z)")
	for _, v := range s {
		first := v[0]
		off := first.Offset() - off0
		// trc("first.Offset() %#0x, off %#0x", first.Offset(), off)
		for off > flds[0].Offset()+flds[0].Type().Size()-1 {
			// trc("skip %q off %#0x", flds[0].Name(), flds[0].Offset())
			flds = flds[1:]
			if len(flds) == 0 {
				panic(todo("", n.Position()))
			}
		}
		f := flds[0]
		if f.IsBitfield() {
			// trc("==== %v: TODO bitfield", cpos(n))
			// trc("%q %s off %#0x ogo %#0x sz %#0x", f.Name(), f.Type(), f.Offset(), f.OuterGroupOffset(), f.Type().Size())
			// for i, v := range v {
			// 	trc("%d: %q %s", i, v.Field().Name(), cc.NodeSource(v.AssignmentExpression))
			// }
			// trc("----")
			for len(flds) != 0 && flds[0].OuterGroupOffset() == f.OuterGroupOffset() {
				flds = flds[1:]
			}
			b.w("%s__ccgo%d: ", tag(field), f.OuterGroupOffset())
			sort.Slice(v, func(i, j int) bool {
				a, b := v[i].Field(), v[j].Field()
				return a.Offset()*8+int64(a.OffsetBits()) < b.Offset()*8+int64(b.OffsetBits())
			})
			ogo := f.OuterGroupOffset()
			gsz := 8 * (int64(f.GroupSize()) + f.Offset() - ogo)
			for i, in := range v {
				if i != 0 {
					b.w("|")
				}
				f = in.Field()
				sh := f.OffsetBits() + 8*int(f.Offset()-ogo)
				b.w("(((%suint%d(%s))&%#0x)<<%d)", tag(preserve), gsz, c.expr(w, in.AssignmentExpression, nil, exprDefault), uint(1)<<f.ValueBits()-1, sh)
			}
			b.w(", ")
			continue
		}

		for isEmpty(v[0].Type()) {
			v = v[1:]
		}
		// trc("f %q %s off %#0x v[0].Type() %v", f.Name(), f.Type(), f.Offset(), v[0].Type())
		flds = flds[1:]
		b.w("%s%s: %s, ", tag(field), c.fieldName(t, f), c.initializer(w, n, v, f.Type(), off0+f.Offset(), false))
	}
	b.w("}")
	return &b
}

func (c *ctx) initializerUnion(w writer, n cc.Node, a []*cc.Initializer, t *cc.UnionType, off0 int64, arrayElem bool) (r *buf) {
	// trc("==== %v: (union A, size %v) %s off0 %#0x, arrayElem %v", n.Position(), t.Size(), t, off0, arrayElem)
	// dumpInitializer(a, "")
	// trc("---- (union Z)")
	var b buf
	if c.isZeroInitializerSlice(a) {
		b.w("%s{}", c.typ(n, t))
		return &b
	}

	switch t.NumFields() {
	case 0:
		c.err(errorf("%v: cannot initialize empty union", n.Position()))
	case 1:
		b.w("%s{%s%s: %s}", c.typ(n, t), tag(field), c.fieldName(t, t.FieldByIndex(0)), c.initializer(w, n, a, t.FieldByIndex(0).Type(), off0, false))
		return &b
	}

	switch len(a) {
	case 1:
		b.w("(*(*%s)(%sunsafe.%sPointer(&struct{ ", c.typ(n, t), tag(importQualifier), tag(preserve))
		b.w("%s", c.initializerUnionOne(w, n, a, t, off0))
		b.w(")))")
	default:
		b.w("(*(*%s)(%sunsafe.%sPointer(&", c.typ(n, t), tag(importQualifier), tag(preserve))
		b.w("%s", c.initializerUnionMany(w, n, a, t, off0, arrayElem))
		b.w(")))")
	}
	return &b
}

func (c *ctx) initializerUnionMany(w writer, n cc.Node, a []*cc.Initializer, t *cc.UnionType, off0 int64, arrayElem bool) (r *buf) {
	// trc("==== %v: (union many A.%v, size %v) type %s off0 %#0x, arrayElem %v", n.Position(), c.pass, t.Size(), t, off0, arrayElem)
	// dumpInitializer(a, "")
	// trc("---- (union many Z)")
	var arrayElemOff int64
	if arrayElem {
		arrayElemOff = off0 - off0%t.Size()
		// trc("adjust %#0x(%[1]v)", arrayElemOff)
	}
	var b buf
	var paths [][]*cc.Initializer
	for _, v := range a {
		var path []*cc.Initializer
		for p := v.Parent(); p != nil; p = p.Parent() {
			path = append(path, p)
		}
		paths = append(paths, path)
	}
	// for _, v := range paths {
	// 	var a []string
	// 	for _, w := range v {
	// 		a = append(a, w.Type().String())
	// 	}
	// 	trc("path A %q", a)
	// }
	// var common []*cc.Initializer
	var lca *cc.Initializer
	for {
		var path *cc.Initializer
		for i, v := range paths {
			if len(v) == 0 {
				goto done
			}

			w := v[len(v)-1]
			if i == 0 {
				path = w
				continue
			}

			if w != path {
				goto done
			}
		}
		lca = path
		if lca.Type() == t {
			goto done
		}

		// common = append(common, lca)
		for i, v := range paths {
			paths[i] = v[:len(v)-1]
		}
	}
done:
	// for _, v := range paths {
	// 	var a []string
	// 	for _, w := range v {
	// 		a = append(a, w.Type().String())
	// 	}
	// 	trc("path Z %q", a)
	// }
	// var aa []string
	// for _, v := range common {
	// 	aa = append(aa, v.Type().String())
	// }
	// trc("common %q", aa)
	if lca == nil {
		trc("MANY %v: (%v:)", pos(n), origin(1))
		b.w("/* %v: TODO */", origin(1))
		return &b
	}

	// trc("lca0 %s, size %v, off %v", lca.Type(), lca.Type().Size(), lca.Offset())
	lcaType, lcaOff := c.fixLCA(t, lca, a, off0)
	if lcaType == nil {
		trc("MANY %v: (%v:)", pos(n), origin(1))
		b.w("/* %v: TODO */", origin(1))
		return &b
	}

	// trc("lcaType %s, size %v, lcaOff %v", lcaType, lcaType.Size(), lcaOff)
	if lcaType.Size() == t.Size() {
		// trc("%v: size ok, initializer(%s off0 %v)", n.Position(), lcaType, off0)
		return c.initializer(w, n, a, lcaType, off0, false)
	}

	pre := lcaOff - arrayElemOff
	post := t.Size() - lcaType.Size() - pre
	b.w("struct{")
	if lcaOff != 0 {
		// trc("pre %v", pre)
		b.w("%s_ [%d]byte;", tag(preserve), pre)
	}
	b.w("%sf ", tag(preserve))
	b.w("%s ", c.typ(n, lcaType))
	if post != 0 {
		// trc("post %v", post)
		b.w("; %s_ [%d]byte", tag(preserve), post)
	}
	b.w("}{%sf: ", tag(preserve))
	// trc("size not ok, initializer(%s off0 %v)", lcaType, off0)
	b.w("%s", c.initializer(w, n, a, lcaType, off0, false))
	b.w("}")
	return &b
}

func (c *ctx) fixLCA(t *cc.UnionType, lca *cc.Initializer, a []*cc.Initializer, off0 int64) (rt cc.Type, off int64) {
	// trc("fixLCA off0 %v\n%5d t   %s\n%5d lca %s", off0, t.Size(), t, lca.Type().Size(), lca.Type())
	rt = lca.Type()
	switch {
	case rt.Size() > t.Size():
		// trc("too big")
	case rt != t:
		// ;trc("size ok and types are different")
		return rt, lca.Offset()
	}

	// trc("self or big, search")
	okField, okName := true, true
	for _, v := range a {
		if v.Field() == nil {
			okField = false
			okName = false
			// trc("nofield")
			break
		}

		if v.Field().Name() == "" {
			// trc("noname")
			okName = false
			break
		}
	}

	// trc("", okField, okName)
	if okField && okName {
	nextUf:
		for i := 0; i < t.NumFields(); i++ {
			uf := t.FieldByIndex(i)
			// trc("%d: %q %s", i, uf.Name(), uf.Type())
			for _, v := range a {
				af := v.Field()
				// trc("%T", af)
				x, ok := uf.Type().(interface{ FieldByName(string) *cc.Field })
				if !ok {
					// trc("continue uf cannot FieldByName")
					continue nextUf
				}

				f := x.FieldByName(af.Name())
				if f == nil {
					// trc("continue no field %q", af.Name())
					continue nextUf
				}

				// ufOff := uf.Offset()
				// ufSize := uf.Type().Size()
				fOff := f.Offset()
				vOff := v.Offset()
				// trc("ufOff %v, ufSize %v, fOff %v, vOff %v", ufOff, ufSize, fOff, vOff)
				if !(vOff-off0 == fOff && f.Type().Size() == v.Type().Size()) {
					// trc("continue bad off")
					continue nextUf
				}
			}
			return uf.Type(), lca.Offset() + uf.Offset()
		}
	}

	f := t.FieldByIndex(0)
	// trc("uf[0] %q %s, %v", f.Name(), f.Type(), f.Type().Size())
	return f.Type(), f.Offset()
}

func (c *ctx) initializerUnionOne(w writer, n cc.Node, a []*cc.Initializer, t *cc.UnionType, off0 int64) (r *buf) {
	var b buf
	in := a[0]
	pre := in.Offset() - off0
	if pre != 0 {
		b.w("%s_ [%d]byte;", tag(preserve), pre)
	}
	b.w("%sf ", tag(preserve))
	f := in.Field()
	switch {
	case f != nil && f.IsBitfield():
		b.w("%suint%d", tag(preserve), f.AccessBytes()*8)
	default:
		b.w("%s ", c.typ(n, in.Type()))
	}
	if post := t.Size() - (pre + in.Type().Size()); post != 0 {
		b.w("; %s_ [%d]byte", tag(preserve), post)
	}
	b.w("}{%sf: ", tag(preserve))
	switch f := in.Field(); {
	case f != nil && f.IsBitfield():
		b.w("(((%suint%d(%s))&%#0x)<<%d)", tag(preserve), f.AccessBytes()*8, c.expr(w, in.AssignmentExpression, nil, exprDefault), uint(1)<<f.ValueBits()-1, f.OffsetBits())
	default:
		b.w("%s", c.expr(w, in.AssignmentExpression, in.Type(), exprDefault))
	}
	b.w("}")
	return &b
}

func sortInitializers(a []*cc.Initializer, group func(int64) int64) (r [][]*cc.Initializer) {
	// [0]6.7.8/23: The order in which any side effects occur among the
	// initialization list expressions is unspecified.
	m := map[int64][]*cc.Initializer{}
	for _, v := range a {
		off := group(v.Offset())
		m[off] = append(m[off], v)
	}
	for _, v := range m {
		sort.Slice(v, func(i, j int) bool {
			a, b := v[i].Offset(), v[j].Offset()
			if a < b {
				return true
			}

			if a > b {
				return false
			}

			c, d := v[i].Field(), v[j].Field()
			if c == nil || d != nil {
				return false
			}

			return c.Index() < d.Index()
		})
		r = append(r, v)
	}
	sort.Slice(r, func(i, j int) bool { return r[i][0].Offset() < r[j][0].Offset() })
	return r
}

//lint:ignore U1000 debug helper
func dumpInitializer(a []*cc.Initializer, pref string) {
	for _, v := range a {
		var t string
		for p := v.Parent(); p != nil; p = p.Parent() {
			switch d := p.Type().Typedef(); {
			case d != nil:
				t = fmt.Sprintf("[%s].", d.Name()) + t
			default:
				switch x, ok := p.Type().(interface{ Tag() cc.Token }); {
				case ok:
					tag := x.Tag()
					t = fmt.Sprintf("[%s].", tag.SrcStr()) + t
				default:
					t = fmt.Sprintf("[%s].", p.Type()) + t
				}
			}
		}
		var fs string
		if f := v.Field(); f != nil {
			var ps string
			for p := f.Parent(); p != nil; p = p.Parent() {
				ps = ps + fmt.Sprintf("{%q %v}", p.Name(), p.Type())
			}
			fs = fmt.Sprintf(
				" %s(field %q, IsBitfield %v, Offset %v, OffsetBits %v, OuterGroupOffset %v, InOverlapGroup %v, Mask %#0x, ValueBits %v)",
				ps, f.Name(), f.IsBitfield(), f.Offset(), f.OffsetBits(), f.OuterGroupOffset(), f.InOverlapGroup(), f.Mask(), f.ValueBits(),
			)
		}
		switch v.Case {
		case cc.InitializerExpr:
			fmt.Printf("%s %v: order %v off %#05x '%s' %s type %q <- %s%s\n", pref, pos(v.AssignmentExpression), v.Order(), v.Offset(), cc.NodeSource(v.AssignmentExpression), t, v.Type(), v.AssignmentExpression.Type(), fs)
		case cc.InitializerInitList:
			if v.InitializerList != nil {
				if uf := v.InitializerList.UnionField(); uf != nil {
					fmt.Printf("%s· union field %q %s\n", pref, uf.Name(), uf.Type())
				}
			}
			s := pref + "· " + fs
			for l := v.InitializerList; l != nil; l = l.InitializerList {
				dumpInitializer([]*cc.Initializer{l.Initializer}, s)
			}
		}
	}
}
