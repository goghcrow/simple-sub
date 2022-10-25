package typer

// Simplifier Pass1 : Canonicalize

// 类似 `compactType`函数 将 SimpleType 转换为 compactType ,
// 不同的是会确保产生一个 canonicalized 的 compactType, 即：所有共现的递归类型会被合并的类型，
// -- 如果它们有不同的周期长度，会创建一个新的递归类型，周期长度是各自原始周期长度的LCD。
// 要做到这一点，在 compacting 类型的同时，我们要跟踪所遍历的 compact type（而不是像 `compactType`函数 那样，仅仅跟踪单个变量）。
// 这需要一个交错的两阶段过程，首先过渡性地合并源类型外层的所有共现变量的边界，然后进一步遍历产生的变量集。
// 这个 "展开递归类型直到它们对齐 "的过程类似于将 NFA 变成 DFA 的 powerset 构造，
// 理论上会产生一个指数级的巨大类型，实际该算法有很好的化简效果。

func (t *Typer) canonicalizeType(st SimpleType) *compactTypeScheme {

	// 将一个SimpleType的最外层变成一个CompactType，不对类型变量进行转换
	var do0 func(SimpleType, bool) *compactType
	do0 = func(st SimpleType, pol bool) *compactType {
		cty := emptyCompactType()
		switch ty := st.(type) {
		case *Primitive:
			cty.prim = onePrimSet(ty)
		case *Function:
			cty.fun = &compactFun{do0(ty.Lhs, !pol), do0(ty.Rhs, pol)}
		case *Tuple:
			cty.tup = make([]*compactType, len(ty.Elms))
			for i, el := range ty.Elms {
				cty.tup[i] = do0(el, pol)
			}
		case *Record:
			rec := nameCompactMap{}
			for _, fd := range ty.Fields {
				rec[fd.Name] = do0(fd.Type, pol)
			}
			cty.rec = rec.ToSorted()
		case *Variable:
			cty.vs = closeOver(unsortedVarSet{}, unsortedVarSet{ty.uid: ty}, pol).ToSorted(ASC)
		default:
			panic("unreached")
		}
		return cty
	}

	recVars := newVarCompactMap()
	recursive := polarCompactMap{}

	// 合并所有给定的 compactType 的类型变量的边界，并遍历结果
	var do1 func(*compactType, bool, polarCompactSet) *compactType
	do1 = func(ty *compactType, pol bool, inProcess polarCompactSet) *compactType {
		if ty.isEmpty() {
			return ty
		}

		pc := newPolarCompact(ty, pol)

		if inProcess.Contains(pc) {
			tv := recursive.Get(pc)
			if tv == nil {
				tv = t.freshVar(0)
				recursive.Put(pc, tv)
			}
			return &compactType{
				vs:   oneVarSet(tv.(*Variable)),
				prim: emptyPrimSet(),
			}
		}

		inProcess.Add(pc)
		defer inProcess.Del(pc)

		res := ty
		for _, tv := range ty.vs.Values() {
			for _, b := range tv.bounds(pol) {
				if _, ok := b.(*Variable); !ok {
					// do1 递归做的事情主要是这一行
					res = merge(res, do0(b, pol), pol)
				}
			}
		}
		adapted := &compactType{
			vs:   res.vs,
			prim: res.prim,
		}
		if res.rec != nil {
			m := nameCompactMap{}
			for _, name := range res.rec.Keys() {
				m[name] = do1(res.rec.Get(name), pol, inProcess)
			}
			adapted.rec = m.ToSorted()
		}
		if res.fun != nil {
			adapted.fun = &compactFun{
				lhs: do1(res.fun.lhs, !pol, inProcess),
				rhs: do1(res.fun.rhs, pol, inProcess),
			}
		}

		r := recursive.Get(pc)
		if r == nil {
			return adapted
		} else {
			tv := r.(*Variable)
			recVars.Put(tv, adapted)
			return &compactType{
				vs:   oneVarSet(tv),
				prim: emptyPrimSet(),
			}
		}
	}

	cty := do0(st, true)
	term := do1(cty, true, polarCompactSet{})
	return &compactTypeScheme{term, recVars.ToSorted()}
}

func closeOver(done unsortedVarSet, todo unsortedVarSet, pol bool) unsortedVarSet {
	if len(todo) == 0 {
		return done
	}

	newDone := unsortedVarSet{}
	for _, v := range done {
		newDone.Add(v)
	}
	for _, v := range todo {
		newDone.Add(v)
	}

	newTodo := unsortedVarSet{}
	for _, v := range todo {
		for _, b := range v.bounds(pol) {
			if v1, ok := b.(*Variable); ok {
				if !newDone.Contains(v1) {
					newTodo.Add(v1)
				}
			}
		}
	}

	return closeOver(newDone, newTodo, pol)
}
