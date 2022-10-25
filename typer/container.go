package typer

import (
	"fmt"
	"github.com/goghcrow/simple-sub/types"
	"reflect"
	"sort"
)

type void *struct{}

var null void = &struct{}{}

type Ord int

const (
	ASC Ord = iota
	DESC
)

////////////////////////////////////////////////////////////////////////////////

type polarVar struct {
	tv       *Variable
	polarity bool // true Positive, false Negative

	_hash string
}

func newPolarVar(tv *Variable, pol bool) *polarVar {
	return &polarVar{tv, pol, fmt.Sprintf("%d_%t", tv.uid, pol)}
}

type polarCompact struct {
	ctv      compactTypeOrVariable
	polarity bool

	_hash string
}

func newPolarCompact(ctv compactTypeOrVariable, pol bool) *polarCompact {
	return &polarCompact{ctv: ctv, polarity: pol, _hash: ctv.polHash(pol)}
}

func (v *Variable) polHash(polarity bool) string    { return fmt.Sprintf("%d_%t", v.uid, polarity) }
func (c *compactType) polHash(polarity bool) string { return fmt.Sprintf("%s_%t", c.hash(), polarity) }

////////////////////////////////////////////////////////////////////////////////

// varVarMap Map[ Variable, Variable]
type varVarMap map[int]*Variable

func (m varVarMap) Contains(k *Variable) bool         { _, ok := m[k.uid]; return ok }
func (m varVarMap) Get(k *Variable) (*Variable, bool) { v, ok := m[k.uid]; return v, ok }
func (m varVarMap) Put(k, v *Variable)                { m[k.uid] = v }

////////////////////////////////////////////////////////////////////////////////

var (
	_emptyVarSet  = &sortedVarSet{m: unsortedVarSet{}, keys: []*Variable{}}
	_emptyPrimSet = &sortedPrimSet{m: unsortedPrimSet{}, keys: []*Primitive{}}
)

////////////////////////////////////////////////////////////////////////////////

type unsortedVarSet map[int]*Variable

func (u unsortedVarSet) Add(v *Variable)           { u[v.uid] = v }
func (u unsortedVarSet) Contains(v *Variable) bool { _, ok := u[v.uid]; return ok }
func (u unsortedVarSet) ToSorted(ord Ord) *sortedVarSet {
	cl := make(unsortedVarSet, len(u))
	keys := make([]*Variable, 0, len(u))
	for k, v := range u {
		cl[k] = v
		keys = append(keys, v)
	}
	if ord == DESC {
		sort.Slice(keys, func(i, j int) bool { return keys[i].uid > keys[j].uid })
	} else {
		sort.Slice(keys, func(i, j int) bool { return keys[i].uid < keys[j].uid })
	}
	return &sortedVarSet{m: cl, keys: keys}
}

////////////////////////////////////////////////////////////////////////////////

type unsortedPrimSet map[string]*Primitive

func (u unsortedPrimSet) Add(p *Primitive) { u[p.Name] = p }
func (u unsortedPrimSet) ToSorted() *sortedPrimSet {
	cl := make(unsortedPrimSet, len(u))
	keys := make([]*Primitive, 0, len(u))
	for k, v := range u {
		cl[k] = v
		keys = append(keys, v)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].Name < keys[j].Name })
	return &sortedPrimSet{m: cl, keys: keys}
}

////////////////////////////////////////////////////////////////////////////////

// sortedVarSet 没必要有序, but go-map 实现的 set 会引起最终结果不稳定
type sortedVarSet struct {
	m    unsortedVarSet
	keys []*Variable
}

func oneVarSet(v *Variable) *sortedVarSet {
	return &sortedVarSet{unsortedVarSet{v.uid: v}, []*Variable{v}}
}
func emptyVarSet() *sortedVarSet { return _emptyVarSet }
func (s *sortedVarSet) Len() int { return len(s.keys) }
func (s *sortedVarSet) Values() []*Variable {
	cl := make([]*Variable, len(s.keys))
	copy(cl, s.keys)
	return cl
}

////////////////////////////////////////////////////////////////////////////////

// sortedPrimSet 没必要有序, but go-map 实现的 set 会引起最终结果不稳定
type sortedPrimSet struct {
	m    unsortedPrimSet
	keys []*Primitive
}

func onePrimSet(p *Primitive) *sortedPrimSet {
	return &sortedPrimSet{unsortedPrimSet{p.Name: p}, []*Primitive{p}}
}
func emptyPrimSet() *sortedPrimSet { return _emptyPrimSet }
func (s *sortedPrimSet) Len() int  { return len(s.keys) }
func (s *sortedPrimSet) Values() []*Primitive {
	cl := make([]*Primitive, len(s.keys))
	copy(cl, s.keys)
	return cl
}

////////////////////////////////////////////////////////////////////////////////

