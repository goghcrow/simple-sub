package typer

import "github.com/goghcrow/simple-sub/terms"

// 注意这里 会将 rhs(let-body) 的类型 wrap 成 PolymorphicType,
// PolymorphicType 简单包装 SimpleType 的 let-body, 并额外记录 above which level
// 出现在 let-body 类型变量会被认为是 universally quantified

// 注意 let 需要 level + 1, 低于当前 level 的 type var 能逃逸当前环境的约束
func (t *Typer) typeLetRhs(let *terms.Declaration, ctx *Ctx, lvl int) *PolymorphicType {
	if let.Rec {
		// 为 let-binding rhs 在 context 绑定一个类型变量, 之后检查( constrain )其为 实际的 rhs 类型的 supertype
		eTy := t.freshVar(lvl + 1)
		ctx = ctx.Extend(let.Name, eTy)
		ty := t.typeTerm(let.Rhs, ctx, lvl+1)
		t.constrain(ty, eTy)
		return PolyType(lvl, eTy)
	} else {
		ty := t.typeTerm(let.Rhs, ctx, lvl+1)
		return PolyType(lvl, ty)
	}
}

// 类型推导
// 找到程序的所有子类型约束(subtyping constraints), 递归传播约束直到类型变量, 并通过改变 bound 来约束类型变量
// 核心函数, 除了 constrain 与传统 HM 合一类似
// 根据上下文推导出 term 的 SimpleType, 其中约束函数作为补充, 把一个类型约束为另一个类型的子类型, 否则报错
func (t *Typer) typeTerm(term terms.Term, ctx *Ctx, lvl int) SimpleType {
	switch tm := term.(type) {
	case *terms.LiteralBool:
		return Bool
	case *terms.LiteralInt:
		return Int
	case *terms.LiteralFloat:
		return Float
	case *terms.LiteralString:
		return String
	case *terms.Variable:
		varTy := ctx.MustLookup(tm.Name)
		// HM 因为 let 多态的原因, var 的需要 instantiate
		return t.instantiate(varTy, lvl)
	case *terms.Lambda:
		param := t.freshVar(lvl)
		nctx := ctx.Extend(tm.Name, param)
		body := t.typeTerm(tm.Rhs, nctx, lvl)
		return Fun(param, body)
	case *terms.Application:
		funLhs := t.typeTerm(tm.Lhs, ctx, lvl)
		arg := t.typeTerm(tm.Rhs, ctx, lvl)
		res := t.freshVar(lvl)
		funRhs := Fun(arg, res)
		// 约束 funLhs <: funRhs, 参数逆变, 返回值协变
		// e.g. int <: float
		// float -> int <: int -> int
		// int -> int <: int -> float
		// float -> int <: int -> float
		t.constrain(funLhs, funRhs)
		return res
	case *terms.Selection:
		rcdLhs := t.typeTerm(tm.Recv, ctx, lvl)
		fd := t.freshVar(lvl)
		rcdRhs := Rcd([]field{{tm.FieldName, fd}})
		// record.field
		// record <: record {field: T}
		// lhs receiver 是一个必须包含 field 字段的记录类型
		// e.g. {a:1}.a  => {a:int} <: {a:int}
		// {a:1,b:"s"}.a =>  {a:int, b:string} <: {a:int}
		t.constrain(rcdLhs, rcdRhs)
		return fd
	case *terms.Tuple:
		xs := make([]SimpleType, len(tm.Elms))
		for i, el := range tm.Elms {
			xs[i] = t.typeTerm(el, ctx, lvl)
		}
		return Tup(xs)
	case *terms.Record:
		xs := make([]field, len(tm.Fields))
		for i, fd := range tm.Fields {
			xs[i] = field{fd.Name, t.typeTerm(fd.Term, ctx, lvl)}
		}
		return Rcd(xs)
	case *terms.LetDefine:
		nTy := t.typeLetRhs(&tm.Declaration, ctx, lvl)
		nctx := ctx.Extend(tm.Name, nTy)
		return t.typeTerm(tm.Body, nctx, lvl)
	default:
		panic("unreached")
	}
}

// PolymorphicType 的 instantiate(lvl) 方法会复制 body
// 并把 above level 的类型变量替换为 level lvl 的 fresh variables (freshenAbove 的工作)
func (t *Typer) instantiate(ty TypeScheme, lvl int) SimpleType {
	if p, ok := ty.(*PolymorphicType); ok {
		return p.instantiate(t, lvl)
	}
	return ty.(SimpleType)
}
