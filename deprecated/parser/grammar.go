package parser

import (
	"github.com/goghcrow/simple-sub/deprecated/oper"
	"github.com/goghcrow/simple-sub/deprecated/token"
	"github.com/goghcrow/simple-sub/terms"
	"github.com/goghcrow/simple-sub/util"
)

type nud func(*parser, oper.BP, *token.Token) terms.Term
type led func(*parser, oper.BP, terms.Term, *token.Token) terms.Term

// 处理字面量、变量、前缀操作符
type prefix struct {
	oper.BP
	nud
}

// 处理中缀、后缀操作符
type infix struct {
	oper.BP
	led
}

// 如果不支持自定操作符, 则 tokenType 可以定义成 int enum
// prefixs & infixs 则可以定义成 tokenType 为下标的数组

type grammar struct {
	prefixs map[token.Type]prefix
	infixs  map[token.Type]infix
}

// 前缀操作符
func (g *grammar) prefix(t token.Type, bp oper.BP, f nud) {
	g.prefixs[t] = prefix{bp, f}
}

// 不结合中缀操作符
func (g *grammar) infix(t token.Type, bp oper.BP, f led) {
	g.infixs[t] = infix{bp, f}
}

// 右结合中缀操作符
func (g *grammar) infixRight(t token.Type, bp oper.BP, f led) {
	g.infix(t, bp, f)
}

// 左结合中缀操作符
func (g *grammar) infixLeft(t token.Type, bp oper.BP, f led) {
	g.infix(t, bp, f)
}

// 后缀操作符（可以看成中缀操作符木有右边操作数）
func (g *grammar) postfix(t token.Type, bp oper.BP, f led) {
	g.infix(t, bp, f)
}

// left binding powers
func (p *grammar) infixLbp(t *token.Token) oper.BP {
	i, ok := p.infixs[t.Type]
	if ok {
		return i.BP
	} else {
		return 0
	}
}

func (g *grammar) mustPrefix(t *token.Token) prefix {
	p, ok := g.prefixs[t.Type]
	util.Assert(ok, "syntax error: %s", t)
	return p
}

func (g *grammar) mustInfix(t *token.Token) infix {
	i, ok := g.infixs[t.Type]
	util.Assert(ok, "syntax error: %s", t)
	return i
}
