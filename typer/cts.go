package typer

import (
	"fmt"
	"github.com/goghcrow/simple-sub/util"
	"sort"
	"strconv"
)

// Compact types representation, useful for simplification

type compactTypeScheme struct {
	term    *compactType
	recVars *sortedVarCompactMap
}

type compactFun struct {
	lhs, rhs *compactType
}

type compactTypeOrVariable interface {
	polHash(polarity bool) string
}

// compactType 出现在正极代表 union, 出现在负极代表 intersection
type compactType struct {
	vs   *sortedVarSet         // variable
	prim *sortedPrimSet        // primitive
	tup  []*compactType        // tuple
	rec  *sortedNameCompactMap // record
	fun  *compactFun           // function

	_hash string
}

func emptyCompactType() *compactType {
	return &compactType{
		vs:   emptyVarSet(),
		prim: emptyPrimSet(),
	}
}

func (c *compactType) isEmpty() bool {
	return c.vs.Len() == 0 && c.prim.Len() == 0 && c.rec == nil && c.fun == nil
}

func (c *compactType) String() string {
	xs := make([]string, 0, c.vs.Len()+c.prim.Len()+2)
	for _, tv := range c.vs.Values() {
		xs = append(xs, tv.String())
	}
	for _, prim := range c.prim.Values() {
		xs = append(xs, prim.Name)
	}
	if c.rec != nil {
		if c.rec.Len() == 0 {
			xs = append(xs, "{}")
		} else {
			xss := make([]string, c.rec.Len())
			for i, name := range c.rec.Keys() {
				xss[i] = fmt.Sprintf("%s: %s", name, c.rec.Get(name))
			}
			xs = append(xs, util.JoinStr(xss, ", ", "{", "}"))
		}
	}
	if c.fun != nil {
		xs = append(xs, fmt.Sprintf("%s -> %s", c.fun.lhs, c.fun.rhs))
	}
	return util.JoinStr(xs, ", ", "‹", "›")
}

// 注意这里假设 compactType 构造完只读
func (c *compactType) hash() string {
	if c._hash != "" {
		return c._hash
	}

	xs := make([]string, 0, c.vs.Len()+c.prim.Len()+2)

	{
		ids := make([]int, c.vs.Len())
		for i, tv := range c.vs.Values() {
			ids[i] = tv.uid
		}
		sort.Ints(ids)
		for _, uid := range ids {
			xs = append(xs, strconv.Itoa(uid))
		}
	}

	{
		names := make([]string, 0, c.prim.Len())
		for _, prim := range c.prim.Values() {
			names = append(names, prim.Name)
		}
		sort.Strings(names)
		for _, name := range names {
			xs = append(xs, name)
		}
	}

	if c.rec != nil && c.rec.Len() != 0 {
		xss := make([]string, c.rec.Len())
		for i, name := range c.rec.Keys() {
			xss[i] = fmt.Sprintf("%s: %s", name, c.rec.Get(name).hash())
		}
		xs = append(xs, util.JoinStr(xss, ", ", "{", "}"))
	}

	if c.fun != nil {
		xs = append(xs, fmt.Sprintf("%s -> %s", c.fun.lhs.hash(), c.fun.rhs.hash()))
	}

	c._hash = util.JoinStr(xs, ", ", "‹", "›")
	return c._hash
}
