package parser

import (
	"github.com/goghcrow/simple-sub/front/lexer"
	"github.com/goghcrow/simple-sub/front/oper"
	"github.com/goghcrow/simple-sub/front/token"
	"github.com/goghcrow/simple-sub/terms"
	"github.com/goghcrow/simple-sub/util"
)

// parser 使用了 Top Down Operator Precedence
// 可以参考道格拉斯的文章 https://www.crockford.com/javascript/tdop/tdop.html

func NewParser(ops []oper.Operator) *parser {
	return &parser{
		grammar: newGrammar(oper.Sort(ops)),
	}
}

func (p *parser) Parse(toks []*token.Token) *terms.Program {
	p.idx = 0
	p.toks = toks

	var xs []*terms.Define
	p.tryEatLines()
	for p.peek() != lexer.EOF {
		xs = append(xs, p.parseTopLevel(0))
		if p.peek() == lexer.EOF {
			break
		}
		p.mustEat(token.NEWLINE)
		p.tryEatLines()
	}
	p.mustEat(token.EOF)
	return terms.Pgrm(xs)
}

type parser struct {
	grammar
	toks []*token.Token
	idx  int
}

func (p *parser) peek() *token.Token {
	if p.idx >= len(p.toks) {
		return lexer.EOF
	}
	return p.toks[p.idx]
}

func (p *parser) eat() *token.Token {
	if p.idx >= len(p.toks) {
		return lexer.EOF
	}
	t := p.toks[p.idx]
	p.idx++
	return t
}

func (p *parser) mustEat(typ token.Type) *token.Token {
	t := p.eat()
	p.syntaxAssert(t.Type == typ, "expect %s actual %s", typ, t)
	return t
}

func (p *parser) tryEat(typ token.Type) *token.Token {
	if p.peek().Type == typ {
		return p.eat()
	} else {
		return nil
	}
}

func (p *parser) tryEatAny(xs ...token.Type) *token.Token {
	for _, typ := range xs {
		tok := p.tryEat(typ)
		if tok != nil {
			return tok
		}
	}
	return nil
}

func (p *parser) mark() int      { return p.idx }
func (p *parser) reset(mark int) { p.idx = mark }

func (p *parser) tryParse(f func(p *parser) terms.Term) (expr terms.Term) {
	marked := p.mark()
	defer func() {
		if r := recover(); r != nil {
			p.reset(marked)
			expr = nil
		}
	}()
	return f(p)
}

func (p *parser) any(fs ...func(p *parser) terms.Term) (expr terms.Term) {
	for _, f := range fs {
		n := p.tryParse(f)
		if n != nil {
			return n
		}
	}
	util.Assert(false, "try parse fail")
	return nil
}

// parser bp > rbp 的表达式
func (p *parser) expr(rbp oper.BP) terms.Term {
	t := p.eat()
	// tok 必须有 prefix 解析器, 否则一定语法错误
	pre := p.mustPrefix(t)
	left := pre.nud(p, pre.BP, t)
	return p.parseInfix(left, rbp)
}

func (p *parser) parseInfix(left terms.Term, rbp oper.BP) terms.Term {
	// 判断下一个 tok 是否要绑定 left ( 优先级 > left)
	for p.infixLbp(p.peek()) > rbp {
		t := p.eat()
		inf := p.mustInfix(t)
		left = inf.led(p, inf.BP, left, t)
	}
	return p.infixNCheck(left)
}

func (p *parser) infixNCheck(expr terms.Term) terms.Term {
	if bin, ok := expr.(*terms.Binary); ok && bin.Fixity == oper.INFIX_N {
		if lhs, ok := bin.Lhs.(*terms.Binary); ok {
			p.syntaxAssert(lhs.Name != bin.Name, "%s non-infix", bin.Name)
		}
		if rhs, ok := bin.Rhs.(*terms.Binary); ok {
			p.syntaxAssert(rhs.Name != bin.Name, "%s non-infix", bin.Name)
		}
	}
	return expr
}

func (p *parser) tryEatLines() {
	for p.tryEat(token.NEWLINE) != nil {
	}
}

func (p *parser) syntaxAssert(cond bool, format string, a ...interface{}) {
	util.Assert(cond, "syntax error: "+format, a...)
}