// varCompactMap Map[ Variable, compactType]
type varCompactMap struct {
	m    map[int]*compactType
	vars []*Variable
}

func newVarCompactMap() *varCompactMap { return &varCompactMap{map[int]*compactType{}, []*Variable{}} }
func (m *varCompactMap) Put(v *Variable, c *compactType) {
	_, ok := m.m[v.uid]
	m.m[v.uid] = c
	if !ok {
		m.vars = append(m.vars, v)
	}
}
func (m *varCompactMap) ToSorted() *sortedVarCompactMap {
	if len(m.m) != len(m.vars) {
		panic("assert")
	}
	clv := make([]*Variable, len(m.vars))
	copy(clv, m.vars)
	sort.Slice(clv, func(i, j int) bool { return clv[i].uid < clv[j].uid })

	clm := make(map[int]*compactType, len(m.vars))
	for k, v := range m.m {
		clm[k] = v
	}
	return &sortedVarCompactMap{m: clm, vars: clv}
}

////////////////////////////////////////////////////////////////////////////////

// sortedVarCompactMap SortedMap[ Variable, compactType](Ordering by (_.uid))
type sortedVarCompactMap struct {
	m    map[int]*compactType
	vars []*Variable
}

func (s *sortedVarCompactMap) Get(v *Variable) *compactType { return s.m[v.uid] }
func (s *sortedVarCompactMap) Keys() []*Variable {
	cl := make([]*Variable, len(s.vars))
	copy(cl, s.vars)
	return cl
}

////////////////////////////////////////////////////////////////////////////////

type nameCompactMap map[string]*compactType // Map[string, compactType]

func (n nameCompactMap) ToSorted() *sortedNameCompactMap {
	m := make(map[string]*compactType, len(n))
	names := make([]string, 0, len(n))
	for k, v := range n {
		m[k] = v
		names = append(names, k)
	}
	sort.Strings(names)
	return &sortedNameCompactMap{m: m, names: names}
}

////////////////////////////////////////////////////////////////////////////////

// sortedNameCompactMap SortedMap[string, compactType]
type sortedNameCompactMap struct {
	m     nameCompactMap
	names []string
}

func (s *sortedNameCompactMap) Get(name string) *compactType { return s.m[name] }
func (s *sortedNameCompactMap) Len() int                     { return len(s.names) }
func (s *sortedNameCompactMap) Keys() []string {
	cl := make([]string, len(s.names))
	copy(cl, s.names)
	return cl
}

////////////////////////////////////////////////////////////////////////////////

type compactTypeThunk func() *compactType

func (c compactTypeThunk) Equals(that compactTypeThunk) bool {
	return reflect.ValueOf(c).Pointer() == reflect.ValueOf(that).Pointer()
}

////////////////////////////////////////////////////////////////////////////////

type nameCompactThunkMap map[string]compactTypeThunk // Map[string, compactTypeThunk]

func (n nameCompactThunkMap) ToSorted() *sortedNameCompactThunkMap {
	m := make(map[string]compactTypeThunk, len(n))
	names := make([]string, 0, len(n))
	for k, v := range n {
		m[k] = v
		names = append(names, k)
	}
	sort.Strings(names)
	return &sortedNameCompactThunkMap{m: m, names: names}
}

////////////////////////////////////////////////////////////////////////////////

// sortedNameCompactThunkMap SortedMap[string, compactTypeThunk]
type sortedNameCompactThunkMap struct {
	m     nameCompactThunkMap
	names []string
}

func (s *sortedNameCompactThunkMap) Get(name string) compactTypeThunk { return s.m[name] }
func (s *sortedNameCompactThunkMap) Keys() []string {
	cl := make([]string, len(s.names))
	copy(cl, s.names)
	return cl
}

////////////////////////////////////////////////////////////////////////////////

// varCompactThunkMap Map[Variable, compactTypeThunk]
type varCompactThunkMap struct {
	m    map[int]compactTypeThunk
	vars []*Variable
}

func newVarCompactThunkMap() *varCompactThunkMap {
	return &varCompactThunkMap{map[int]compactTypeThunk{}, []*Variable{}}
}
func (m *varCompactThunkMap) Del(v *Variable)                  { delete(m.m, v.uid) }
func (m *varCompactThunkMap) Get(v *Variable) compactTypeThunk { return m.m[v.uid] }
func (m *varCompactThunkMap) Put(v *Variable, c compactTypeThunk) {
	_, ok := m.m[v.uid]
	m.m[v.uid] = c
	if !ok {
		m.vars = append(m.vars, v)
	}
}
func (m *varCompactThunkMap) ToSorted() *sortedVarCompactThunkMap {
	if len(m.m) != len(m.vars) {
		panic("assert")
	}
	clv := make([]*Variable, len(m.vars))
	copy(clv, m.vars)
	sort.Slice(clv, func(i, j int) bool { return clv[i].uid < clv[j].uid })

	clm := make(map[int]compactTypeThunk, len(m.vars))
	for k, v := range m.m {
		clm[k] = v
	}
	return &sortedVarCompactThunkMap{m: clm, vars: clv}
}
func (m *varCompactThunkMap) String() string { return fmt.Sprintf("%s", m.vars) }

