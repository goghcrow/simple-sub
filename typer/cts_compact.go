package typer

// 已经替换成 canonicalize
// 需要把推导结果的类型转换为 compact type 合并类型变量的边界，才能更准确的进行共现分析和hash-consing
// e.g. {x: A}, {x: B; y: C} 转换边界 {x: A ∧ B; y: C}

func (t *Typer) compactType(st SimpleType) *compactTypeScheme {
	recVars := newVarCompactMap()
	recursive := polarVarMap{}

	// `parents` 缓存 bounds 被 compact 的变量，以消除假循环，如 ?a<: ?b 和 ?b<: ?a，并不是实际递归类型
	var do func(SimpleType, bool, varSet, polarVarSet) *compactType
	do = func(st SimpleType, pol bool, parents varSet, inProcess polarVarSet) *compactType {
		switch ty := st.(type) {
		case *Primitive:
			return &compactType{
				vs:   emptyVarSet(),
				prim: onePrimSet(ty),
			}
		case *Function:
			return &compactType{
				fun: &compactFun{
					lhs: do(ty.Lhs, !pol, varSet{}, inProcess),
					rhs: do(ty.Rhs, pol, varSet{}, inProcess),
				},
				vs:   emptyVarSet(),
				prim: emptyPrimSet(),
			}
		case *Tuple:
			tup := make([]*compactType, len(ty.Elms))
			for i, el := range ty.Elms {
				tup[i] = do(el, pol, varSet{}, inProcess)
			}
			return &compactType{
				tup:  tup,
				vs:   emptyVarSet(),
				prim: emptyPrimSet(),
			}
		case *Record:
			rec := nameCompactMap{}
			for _, fd := range ty.Fields {
				rec[fd.Name] = do(fd.Type, pol, varSet{}, inProcess)
			}
			return &compactType{
				rec:  rec.ToSorted(),
				vs:   emptyVarSet(),
				prim: emptyPrimSet(),
			}
		case *Variable:
			pv := newPolarVar(ty, pol)
			if inProcess.Contains(pv) {
				if parents.Contains(ty) {
					return emptyCompactType()
				}
				// get or update
				tv := recursive.Get(pv)
				if tv == nil {
					tv = t.freshVar(0)
					recursive.Put(pv, tv)
				}
				return &compactType{
					vs:   oneVarSet(tv),
					prim: emptyPrimSet(),
				}
			} else {
				inProcess.Add(pv)
				defer inProcess.Del(pv)

				parents.Add(ty)
				defer parents.Del(ty)

				bound := &compactType{
					vs:   oneVarSet(ty),
					prim: emptyPrimSet(),
				}
				for _, b := range ty.bounds(pol) {
					bound = merge(bound, do(b, pol, parents, inProcess), pol)
				}

				tv := recursive.Get(pv)
				if tv == nil {
					return bound
				}

				recVars.Put(tv, bound)
				return &compactType{
					vs:   oneVarSet(tv),
					prim: emptyPrimSet(),
				}
			}
		default:
			panic("unreached")
		}
	}

	term := do(st, true, varSet{}, polarVarSet{})
	return &compactTypeScheme{term, recVars.ToSorted()}
}
