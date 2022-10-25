package typer

import "github.com/goghcrow/simple-sub/util"

var (
	Bool   = Prim("bool")
	Int    = Prim("int")
	Float  = Prim("float")
	String = Prim("string")
)

var primitives = map[string]*Primitive{}

func Prim(name string) *Primitive {
	p := primitives[name]
	if p == nil {
		p = &Primitive{Name: name}
		primitives[name] = p
	}
	return p
}

func Fun(lhs SimpleType, rhs SimpleType) *Function { return &Function{Lhs: lhs, Rhs: rhs, _level: -1} }
func Tup(elms []SimpleType) *Tuple                 { return &Tuple{Elms: elms, _level: -1} }
func Rcd(fields []field) *Record                   { return &Record{Fields: fields, _level: -1} }

func PolyType(lvl int, body SimpleType) *PolymorphicType {
	return &PolymorphicType{Body: body, _level: lvl}
}

func Funx(lhs []SimpleType, rhs SimpleType) *Function {
	argc := len(lhs)
	util.Assert(argc >= 1, "expect at least one param")
	f := Fun(lhs[argc-1], rhs)
	for i := argc - 2; i >= 0; i-- {
		f = Fun(lhs[i], f)
	}
	return f
}

func Var(uid int, lvl int, lowerBounds []SimpleType, upperBounds []SimpleType) *Variable {
	return &Variable{
		VariableState: VariableState{
			LowerBounds: lowerBounds,
			UpperBounds: upperBounds,
		},
		_level: lvl,
		uid:    uid,
	}
}
