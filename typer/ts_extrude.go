package typer

// extrude 复制 problematic 类型, 修正 level
//
// 过程:
//
//	递归的遍历参数 st 的类型树直到子树的 level 正确。
//	当发现一个 level 错误的类型变量 vs 时, 会使用当前参数传入的 lvl 创建 faulty 类型变量的副本 nvs, 并注册必要的约束。
//
// 原理:
//
//	nvs 的 level 比 vs 低, 满足 level 的约束。
//	需要递归的 extrude vs 的 bounds 复制到 nvs 中。
//	需要缓存已经 extruded 过的极变量避免 bound 成环。
//	总而言之，extrude 不仅复制了类型树，而且复制了以这些类型树为根的类型变量 bounds 的潜在循环子图。
func (t *Typer) extrude(st SimpleType, pol bool, lvl int) SimpleType {
	cache := polarVarMap{}

	var do func(SimpleType, bool, int) SimpleType
	do = func(st SimpleType, pol bool, lvl int) SimpleType {
		if st.level() <= lvl {
			// level 正确, 无需修复, 直接返回
			return st
		}

		// levelOf(ty) > lvl
		switch ty := st.(type) {
		case *Function:
			return Fun(do(ty.Lhs, !pol, lvl), do(ty.Rhs, pol, lvl))
		case *Tuple:
			xs := make([]SimpleType, len(ty.Elms))
			for i, el := range ty.Elms {
				xs[i] = do(el, pol, lvl)
			}
			return Tup(xs)
		case *Record:
			xs := make([]field, len(ty.Fields))
			for i, fd := range ty.Fields {
				xs[i] = field{fd.Name, do(fd.Type, pol, lvl)}
			}
			return Rcd(xs)
		case *Variable:
			pv := newPolarVar(ty, pol)
			extruded := cache.Get(pv)
			if extruded != nil {
				return extruded
			}
			// 创建 level 错误的变量副本并注册约束
			nvs := t.freshVar(lvl)
			cache.Put(pv, nvs)
			if pol {
				ty.prependUpper(nvs)
				nvs.LowerBounds = make([]SimpleType, len(ty.LowerBounds))
				for i, b := range ty.LowerBounds {
					nvs.LowerBounds[i] = do(b, pol, lvl)
				}
			} else {
				ty.prependLower(nvs)
				nvs.UpperBounds = make([]SimpleType, len(ty.UpperBounds))
				for i, b := range ty.UpperBounds {
					nvs.UpperBounds[i] = do(b, pol, lvl)
				}
			}
			return nvs
		case *Primitive:
			return ty
		default:
			panic("unreached")
		}
	}

	return do(st, pol, lvl)
}