////////////////////////////////////////////////////////////////////////////////

// sortedVarCompactThunkMap SortedMap[Variable, compactTypeThunk](Ordering by (_.uid))
type sortedVarCompactThunkMap struct {
	m    map[int]compactTypeThunk
	vars []*Variable
}

func (s *sortedVarCompactThunkMap) Get(v *Variable) compactTypeThunk { return s.m[v.uid] }
func (s *sortedVarCompactThunkMap) Keys() []*Variable {
	cl := make([]*Variable, len(s.vars))
	copy(cl, s.vars)
	return cl
}

////////////////////////////////////////////////////////////////////////////////

// coOccurrences Map[(Boolean, Variable), Set[SimpleType]]
type coOccurrences map[string]*linkedSimpleTypeSet

func (o coOccurrences) Get(p *polarVar) *linkedSimpleTypeSet    { return o[p._hash] }
func (o coOccurrences) Put(p *polarVar, s *linkedSimpleTypeSet) { o[p._hash] = s }

type linkedSimpleTypeSet struct {
	set  map[string]SimpleType
	link []SimpleType
}

func newLinkedSimpleTypeSet() *linkedSimpleTypeSet {
	return &linkedSimpleTypeSet{
		set:  map[string]SimpleType{},
		link: []SimpleType{},
	}
}
func (l *linkedSimpleTypeSet) hash(st SimpleType) string {
	switch x := st.(type) {
	case *Variable:
		return x.hash()
	case *Primitive:
		return x.hash()
	default:
		panic("unreached")
	}
}
func (l *linkedSimpleTypeSet) Add(st SimpleType) {
	hash := l.hash(st)
	_, ok := l.set[hash]
	if ok {
		return
	}
	l.set[hash] = st
	l.link = append(l.link, st)
}
func (l *linkedSimpleTypeSet) Contains(st SimpleType) bool { _, ok := l.set[l.hash(st)]; return ok }
func (l *linkedSimpleTypeSet) Values() []SimpleType {
	cl := make([]SimpleType, len(l.link))
	copy(cl, l.link)
	return cl
}
func (l *linkedSimpleTypeSet) String() string { return fmt.Sprintf("%s", l.link) }

////////////////////////////////////////////////////////////////////////////////

// varSet Set[ Variable]
type varSet map[int]void

func (p varSet) Add(v *Variable)           { p[v.uid] = null }
func (p varSet) Del(v *Variable)           { delete(p, v.uid) }
func (p varSet) Contains(v *Variable) bool { return p[v.uid] == null }

////////////////////////////////////////////////////////////////////////////////

// polarVarSet Set[ polarVar]
type polarVarSet map[string]void

func (p polarVarSet) Add(pv *polarVar)           { p[pv._hash] = null }
func (p polarVarSet) Del(pv *polarVar)           { delete(p, pv._hash) }
func (p polarVarSet) Contains(pv *polarVar) bool { return p[pv._hash] == null }

////////////////////////////////////////////////////////////////////////////////

// polarCompactSet Set[ polarCompact]
type polarCompactSet map[string]void

func (p polarCompactSet) Add(pc *polarCompact)           { p[pc._hash] = null }
func (p polarCompactSet) Del(pc *polarCompact)           { delete(p, pc._hash) }
func (p polarCompactSet) Contains(pc *polarCompact) bool { return p[pc._hash] == null }

////////////////////////////////////////////////////////////////////////////////

// polarVarMap Map[ polarVar, Variable]
type polarVarMap map[string]*Variable

func (p polarVarMap) Get(pv *polarVar) *Variable      { return p[pv._hash] }
func (p polarVarMap) Put(pv *polarVar, val *Variable) { p[pv._hash] = val }

////////////////////////////////////////////////////////////////////////////////

type typeVarThunk func() *types.TypeVariable

// polarCompactMap Map[(compactTypeOrVariable, Boolean), () => TypeVariable | *Variable]
type polarCompactMap map[string]interface{}

func (p polarCompactMap) Get(pc *polarCompact) interface{}    { return p[pc._hash] }
func (p polarCompactMap) Put(pc *polarCompact, i interface{}) { p[pc._hash] = i }
func (p polarCompactMap) Del(pc *polarCompact)                { delete(p, pc._hash) }

////////////////////////////////////////////////////////////////////////////////

// typeVarMap Map[PolarVariable, TypeVariable]
type typeVarMap map[string]*types.TypeVariable

func (t typeVarMap) Get(p *polarVar) *types.TypeVariable     { return t[p._hash] }
func (t typeVarMap) Put(p *polarVar, tv *types.TypeVariable) { t[p._hash] = tv }
