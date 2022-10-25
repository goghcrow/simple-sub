package typer

import (
	"github.com/goghcrow/simple-sub/types"
)

// coalesceType 类型聚合, 将 SimpleType 转变成 types.Type
// 把出现在正极的类型变量替换成其与下界的并集
// 把出现在负极的类型变量替换成其与上界的交集
func (t *Typer) coalesceType(st SimpleType) types.Type {
	// 通过记录可能通过 bounds 引用自身的类型变量，来构造递归类型
	recursive := typeVarMap{}

	// 跟踪当前 当前极性 和 正在聚合的极变量
	var do func(SimpleType, bool, polarVarSet) types.Type
	do = func(st SimpleType, pol bool, inProcess polarVarSet) types.Type {
		switch ty := st.(type) {
		case *Primitive:
			return types.Prim(st.(*Primitive).Name)
		case *Function:
			// input positive 颠倒极性, contra-variance
			return types.Func(
				do(ty.Lhs, !pol, inProcess),
				do(ty.Rhs, pol, inProcess),
			)
		case *Tuple:
			xs := make([]types.Type, len(ty.Elms))
			for i, el := range ty.Elms {
				xs[i] = do(el, pol, inProcess)
			}
			return types.Tuple(xs)
		case *Record:
			xs := make([]types.Field, len(ty.Fields))
			for i, fd := range ty.Fields {
				ft := do(fd.Type, pol, inProcess)
				xs[i] = types.Field{Name: fd.Name, Type: ft}
			}
			return types.Record(xs)
		case *Variable:
			// 用极变量而不是变量做 key 保证只会生成"极"递归类型
			pv := newPolarVar(ty, pol)

			if inProcess.Contains(pv) {
				// 为递归的极变量生成 freshVar
				tyVar := recursive.Get(pv)
				if tyVar == nil {
					tyVar = t.freshVar(0).asTypeVar()
					recursive.Put(pv, tyVar)
				}
				return tyVar
			} else {
				inProcess.Add(pv)
				defer inProcess.Del(pv)

				// 把出现在正极的类型变量替换成其与下界的并集
				// 把出现在负极的类型变量替换成其与上界的交集
				res := types.Type(ty.asTypeVar())
				if pol {
					for _, b := range ty.LowerBounds {
						res = types.Union(res, do(b, pol, inProcess))
					}
				} else {
					for _, b := range ty.UpperBounds {
						res = types.Inter(res, do(b, pol, inProcess))
					}
				}

				rec := recursive.Get(pv)
				if rec != nil {
					// 通过记录的 freshVar 生成递归类型
					return types.Recur(rec, res)
				}

				return res
			}
		default:
			panic("unreached")
		}
	}

	return do(st, true, polarVarSet{})
}
