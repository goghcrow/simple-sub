package types

import (
	"fmt"
	"github.com/goghcrow/simple-sub/util"
)

func (t *typeImpl) typeVarsList() []*TypeVariable {
	if t.typeVarsListCache != nil {
		return t.typeVarsListCache
	}

	switch ty := t.Type.(type) {
	case *TypeVariable:
		return []*TypeVariable{ty}
	case *RecursiveType:
		xs := ty.Body.impl().typeVarsList()
		nxs := make([]*TypeVariable, 1+len(xs))
		nxs[0] = ty.UV
		copy(nxs[1:], xs)
		//sort.SliceStable(nxs, func(i, j int) bool { return nxs[i].NameHint < nxs[j].NameHint })
		return nxs
	default:
		xss := make([]*TypeVariable, 0)
		for _, xs := range t.children() {
			xss = append(xss, xs.impl().typeVarsList()...)
		}
		//sort.SliceStable(xss, func(i, j int) bool { return xss[i].NameHint < xss[j].NameHint })
		return xss
	}
}

func (t *typeImpl) Show() string {
	// distinct
	set := MutTypeVarSet{}
	idx := 0
	ctx := Ctx{}
	for _, vr := range t.typeVarsList() {
		if set.Contains(vr) {
			continue
		}
		set.Add(vr)
		util.Assert(idx <= 'z'-'a', "TODO")
		ctx.Put(vr, "'"+string(rune('a'+idx)))
		idx++
	}
	return t.showIn(ctx, 0)
}

func (t *typeImpl) parensIf(str string, cnd bool) string {
	if cnd {
		return "(" + str + ")"
	}
	return str
}

func (t *typeImpl) showIn(ctx Ctx, outerPrec int) string {
	switch ty := t.Type.(type) {
	case *TopType:
		return "⊤"
	case *BotType:
		return "⊥"
	case *PrimitiveType:
		return ty.Name
	case *TypeVariable:
		return ctx.Get(ty)
	case *RecursiveType:
		body := ty.Body.impl().showIn(ctx, 31)
		return fmt.Sprintf("%s as %s", body, ctx[ty.UV.hash])
	case *FunctionType:
		lhs := ty.Lhs.impl().showIn(ctx, 11)
		rhs := ty.Rhs.impl().showIn(ctx, 10)
		return t.parensIf(fmt.Sprintf("%s -> %s", lhs, rhs), outerPrec > 10)
	case *TupleType:
		xs := make([]string, len(ty.Elms))
		for i, el := range ty.Elms {
			xs[i] = el.impl().showIn(ctx, 0)
		}
		return util.JoinStr(xs, ", ", "(", ")")
	case *RecordType:
		xs := make([]string, len(ty.Fields))
		for i, fd := range ty.Fields {
			xs[i] = fmt.Sprintf("%s: %s", fd.Name, fd.Type.impl().showIn(ctx, 0))
		}
		return util.JoinStr(xs, ", ", "{", "}")
	case *UnionType:
		lhs := ty.Lhs.impl().showIn(ctx, 20)
		rhs := ty.Rhs.impl().showIn(ctx, 20)
		return t.parensIf(fmt.Sprintf("%s ∨ %s", lhs, rhs), outerPrec > 20)
	case *InterType:
		lhs := ty.Lhs.impl().showIn(ctx, 25)
		rhs := ty.Rhs.impl().showIn(ctx, 25)
		return t.parensIf(fmt.Sprintf("%s ∧ %s", lhs, rhs), outerPrec > 25)
	default:
		panic("unreached")
	}
}

func (t *typeImpl) children() []Type {
	switch ty := t.Type.(type) {
	case *PrimitiveType:
		return []Type{}
	case *TypeVariable:
		return []Type{}
	case *TopType:
		return []Type{}
	case *BotType:
		return []Type{}
	case *FunctionType:
		return []Type{ty.Lhs, ty.Rhs}
	case *TupleType:
		xs := make([]Type, len(ty.Elms))
		for i, el := range ty.Elms {
			xs[i] = el
		}
		return xs
	case *RecordType:
		xs := make([]Type, len(ty.Fields))
		for i, fd := range ty.Fields {
			xs[i] = fd.Type
		}
		return xs
	case *UnionType:
		return []Type{ty.Lhs, ty.Rhs}
	case *InterType:
		return []Type{ty.Lhs, ty.Rhs}
	case *RecursiveType:
		return []Type{ty.Type}
	default:
		panic("unreached")
	}
}
