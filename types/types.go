package types

// 推导结果 user-facing type representations
// 其中 union, intersection, top, bottom, and recursive types 都是类型推导的展示友好类型,
// 用来化简与 pretty-print, 与推导核心算法无关。
// type variable 的上下界 约束并不是算法的输出, 最终会被编码为 union 和 intersection 类型

type Type interface {
	Show() string
	impl() *typeImpl
}

type typeImpl struct {
	Type
	typeVarsListCache []*TypeVariable
}

type (
	TopType struct {
		*typeImpl
	}
	BotType struct {
		*typeImpl
	}
	UnionType struct {
		*typeImpl
		Lhs Type
		Rhs Type
	}
	InterType struct {
		*typeImpl
		Lhs Type
		Rhs Type
	}
	FunctionType struct {
		*typeImpl
		Lhs Type
		Rhs Type
	}
	TupleType struct {
		*typeImpl
		Elms []Type
	}
	Field struct {
		Name string
		Type Type
	}
	RecordType struct {
		*typeImpl
		Fields []Field
	}
	RecursiveType struct {
		*typeImpl
		UV   *TypeVariable
		Body Type
	}
	PrimitiveType struct {
		*typeImpl
		Name string
	}
	TypeVariable struct {
		*typeImpl
		NameHint string
		hash     int
	}
)

var (
	Top Type = newTop()
	Bot Type = newBot()
)

func (t *TopType) impl() *typeImpl       { return t.typeImpl }
func (b *BotType) impl() *typeImpl       { return b.typeImpl }
func (u *UnionType) impl() *typeImpl     { return u.typeImpl }
func (i *InterType) impl() *typeImpl     { return i.typeImpl }
func (f *FunctionType) impl() *typeImpl  { return f.typeImpl }
func (r *TupleType) impl() *typeImpl     { return r.typeImpl }
func (r *RecordType) impl() *typeImpl    { return r.typeImpl }
func (r *RecursiveType) impl() *typeImpl { return r.typeImpl }
func (p *PrimitiveType) impl() *typeImpl { return p.typeImpl }
func (t *TypeVariable) impl() *typeImpl  { return t.typeImpl }
