package typer

// Simplifier Pass2 : SimplifyType

// 思路：如果一个类型变量 'a 总是与类型变量 'b 同时出现在正极，这意味着两者不可区分，因此可以被合并，反之亦然。
//   Ex: ('a & 'b) -> ('a, 'b) 等同 'a -> ('a, 'a)
//   Ex: ('a & 'b) -> 'b -> ('a, 'b) 不同 'a -> 'a -> ('a, 'a)
//	 	没有人任何 'a 的值可以使 'a -> 'a -> ('a, 'a) <: (a & b) -> b -> (a, b)
//		我们需要的 'a :> b | a & b <: a & b 并不是合法边界
//	 Ex: 'a -> 'b -> 'a | 'b 等同 'a -> 'a -> 'a
//	 理由：另一个变量 'b 总是可以被看作是 'a & 'b（resp. a | b）而不损失信息。
//	 	事实上，在正极我们会有 'a <: 'a & 'b，'b <: 'a & 'b，而在负极，我们总是有 'a 和 'b 在一起，即，'a & 'b

// 另一个思路：移除那些总是与其他类型的变量同时出现的正极和负极的变量。
// 会产生约束：'a :> Int <: 'b 和 'b <: Int（等同于 'a =:= 'b =:= Int)
//	 Ex: 'a ∧ Int -> 'a ∨ Int 等同于 Int -> Int
// 	 目前只对 Primitive type 做上述处理。理论上也可以对 Function 和 Record 做同样的处理
//   注意：从概念上来说，这个思路包含了删除只出现正极或负极变量的化简思路。
//   事实上，如果 'a 没有出现在正极，就好像它总是和底类型一起出现在正极和负极,所以可以用底类型代替'a。

