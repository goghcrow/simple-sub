package typer

import (
	"github.com/goghcrow/simple-sub/terms"
)

// 使用内部表示的 compactType 进行类型推导
// 首先推断出的 SimpleType 值然后转化为 compactType 值方便进行简化。
// 最后使用 coalesceCompactType 将 compactType 转变成 types.Type
//
// SimpleType = typer.inferType(term)
// compactTypeScheme = typer.canonicalizeType(SimpleType)
// compactTypeScheme = typer.simplifyType(compactTypeScheme)
// types.Type = coalesceCompactType(compactTypeScheme)

type Typer struct {
	freshCount int
}

func NewTyper() *Typer {
	return &Typer{}
}

func (t *Typer) uuid() int { t.freshCount++; return t.freshCount - 1 }
func (t *Typer) freshVar(lvl int) *Variable {
	uid := t.uuid()
	tv := Var(uid, lvl, []SimpleType{}, []SimpleType{})
	return tv
}

func (t *Typer) Builtins() *Ctx {
	return NewCtx(map[string]TypeScheme{
		"true":  Bool,
		"false": Bool,
		"not":   Fun(Bool, Bool),
		"succ":  Fun(Int, Int),
		"add":   Fun(Int, Fun(Int, Int)),
		// ∀𝛼, 𝛽. bool → 𝛼 → 𝛽 → 𝛼 ⊔ 𝛽
		// ∀𝛼. bool → 𝛼 → 𝛼 → 𝛼
		"if": func() *PolymorphicType {
			tv := t.freshVar(1) // 类型变量的 level 要大于 polyType 的 level
			return PolyType(0, Funx([]SimpleType{Bool, tv, tv}, tv))
		}(),
	})
}

func (t *Typer) inferTypes(pgrm *terms.Program, ctx *Ctx) (res []*PolymorphicType, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = TypeErrorOf(r)
		}
	}()
	res = make([]*PolymorphicType, len(pgrm.Defs))
	for i, def := range pgrm.Defs {
		res[i] = t.typeLetRhs(def, ctx, 0)
		ctx.Add(def.Name, res[i])
	}
	return
}

func (t *Typer) inferType(term terms.Term, ctx *Ctx) SimpleType {
	return t.typeTerm(term, ctx, 0)
}

func (t *Typer) show(st SimpleType) string {
	return t.coalesceType(st).Show()
}
