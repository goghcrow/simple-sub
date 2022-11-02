package parser

import (
	"github.com/goghcrow/simple-sub/deprecated/lexer"
	"github.com/goghcrow/simple-sub/deprecated/oper"
	"github.com/goghcrow/simple-sub/deprecated/token"
	"github.com/goghcrow/simple-sub/terms"
)

// token.Type 不重复, 顺序无关, 重复默认覆盖
func newGrammar(ops []oper.Operator) grammar {
	g := grammar{
		prefixs: map[token.Type]prefix{},
		infixs:  map[token.Type]infix{},
	}

	g.prefix(token.NAME, oper.BP_NONE, parseIdent)

	g.prefix(token.TRUE, oper.BP_NONE, parseBool)
	g.prefix(token.FALSE, oper.BP_NONE, parseBool)
	g.prefix(token.INT, oper.BP_NONE, parseInt)
	g.prefix(token.FLOAT, oper.BP_NONE, parseFloat)
	g.prefix(token.STR, oper.BP_NONE, parseString)

	//g.prefix(token.LEFT_BRACKET, oper.BP_NONE, parseListMap)
	g.prefix(token.LEFT_PAREN, oper.BP_NONE, parseParen)
	g.prefix(token.LEFT_BRACE, oper.BP_NONE, parseRecord)

	g.prefix(token.IF, oper.BP_NONE, parseIf)
	g.prefix(token.LET, oper.BP_NONE, parseLet)
	g.prefix(token.FUN, oper.BP_NONE, parseFun)

	// todo 这里处理成动态结构, 支持边 parse 边动态注册
	for _, op := range ops {
		switch op.Fixity {
		case oper.PREFIX:
			g.prefix(op.Type, op.BP, unaryPrefix)
		case oper.INFIX_N:
			g.infix(op.Type, op.BP, binaryN)
		case oper.INFIX_L:
			g.infix(op.Type, op.BP, binaryL)
		case oper.INFIX_R:
			g.infix(op.Type, op.BP, binaryR)
		case oper.POSTFIX:
			g.postfix(op.Type, op.BP, unaryPostfix)
		}
	}

	// 放在自定义操作符后面, 防止被覆盖
	g.infixRight(token.ARROW, oper.BP_MEMBER, parseArrow)
	g.infixLeft(token.DOT, oper.BP_MEMBER, parseDot)
	g.infixLeft(token.LEFT_PAREN, oper.BP_CALL, parseCall)
	//g.infixLeft(token.LEFT_BRACKET, oper.BP_MEMBER, parseSubscript)
	return g
}

func binaryL(p *parser, bp oper.BP, lhs terms.Term, t *token.Token) terms.Term {
	rhs := p.expr(bp)
	return terms.Bin(t.Lexeme, oper.INFIX_L, lhs, rhs)
}

func binaryR(p *parser, bp oper.BP, lhs terms.Term, t *token.Token) terms.Term {
	rhs := p.expr(bp - 1)
	return terms.Bin(t.Lexeme, oper.INFIX_R, lhs, rhs)
}

func binaryN(p *parser, bp oper.BP, lhs terms.Term, t *token.Token) terms.Term {
	rhs := p.expr(bp) // 这里是否-1无所谓, 之后会检查
	return terms.Bin(t.Lexeme, oper.INFIX_N, lhs, rhs)
}

func unaryPrefix(p *parser, bp oper.BP, t *token.Token) terms.Term {
	term := p.expr(bp)
	return terms.Un(t.Lexeme, term, true)
}

func unaryPostfix(p *parser, bp oper.BP, lhs terms.Term, t *token.Token) terms.Term {
	return terms.Un(t.Lexeme, lhs, false)
}

func parseIdent(p *parser, bp oper.BP, t *token.Token) terms.Term {
	p.syntaxAssert(!lexer.Reserved(t.Lexeme), "%s reserved", t.Lexeme)
	return terms.Var(t.Lexeme)
}

