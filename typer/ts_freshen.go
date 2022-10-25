package typer

// 即 PolymorphicType 的 instantiate(lvl) 方法, 递归实例化所有 universally quantified 类型变量
//
// 复制 PolymorphicType.Body 并把 above level 的类型变量替换为 lvl 的 fresh variables
// > level 的类型变量: forall, universally quantified 变量, 需要实例化
// <= level, 引用环境中的类型变量, 不能实例化
// lim = PolymorphicType.level, lvl 当前 level
func (t *Typer) freshenAbove(lim int, st SimpleType, lvl int) SimpleType {
	freshened := varVarMap{}

	var freshen func(SimpleType) SimpleType
	freshen = func(st SimpleType) SimpleType {
		// 引用环境中的类型变量, 不能实例化
		if st.level() <= lim {
			return st
		}

		switch ty := st.(type) {
		case *Primitive:
			return ty
		case *Function:
			return Fun(freshen(ty.Lhs), freshen(ty.Rhs))
		case *Tuple:
			xs := make([]SimpleType, len(ty.Elms))
			for i, el := range ty.Elms {
				xs[i] = freshen(el)
			}
			return Tup(xs)
		case *Record:
			xs := make([]field, len(ty.Fields))
			for i, fd := range ty.Fields {
				xs[i] = field{fd.Name, freshen(fd.Type)}
			}
			return Rcd(xs)
		case *Variable:
			vs, ok := freshened.Get(ty)
			if ok {
				return vs
			}

			nvs := t.freshVar(lvl)
			freshened.Put(ty, nvs)

			// 正序遍历会导致了不同的 freshVar 创建顺序
			// 进而导致了一些类型在被放入 let 绑定的 RHS 中时不会被化简

			sz := len(ty.LowerBounds)
			nvs.LowerBounds = make([]SimpleType, sz)
			for i := sz - 1; i >= 0; i-- {
				nvs.LowerBounds[i] = freshen(ty.LowerBounds[i])
			}

			sz = len(ty.UpperBounds)
			nvs.UpperBounds = make([]SimpleType, sz)
			for i := sz - 1; i >= 0; i-- {
				nvs.UpperBounds[i] = freshen(ty.UpperBounds[i])
			}
			return nvs
		default:
			panic("unreached")
		}
	}

	return freshen(st)
}
