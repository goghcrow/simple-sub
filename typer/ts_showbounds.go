package typer

import (
	"sort"
	"strings"
)

func ShowBounds(st SimpleType) string {
	var b strings.Builder
	fst := true
	for _, v := range getVars(st) {
		lbSz := len(v.LowerBounds)
		ubSz := len(v.UpperBounds)
		if lbSz == 0 && ubSz == 0 {
			continue
		}
		if fst {
			fst = false
		} else {
			b.WriteString(", ")
		}
		b.WriteString(v.String())
		if lbSz != 0 {
			xs := make([]string, len(v.LowerBounds))
			for i, x := range v.LowerBounds {
				xs[i] = x.String()
			}
			b.WriteString(" :> ")
			b.WriteString(strings.Join(xs, " | "))
		}
		if ubSz != 0 {
			xs := make([]string, len(v.UpperBounds))
			for i, x := range v.UpperBounds {
				xs[i] = x.String()
			}
			b.WriteString(" <: ")
			b.WriteString(strings.Join(xs, " & "))
		}
	}
	return b.String()
}

func children(st SimpleType) []SimpleType {
	switch ty := st.(type) {
	case *Variable:
		xs := make([]SimpleType, len(ty.LowerBounds)+len(ty.UpperBounds))
		// [...ty.LowerBounds, ...ty.UpperBounds]
		copy(xs[copy(xs, ty.LowerBounds):], ty.UpperBounds)
		return xs
	case *Function:
		return []SimpleType{ty.Lhs, ty.Rhs}
	case *Tuple:
		xs := make([]SimpleType, len(ty.Elms))
		for i, el := range ty.Elms {
			xs[i] = el
		}
		return xs
	case *Record:
		xs := make([]SimpleType, len(ty.Fields))
		for i, fd := range ty.Fields {
			xs[i] = fd.Type
		}
		return xs
	case *Primitive:
		return []SimpleType{}
	default:
		panic("unreached")
	}
}

func getVars(st SimpleType) []*Variable {
	var res []*Variable

	set := varSet{}
	q := []SimpleType{st}
	for len(q) != 0 {
		x := q[0]
		q = q[1:]
		v, ok := x.(*Variable)
		if ok {
			if set.Contains(v) {
				continue
			}
			set.Add(v)
			res = append(res, v)
		}
		q = append(q, children(x)...)
	}

	sort.SliceStable(res, func(i, j int) bool { return res[i].uid < res[j].uid })
	return res
}
