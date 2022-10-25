package typer

import (
	"fmt"
	"github.com/goghcrow/simple-sub/types"
	"github.com/goghcrow/simple-sub/util"
	"strconv"
	"strings"
)

// 以下类型都是内部类型推导使用

// TypeScheme
// 可能包含 universally quantified 类型变量的类型
// 可以被 instantiate 到一个特定的 level
type TypeScheme interface {
	// instantiate(t *Typer, lvl int)
	level() int
}

// PolymorphicType 多态类型
// 包含 universally quantified 类型变量的类型
// 注意: Body 中 level > _level 的类型变量才是 quantified 的
type PolymorphicType struct {
	// wraps a simple type body
	Body SimpleType

	// 并记录在哪个 _level 之上, 出现在 Body 中的类型变量是 universally quantified
	_level int
}

// SimpleType 单态类型
// 不包含 universally quantified 类型变量的类型
type SimpleType interface {
	TypeScheme
	fmt.Stringer
	hash() string
}

type (
	Function struct {
		Lhs SimpleType
		Rhs SimpleType

		_level int
		_hash  string
	}
	Primitive struct {
		Name string
	}
	Tuple struct {
		Elms []SimpleType

		_level int
		_hash  string
	}
	field struct {
		Name string
		Type SimpleType
	}
	Record struct {
		Fields []field

		_level int
		_hash  string
		_map   map[string]SimpleType
	}
	VariableState struct {
		LowerBounds []SimpleType
		UpperBounds []SimpleType
	}
	// Variable
	// 特定 PolymorphicType._level 的类型变量
	// 注意约束: 出现在 bounds 中的类型变量的 level 永远不会比变量的 _level 更高
	Variable struct {
		VariableState
		uid int

		_level   int
		_typeVar *types.TypeVariable
	}
)

// //////////////////////////////////////////////////////////////////////////////
func (v *Variable) bounds(pol bool) []SimpleType {
	if pol {
		return v.LowerBounds
	} else {
		return v.UpperBounds
	}
}

func (r *Record) fieldMap() map[string]SimpleType {
	if r._map == nil {
		r._map = make(map[string]SimpleType)
		for _, fd := range r.Fields {
			r._map[fd.Name] = fd.Type
		}
	}
	return r._map
}

func (p *PolymorphicType) instantiate(t *Typer, lvl int) SimpleType {
	return t.freshenAbove(p.level(), p.Body, lvl)
}

func (v *Variable) asTypeVar() *types.TypeVariable {
	if v._typeVar == nil {
		v._typeVar = types.TypeVar("α", v.uid)
	}
	return v._typeVar
}

func (v *Variable) prependLower(st SimpleType) {
	v.LowerBounds = append(v.LowerBounds, nil)
	copy(v.LowerBounds[1:], v.LowerBounds)
	v.LowerBounds[0] = st
}

func (v *Variable) prependUpper(st SimpleType) {
	v.UpperBounds = append(v.UpperBounds, nil)
	copy(v.UpperBounds[1:], v.UpperBounds)
	v.UpperBounds[0] = st
}

////////////////////////////////////////////////////////////////////////////////

func (p *Primitive) level() int       { return 0 }
func (v *Variable) level() int        { return v._level }
func (p *PolymorphicType) level() int { return p._level }
func (f *Function) level() int {
	if f._level < 0 {
		f._level = util.MaxInt(f.Lhs.level(), f.Rhs.level())
	}
	return f._level
}
func (t *Tuple) level() int {
	if t._level < 0 {
		t._level = 0
		for _, el := range t.Elms {
			lv := el.level()
			if lv > t._level {
				t._level = lv
			}
		}
	}
	return t._level
}
func (r *Record) level() int {
	if r._level < 0 {
		r._level = 0
		for _, fd := range r.Fields {
			lv := fd.Type.level()
			if lv > r._level {
				r._level = lv
			}
		}
	}
	return r._level
}

////////////////////////////////////////////////////////////////////////////////

// 注意 hash 与 VariableState 无关, 只与 uid 有关
func (v *Variable) hash() string  { return "var_" + strconv.Itoa(v.uid) }
func (p *Primitive) hash() string { return p.Name }
func (f *Function) hash() string {
	if f._hash == "" {
		f._hash = fmt.Sprintf("[%d]fun %s -> %s", f.level(), f.Lhs.hash(), f.Rhs.hash())
	}
	return f._hash
}
func (t *Tuple) hash() string {
	if t._hash == "" {
		t._hash = fmt.Sprintf("[%d]%s", t.level(), stringifyTuple(t, hashSimpleType))
	}
	return t._hash
}
func (r *Record) hash() string {
	if r._hash == "" {
		r._hash = fmt.Sprintf("[%d]%s", r.level(), stringifyRecord(r, hashSimpleType))
	}
	return r._hash
}

////////////////////////////////////////////////////////////////////////////////

func (p *Primitive) String() string { return p.Name }
func (v *Variable) String() string  { return fmt.Sprintf("α%d%s", v.uid, stringifyLevel(v.level())) }
func (t *Tuple) String() string     { return stringifyTuple(t, stringifySimpleType) }
func (r *Record) String() string    { return stringifyRecord(r, stringifySimpleType) }
func (f *Function) String() string  { return fmt.Sprintf("(%s -> %s)", f.Lhs, f.Rhs) }

func stringifyLevel(cnt int) string            { return strings.Repeat("'", cnt) }
func stringifySimpleType(st SimpleType) string { return st.String() }
func hashSimpleType(st SimpleType) string      { return st.hash() }

func stringifyTuple(t *Tuple, f func(t SimpleType) string) string {
	xs := make([]string, len(t.Elms))
	for i, el := range t.Elms {
		xs[i] = f(el)
	}
	return util.JoinStr(xs, ", ", "(", ")")
}
func stringifyRecord(r *Record, f func(t SimpleType) string) string {
	xs := make([]string, len(r.Fields))
	for i, fd := range r.Fields {
		xs[i] = fmt.Sprintf("%s: %s", fd.Name, f(fd.Type))
	}
	return util.JoinStr(xs, ", ", "{", "}")
}