func parseRecord(p *parser, bp oper.BP, t *token.Token) terms.Term {
	var xs []terms.Field
	for {
		p.tryEatLines()
		if p.peek().Type == token.RIGHT_BRACE {
			break
		}
		n := mustEatName(p)
		p.tryEatLines()
		p.mustEat(token.COLON)
		p.tryEatLines()
		v := p.expr(0)
		p.tryEatLines()
		xs = append(xs, terms.Field{Name: n.Lexeme, Term: v})
		if p.tryEatAny(token.COMMA, token.NEWLINE) == nil {
			break
		}
	}
	p.tryEatLines()
	p.mustEat(token.RIGHT_BRACE)
	return terms.Rcd(xs)
}

func parseParen(p *parser, bp oper.BP, t *token.Token) terms.Term {
	return p.any(parseTuple, parseGroup)
}

func parseTuple(p *parser) terms.Term {
	// () (1, ) (1,2)
	p.tryEatLines()
	var xs []terms.Term
	for p.tryEat(token.RIGHT_PAREN) == nil {
		p.tryEatLines()
		xs = append(xs, p.expr(0))
		p.tryEatLines()
		if p.tryEat(token.RIGHT_PAREN) != nil {
			// group 和 (1) 冲突, 需要 (1, )
			p.syntaxAssert(len(xs) != 1, "expect (x, )")
			break
		}
		p.mustEat(token.COMMA)
		p.tryEatLines()
	}
	return terms.Tup(xs...)
}

func parseGroup(p *parser) terms.Term {
	p.tryEatLines()
	term := p.expr(0)
	p.tryEatLines()
	p.mustEat(token.RIGHT_PAREN)
	return terms.Grp(term)
}

func parseDot(p *parser, bp oper.BP, obj terms.Term, t *token.Token) terms.Term {
	p.tryEatLines()
	name := mustEatName(p)
	//name := p.eat()
	// 放开限制则可以写 1. +(1), 1可以看成对象, .和+必须有空格是因为否则会匹配自定义操作符
	//p.syntaxAssert(name.Type == token.NAME || name.Type == token.TRUE || name.Type == token.FALSE,
	//	"syntax error: %s", name.Lexeme)
	sel := terms.Sel(obj, name.Lexeme)
	marked := p.mark()
	p.tryEatLines()
	lp := p.tryEat(token.LEFT_PAREN)
	if lp == nil {
		p.reset(marked)
		return sel
	} else {
		return parseCall(p, bp, sel, lp)
	}
}

func parseCall(p *parser, bp oper.BP, callee terms.Term, t *token.Token) terms.Term {
	p.tryEatLines()
	var xs []terms.Term
	for {
		// 至少一个参数
		xs = append(xs, p.expr(0))
		p.tryEatLines()
		if p.tryEat(token.COMMA) == nil {
			break
		}
		p.tryEatLines()
	}
	p.mustEat(token.RIGHT_PAREN)
	return terms.AppN(callee, xs...)
}

func parseIf(p *parser, bp oper.BP, iff *token.Token) terms.Term {
	p.tryEatLines()
	cond := p.expr(0)
	p.tryEatLines()
	p.mustEat(token.THEN)
	p.tryEatLines()
	then := p.expr(0)
	p.tryEatLines()
	p.mustEat(token.ELSE)
	p.tryEatLines()
	els := p.expr(0)
	return terms.Iff(cond, then, els)
}

func (p *parser) parseTopLevel(bp oper.BP) *terms.Define {
	return doParseLet(true, p, bp, p.mustEat(token.LET)).(*terms.Define)
}

func parseLet(p *parser, bp oper.BP, let *token.Token) terms.Term {
	return doParseLet(false, p, bp, let)
}

