package typer

type cstCacheSet map[string]void

func (c cstCacheSet) key(lhs, rhs SimpleType) string    { return lhs.hash() + " <: " + rhs.hash() }
func (c cstCacheSet) Add(lhs, rhs SimpleType)           { c[c.key(lhs, rhs)] = null }
func (c cstCacheSet) Contains(lhs, rhs SimpleType) bool { _, ok := c[c.key(lhs, rhs)]; return ok }

// Constraining with levels.
// 通过给 constraining 算法加入 level guard,
// 确保 higher level 的类型变量不会 escape into lower level 类型变量的 bounds

// 约束类型满足 lhs <: rhs 关系
func (t *Typer) constrain(lhs, rhs SimpleType) {
	// 避免死循环和避免重复 constrain, 降低算法复杂度
	cache := cstCacheSet{}

	var do func(SimpleType, SimpleType)
	do = func(lhs, rhs SimpleType) {
		if lhs == rhs {
			return
		}

		lVar, lIsVar := lhs.(*Variable)
		rVar, rIsVar := rhs.(*Variable)
		if lIsVar || rIsVar {
			// 没有必要缓存不涉及类型变量的子类型测试，因为只有类型变量的界可能成环
			if cache.Contains(lhs, rhs) {
				return
			}
			cache.Add(lhs, rhs)
		}

		lPrim, lIsPrim := lhs.(*Primitive)
		rPrim, rIsPrim := rhs.(*Primitive)
		if lIsPrim && rIsPrim {
			// lhs == rhs 已经处理过, prim 引用相等
			// if lPrim.Name == rPrim.Name { return }
			// int <: float
			if lPrim == Int && rPrim == Float {
				return
			}
		}

		lFun, lIsFun := lhs.(*Function)
		rFun, rIsFun := rhs.(*Function)
		if lIsFun && rIsFun {
			// 参数逆变, 返回值协变
			do(rFun.Lhs, lFun.Lhs)
			do(lFun.Rhs, rFun.Rhs)
			return
		}

		lRcd, lIsRcd := lhs.(*Record)
		rRcd, rIsRcd := rhs.(*Record)
		if lIsRcd && rIsRcd {
			lm := lRcd.fieldMap()
			// 遍历 rhs 找 lhs, 宽度子类型, e.g. {a:int,b:string} <: {a:int}
			for _, rfd := range rRcd.Fields {
				lTy, ok := lm[rfd.Name]
				if !ok {
					panic(NewTypeError("missing field: %s in %s", rfd.Name, t.show(lhs)))
				}
				rTy := rfd.Type
				// 深度子类型 e.g. {a:int} <: {a:float}
				do(lTy, rTy)
			}
			return
		}

		// 当 lhs 或 rhs 是类型变量时, 需要更新边界 (传统合一, 直接更新 subs)
		// 修改边界后, 需要迭代被约束变量的相反边界, 确保与新边界约束一致

		// 低 level 的类型变量永远不会通过边界引用高 level 类型变量
		// 加入边界的 level 一定比自己低

		// α <: rhs
		if lIsVar && rhs.level() <= lhs.level() {
			// 先更新上界, 重新约束下界
			lVar.prependUpper(rhs)
			// every lowerBound <: rhs
			for _, lb := range lVar.LowerBounds {
				do(lb, rhs)
			}
			return
		}

		// lhs <: α
		if rIsVar && lhs.level() <= rhs.level() {
			// 先更新下界, 重新约束上界
			rVar.prependLower(lhs)
			// lhs <: every upperBound
			for _, ub := range rVar.UpperBounds {
				do(lhs, ub)
			}
			return
		}

		// 下面两个 case 是存在 lhs rhs 是类型变量, but level violation
		// 通过 extrude 函数复制 problematic 类型, 直到 level violation 的类型变量及其边界
		// extrude 函数会镜像原类型的结构, 返回正确 level 的 类型

		if lIsVar {
			do(lhs, t.extrude(rhs, false, lhs.level()))
			return
		}
		if rIsVar {
			do(t.extrude(lhs, true, rhs.level()), rhs)
			return
		}

		panic(NewTypeError("cannot constrain %s <: %s", t.show(lhs), t.show(rhs)))
	}

	do(lhs, rhs)
}