func (t *Typer) simplifyType(cty *compactTypeScheme) *compactTypeScheme {
	// analysis phase 收集的状态
	allVars := unsortedVarSet{}
	for _, v := range cty.recVars.Keys() {
		allVars.Add(v)
	}
	recVars := newVarCompactThunkMap()
	coOccurs := coOccurrences{} // var -> list[coOccur]

	// analysis phase 之后填充,影响 reconstruction phase
	varSubst := varVarMap{} // var -> substVar 注意 value 可能为 nil

	// 遍历类型进行分析, 返回之后 reconstruct 的 thunk
	var do func(*compactType, bool) compactTypeThunk
	do = func(ty *compactType, pol bool) compactTypeThunk {
		for _, vs := range ty.vs.Values() {
			allVars.Add(vs)

			// 计算极变量 pv 的共现
			{
				pv := newPolarVar(vs, pol)
				// 与 vs 共现的类型
				newOccs := newLinkedSimpleTypeSet()
				for _, tv1 := range ty.vs.Values() {
					newOccs.Add(tv1)
				}
				for _, prim := range ty.prim.Values() {
					newOccs.Add(prim)
				}

				// 计算新旧交集
				oldOccs := coOccurs.Get(pv)
				if oldOccs == nil {
					log("[simplifyType.do]add %t %v", pol, newOccs)
					coOccurs.Put(pv, newOccs)
				} else {
					// 计算交集
					interOccs := newLinkedSimpleTypeSet()
					for _, st := range oldOccs.Values() {
						if newOccs.Contains(st) {
							interOccs.Add(st)
						}
					}
					coOccurs.Put(pv, interOccs)
				}
			}

			// 如果 `vs`递归, 处理他的边界 `b`
			b := cty.recVars.Get(vs)
			if b != nil {
				if recVars.Get(vs) == nil {
					// 确保在 recursing 之前注册 `vs`避免死循环
					thunk := func() *compactType { return do(b, pol)() }
					recVars.Put(vs, thunk)
					thunk()
				}
			}
		}

		var recThunk *sortedNameCompactThunkMap
		if ty.rec != nil {
			rec := nameCompactThunkMap{}
			for _, name := range ty.rec.Keys() {
				rec[name] = do(ty.rec.Get(name), pol)
			}
			recThunk = rec.ToSorted()
		}

		var lhsThunk compactTypeThunk
		var rhsThunk compactTypeThunk
		if ty.fun != nil {
			lhsThunk = do(ty.fun.lhs, !pol)
			rhsThunk = do(ty.fun.rhs, pol)
		}

		return func() *compactType {
			newVars := unsortedVarSet{}
			for _, tv := range ty.vs.Values() {
				sub, ok := varSubst.Get(tv)
				if !ok {
					newVars.Add(tv)
				} else if sub != nil {
					newVars.Add(sub)
				}
			}

			var rec *sortedNameCompactMap
			if recThunk != nil {
				m := nameCompactMap{}
				for _, name := range recThunk.Keys() {
					m[name] = recThunk.Get(name)()
				}
				rec = m.ToSorted()
			}

			var fun *compactFun
			if lhsThunk != nil {
				fun = &compactFun{
					lhs: lhsThunk(),
					rhs: rhsThunk(),
				}
			}

			return &compactType{
				vs:   newVars.ToSorted(ASC),
				prim: ty.prim,
				rec:  rec,
				fun:  fun,
			}
		}
	}

	gone := do(cty.term, true)

	log("[simplifyType.do.term]%v", cty.term)
	log("[simplifyType.do.occ]%v", coOccurs)
	log("[simplifyType.do.rec]%v", recVars)

	// 化简掉那些只出现在正极或负极的非递归变量
	sortedAllVars := allVars.ToSorted(DESC)
	for _, tv := range sortedAllVars.Values() {
		if recVars.Get(tv) == nil {
			lhs := coOccurs.Get(newPolarVar(tv, true))
			rhs := coOccurs.Get(newPolarVar(tv, false))
			if (lhs != nil && rhs == nil) || (lhs == nil && rhs != nil) {
				log("[simplifyType.!] %v", tv)
				varSubst.Put(tv, nil)
			} else if lhs == nil && rhs == nil {
				panic("assert")
			}
		}
	}

	// 根据极性共现分析，合并等价变量
	pols := []bool{true, false}
	for _, v := range sortedAllVars.Values() {
		log("[simplifyType.vv] %v", v)
		if varSubst.Contains(v) {
			continue
		}

		if DBG {
			log("[simplifyType.v] %v | %v %v", v,
				coOccurs.Get(newPolarVar(v, true)),
				coOccurs.Get(newPolarVar(v, false)),
			)
		}

		for _, pol := range pols {
			pv := newPolarVar(v, pol)
			vOccurs := coOccurs.Get(pv)
			if vOccurs == nil {
				continue
			}

			for _, st := range vOccurs.Values() {
				switch w := st.(type) {
				case *Variable:
					if w == v || varSubst.Contains(w) {
						continue
					}
					bt := recVars.Get(v)
					ct := recVars.Get(w)
					if !bt.Equals(ct) {
						continue
					}

					wOccurs := coOccurs.Get(newPolarVar(w, pol))
					log("[simplifyType.w] %v %v", w, wOccurs)
					// 避免合并 rec 和非 rec 变量，因为非 rec 变量可能不是严格意义上的极变量 [test:T1]
					if wOccurs != nil && !wOccurs.Contains(v) {
						continue
					}

					log("[simplifyType.U] %v = %v", w, v)
					varSubst.Put(w, v) // unify w into v

					// 由于合并了 w 和 v, 如果它们是递归的，我们需要合并它们的边界，否则就合并 v 和 w 的其他共现的极性(!pol)。
					// 例如,
					//  考虑如果我们合并 v 和 w  `(v & w) -> v & x -> w -> x`
					//  得到 `v -> v & x -> v -> x`
					// 	然后 the old positive co-occ of v 应该变更为 {v,x} & {w,v} == {v}!
					if ct == nil {
						// ^ `w`不是递归， `v`也不是 (参见 bt.Equals(ct))
						wOccurs1 := coOccurs.Get(newPolarVar(w, !pol))
						// ^  这必须被定义，否则我们就已经把非 rec 变量给简化了
						pv1 := newPolarVar(v, !pol)
						replace := newLinkedSimpleTypeSet()
						for _, ty := range coOccurs.Get(pv1).Values() {
							if ty == v || wOccurs1.Contains(ty) {
								replace.Add(ty)
							}
						}
						coOccurs.Put(pv1, replace)
					} else {
						// `w` 是递归变量, `v`也是 (参见 bt.Equals(ct))
						// 递归类型需要有严格的极性
						if coOccurs.Get(newPolarVar(w, !pol)) != nil {
							panic("assert")
						}
						// w 已经合并到 v, 移除 w
						recVars.Del(w)
						// `v`是递归的，所以`recVars(v)` 被定义，并记录 v 的新的递归约束。
						recVars.Put(v, func() *compactType { return merge(bt(), ct(), pol) })
					}
				case *Primitive:
					// v 两级共现都有 w, 干掉 v
					vOccurs1 := coOccurs.Get(newPolarVar(v, !pol))
					if vOccurs1 != nil {
						if vOccurs1.Contains(w) {
							varSubst.Put(v, nil)
						}
					}
				}
			}
		}
	}

	log("[simplifyType.sub] %v", varSubst)
	term := gone()
	log("[simplifyType.gone] %v", term)

	varCmtMap := newVarCompactMap()
	sortedRecVars := recVars.ToSorted()
	for _, v := range sortedRecVars.Keys() {
		compact := sortedRecVars.Get(v)()
		varCmtMap.Put(v, compact)
	}
	return &compactTypeScheme{term, varCmtMap.ToSorted()}
}
