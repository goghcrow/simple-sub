package typer

import (
	"github.com/goghcrow/simple-sub/types"
)

// Simplifier Pass3 : Coalesce

// 将一个 compactTypeScheme coalesce 成一个 types.Type，
// 同时执行 hash-consing，尽可能将递归类型变得紧凑
func (t *Typer) coalesceCompactType(cty *compactTypeScheme) types.Type {
	var do func(compactTypeOrVariable, bool, polarCompactMap) types.Type
	do = func(ctv compactTypeOrVariable, pol bool, inProcess polarCompactMap) types.Type {
		pc := newPolarCompact(ctv, pol)

		thunk := inProcess.Get(pc)
		if thunk != nil {
			res := thunk.(typeVarThunk)()
			log("[coalesceCompactType.REC][%t] %v -> %v", pol, ctv, res)
			return res
		}

		isRecursive := false

		// lazyV
		v := func() func() *types.TypeVariable {
			var cache *types.TypeVariable
			return func() *types.TypeVariable {
				if cache == nil {
					isRecursive = true
					tv, ok := ctv.(*Variable)
					if ok {
						cache = tv.asTypeVar()
					} else {
						cache = t.freshVar(0).asTypeVar()
					}
				}
				return cache
			}
		}()

		inProcess.Put(pc, typeVarThunk(v))
		defer inProcess.Del(pc)

		var res types.Type
		switch ty := ctv.(type) {
		case *Variable:
			compact := cty.recVars.Get(ty)
			if compact == nil {
				res = ty.asTypeVar()
			} else {
				res = do(compact, pol, inProcess)
			}
		case *compactType:
			var lst []types.Type
			for _, vs := range ty.vs.Values() {
				lst = append(lst, do(vs, pol, inProcess))
			}
			for _, prim := range ty.prim.Values() {
				lst = append(lst, types.Prim(prim.Name))
			}
			if ty.rec != nil {
				xs := make([]types.Field, ty.rec.Len())
				for i, name := range ty.rec.Keys() {
					fty := do(ty.rec.Get(name), pol, inProcess)
					xs[i] = types.Field{Name: name, Type: fty}
				}
				rt := types.Record(xs)
				lst = append(lst, rt)
			}
			if ty.fun != nil {
				ft := types.Func(do(ty.fun.lhs, !pol, inProcess), do(ty.fun.rhs, pol, inProcess))
				lst = append(lst, ft)
			}
			res = mergeTypes(lst, pol)
		}

		if isRecursive {
			return types.Recur(v(), res)
		} else {
			return res
		}
	}

	return do(cty.term, true, polarCompactMap{})
}

func mergeTypes(lst []types.Type, pol bool) types.Type {
	if len(lst) == 0 {
		if pol {
			return types.Bot
		} else {
			return types.Top
		}
	}

	res := lst[0]
	if pol {
		for i := 1; i < len(lst); i++ {
			res = types.Union(res, lst[i])
		}
	} else {
		for i := 1; i < len(lst); i++ {
			res = types.Inter(res, lst[i])
		}
	}
	return res
}
