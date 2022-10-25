package types

func newTop() *TopType {
	t := &TopType{}
	t.typeImpl = &typeImpl{Type: t}
	return t
}
func newBot() *BotType {
	t := &BotType{}
	t.typeImpl = &typeImpl{Type: t}
	return t
}
func Union(lhs Type, rhs Type) *UnionType {
	t := &UnionType{Lhs: lhs, Rhs: rhs}
	t.typeImpl = &typeImpl{Type: t}
	return t
}
func Inter(lhs Type, rhs Type) *InterType {
	t := &InterType{Lhs: lhs, Rhs: rhs}
	t.typeImpl = &typeImpl{Type: t}
	return t
}
func Func(lhs Type, rhs Type) *FunctionType {
	t := &FunctionType{Lhs: lhs, Rhs: rhs}
	t.typeImpl = &typeImpl{Type: t}
	return t
}
func Tuple(elms []Type) *TupleType {
	t := &TupleType{Elms: elms}
	t.typeImpl = &typeImpl{Type: t}
	return t
}
func Record(fields []Field) *RecordType {
	t := &RecordType{Fields: fields}
	t.typeImpl = &typeImpl{Type: t}
	return t
}
func Recur(UV *TypeVariable, body Type) *RecursiveType {
	t := &RecursiveType{UV: UV, Body: body}
	t.typeImpl = &typeImpl{Type: t}
	return t
}
func Prim(name string) *PrimitiveType {
	t := &PrimitiveType{Name: name}
	t.typeImpl = &typeImpl{Type: t}
	return t
}
func TypeVar(nameHint string, hash int) *TypeVariable {
	t := &TypeVariable{NameHint: nameHint, hash: hash}
	t.typeImpl = &typeImpl{Type: t}
	return t
}