// let rec id = e
// let rec x = e1 in e2
// let rec f a b c ... = e1 in e2
func doParseLet(topLevel bool, p *parser, bp oper.BP, let *token.Token) terms.Term {
	p.tryEatLines()
	rec := p.tryEat(token.REC)
	if rec != nil {
		p.tryEatLines()
	}

	// 至少 1 个
	var xs []string
	for {
		xs = append(xs, mustEatName(p).Lexeme)
		p.tryEatLines()
		if p.tryEat(token.ASSIGN) != nil {
			break
		}
	}
	p.tryEatLines()
	rhs := p.expr(0)

	marked := p.mark()
	p.tryEatLines()
	if p.tryEat(token.IN) != nil {
		p.tryEatLines()
		body := p.expr(0)
		if len(xs) == 1 {
			// let rec x = e1 in e2
			return terms.Let(xs[0], rhs, body, rec != nil)
		} else {
			// let f a b c ... = e1 in e2
			return terms.Let(xs[0], terms.LamN(xs[1:], rhs), body, rec != nil)
		}
	} else {
		p.reset(marked)
		p.syntaxAssert(topLevel, "expect in")
		if len(xs) == 1 {
			// let rec id = e
			return terms.Def(xs[0], rhs, rec != nil)
		} else {
			// let rec f x1 x2 ... xn = e
			return terms.Def(xs[0], terms.LamN(xs[1:], rhs), rec != nil)
		}
	}
}

// 不支持零参函数 fun x1 ... xn -> e
func parseFun(p *parser, bp oper.BP, fun *token.Token) terms.Term {
	p.tryEatLines()
	var xs []string
	for {
		xs = append(xs, mustEatName(p).Lexeme)
		p.tryEatLines()
		if p.tryEat(token.ARROW) != nil {
			break
		}
	}
	p.tryEatLines()
	rhs := p.expr(0)
	return terms.LamN(xs, rhs)
}

func parseArrow(p *parser, bp oper.BP, l terms.Term, a *token.Token) terms.Term {
	id, ok := l.(*terms.Variable)
	p.syntaxAssert(ok, "expect id -> ")
	p.tryEatLines()
	return terms.Lam(id.Name, p.expr(0))
}

func mustEatName(p *parser) *token.Token {
	name := p.mustEat(token.NAME)
	p.syntaxAssert(!lexer.Reserved(name.Lexeme), "%s reserved", name.Lexeme)
	return name
}

//func parseListMap(p *parser, bp oper.BP, t *token.Token) terms.Term {
//	if p.tryEat(token.COLON) != nil {
//		p.mustEat(token.RIGHT_BRACKET)
//		return terms.Map([]terms.Pair{})
//	}
//	return p.any(parseList, parseMap)
//}
//func parseList(p *parser) terms.Term {
//	elems := make([]terms.Term, 0)
//	for {
//		if p.peek().Type == token.RIGHT_BRACKET {
//			break
//		}
//		el := p.expr(0)
//		elems = append(elems, el)
//		if p.tryEat(token.COMMA) == nil {
//			break
//		}
//	}
//	p.mustEat(token.RIGHT_BRACKET)
//	return terms.List(elems)
//}
//func parseMap(p *parser) terms.Term {
//	pairs := make([]terms.Pair, 0)
//	for {
//		if p.peek().Type == token.RIGHT_BRACKET {
//			break
//		}
//		k := p.expr(0)
//		p.mustEat(token.COLON)
//		v := p.expr(0)
//		pairs = append(pairs, terms.Pair{Key: k, Val: v})
//		if p.tryEat(token.COMMA) == nil {
//			break
//		}
//	}
//	p.mustEat(token.RIGHT_BRACKET)
//	return terms.Map(pairs)
//}
//func parseSubscript(p *parser, bp oper.BP, list terms.Term, t *token.Token) terms.Term {
//	term := p.expr(0)
//	p.mustEat(token.RIGHT_BRACKET)
//	return terms.Subscript(list, term)
//}
