package typer

func merge(lhs, rhs *compactType, pol bool) *compactType {
	vars := unsortedVarSet{}
	for _, tv := range lhs.vs.Values() {
		vars.Add(tv)
	}
	for _, tv := range rhs.vs.Values() {
		vars.Add(tv)
	}

	prims := unsortedPrimSet{}
	for _, prim := range lhs.prim.Values() {
		prims.Add(prim)
	}
	for _, prim := range rhs.prim.Values() {
		prims.Add(prim)
	}

	return &compactType{
		vs:   vars.ToSorted(ASC),
		prim: prims.ToSorted(),
		rec:  mergeRec(lhs.rec, rhs.rec, pol),
		fun:  mergeFun(lhs.fun, rhs.fun, pol),
	}
}

func mergeRec(lhs, rhs *sortedNameCompactMap, pol bool) *sortedNameCompactMap {
	switch {
	case lhs != nil && rhs != nil:
		rec := nameCompactMap{}
		if pol {
			// 交集
			for _, name := range lhs.Keys() {
				lv := lhs.Get(name)
				rv := rhs.Get(name)
				if rv != nil {
					rec[name] = merge(lv, rv, pol)
				}
			}
		} else {
			// 并集
			for _, name := range lhs.Keys() {
				rec[name] = lhs.Get(name)
			}
			for _, name := range rhs.Keys() {
				lv := rec[name]
				rv := rhs.Get(name)
				if lv == nil {
					rec[name] = rv
				} else {
					rec[name] = merge(lv, rv, pol)
				}
			}
		}
		return rec.ToSorted()
	case lhs != nil && rhs == nil:
		return lhs
	case lhs == nil && rhs != nil:
		return rhs
	default:
		return nil
	}
}

func mergeFun(lhs, rhs *compactFun, pol bool) *compactFun {
	switch {
	case lhs != nil && rhs != nil:
		return &compactFun{
			merge(lhs.lhs, rhs.lhs, !pol),
			merge(lhs.rhs, rhs.rhs, pol),
		}
	case lhs != nil && rhs == nil:
		return lhs
	case lhs == nil && rhs != nil:
		return rhs
	default:
		return nil
	}
}
