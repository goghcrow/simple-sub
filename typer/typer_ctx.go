package typer

// TODO HAMT 实现
// 如果环境可以修改, fun 和 let 的子作用域可以处理成 parent 形式
// 这里 ctx 只读, 最好使用 immutable map, 可以使用 HAMT 实现, 这里简单处理

type Ctx struct {
	env map[string]TypeScheme
}

func NewCtx(env map[string]TypeScheme) *Ctx          { return &Ctx{env: env} }
func (c *Ctx) Add(nme string, ty TypeScheme)         { c.env[nme] = ty }
func (c *Ctx) Extend(nme string, ty TypeScheme) *Ctx { nc := c.clone(); nc.Add(nme, ty); return nc }
func (c *Ctx) MustLookup(nme string) TypeScheme {
	ts, ok := c.env[nme]
	if !ok {
		panic(NewTypeError("identifier not found: " + nme))
	}
	return ts
}

func (c *Ctx) clone() *Ctx {
	m := make(map[string]TypeScheme, len(c.env))
	for k, v := range c.env {
		m[k] = v
	}
	return NewCtx(m)
}

//import "github.com/goghcrow/simple-sub/util"
//
//// 如果环境可以修改, fun 和 let 的子作用域可以处理成 parent 形式
//// 这里 ctx 只读, 最好使用 immutable map, 可以使用 HAMT 实现, 这里简单处理
//
//type Ctx struct {
//	parent *Ctx
//	env    map[string]TypeScheme
//}
//
//func NewCtx(env map[string]TypeScheme) *Ctx { return &Ctx{env: env} }
//func (c *Ctx) Add(nme string, ty TypeScheme) {
//	util.Assert(c.env[nme] == nil, "redefine %s", nme)
//	c.env[nme] = ty
//}
//func (c *Ctx) Extend(nme string, ty TypeScheme) *Ctx { return &Ctx{c, map[string]TypeScheme{nme: ty}} } // 这里会退化成线性, 应该用 immutable map 实现
//func (c *Ctx) Lookup(nme string) (TypeScheme, bool) {
//	ts, ok := c.env[nme]
//	if ok {
//		return ts, true
//	}
//	if c.parent != nil {
//		return c.parent.Lookup(nme)
//	}
//	return nil, false
//}
//func (c *Ctx) MustLookup(nme string) TypeScheme {
//	ts, ok := c.Lookup(nme)
//	if !ok {
//		panic(NewTypeError("identifier not found: " + nme))
//	}
//	return ts
//}
