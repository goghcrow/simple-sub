package types

type void *struct{}

var null void = &struct{}{}

type Ctx map[int]string

func (c Ctx) Get(t *TypeVariable) string       { return c[t.hash] }
func (c Ctx) Put(t *TypeVariable, name string) { c[t.hash] = name }

////////////////////////////////////////////////////////////////////////////////

type MutTypeVarSet map[int]void

func (s MutTypeVarSet) Add(t *TypeVariable)           { s[t.hash] = null }
func (s MutTypeVarSet) Contains(t *TypeVariable) bool { _, ok := s[t.hash]; return ok }
