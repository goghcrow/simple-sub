package parser

import (
	. "github.com/goghcrow/go-parsec"
	"github.com/goghcrow/go-parsec/lexer"
	"github.com/goghcrow/simple-sub/terms"
)

//goland:noinspection GoSnakeCaseUsage,SpellCheckingInspection
const (
	LET lexer.TokenKind = iota + 1
	REC
	IN
	FUN
	IF
	THEN
	ELSE
	TRUE
	FALSE
	NOT

	IDENT

	FLOAT
	INT
	STR

	COLON
	COMMA
	ASSIGN

	DOT
	ARROW

	//PLUS
	//PLUSF
	//SUB
	//SUBF
	//MUL
	//MULF
	//DIV
	//DIVF

	LE
	LT
	GE
	GT
	EQ
	NE

	LOGIC_OR
	LOGIC_AND
	LOGIC_NOT

	LEFT_PAREN
	RIGHT_PAREN
	LEFT_BRACKET
	RIGHT_BRACKET
	LEFT_BRACE
	RIGHT_BRACE

	NEWLINE
	WHITESPACE
	BLOCK_COMMENT
	LINE_COMMENT
)

var lex = lexer.BuildLexer(func(l *lexer.Lexicon) {
	l.Regex(WHITESPACE, "[ \r\t]+").Skip() // 不能使用\s+, 要单独处理换行
	l.Regex(BLOCK_COMMENT, "/\\*[\\s\\S]*?\\*+/").Skip()
	l.Regex(LINE_COMMENT, "//.*").Skip()

	l.Str(NEWLINE, "\n").Skip()
	l.Str(COLON, ":")
	l.Str(COMMA, ",")

	l.Str(ASSIGN, "=")
	l.Str(LEFT_PAREN, "(")
	l.Str(RIGHT_PAREN, ")")
	l.Str(LEFT_BRACKET, "[")
	l.Str(RIGHT_BRACKET, "]")
	l.Str(LEFT_BRACE, "{")
	l.Str(RIGHT_BRACE, "}")

	l.Keyword(LET, "let")
	l.Keyword(REC, "rec")
	l.Keyword(IN, "in")
	l.Keyword(FUN, "fun")
	l.Keyword(IF, "if")
	l.Keyword(THEN, "then")
	l.Keyword(ELSE, "else")
	l.Keyword(TRUE, "true")
	l.Keyword(FALSE, "false")
	l.Keyword(NOT, "not")

	l.Oper(DOT, ".")
	l.Oper(ARROW, "->")

	//l.Oper(PLUSF, "+.")
	//l.Oper(PLUS, "+")
	//l.Oper(SUBF, "-.")
	//l.Oper(SUB, "-")
	//l.Oper(MULF, "*.")
	//l.Oper(MUL, "*")
	//l.Oper(DIVF, "/.")
	//l.Oper(DIV, "/")

	l.Oper(LE, "<=")
	l.Oper(LT, "<")
	l.Oper(GE, ">=")
	l.Oper(GT, ">")
	l.Oper(EQ, "==")
	l.Oper(NE, "!=")

	l.Oper(LOGIC_OR, "||")
	l.Oper(LOGIC_AND, "&&")
	l.Oper(LOGIC_NOT, "!")

	l.Regex(FLOAT, "[-+]?(?:0|[1-9][0-9]*)(?:[.][0-9]+)+(?:[eE][-+]?[0-9]+)?")
	l.Regex(FLOAT, "[-+]?(?:0|[1-9][0-9]*)(?:[.][0-9]+)?(?:[eE][-+]?[0-9]+)+")
	l.Regex(INT, "[-+]?0b(?:0|1[0-1]*)")
	l.Regex(INT, "[-+]?0x(?:0|[1-9a-fA-F][0-9a-fA-F]*)")
	l.Regex(INT, "[-+]?0o(?:0|[1-7][0-7]*)")
	l.Regex(INT, "[-+]?(?:0|[1-9][0-9]*)")

	l.Regex(STR, "\"(?:[^\"\\\\]*|\\\\[\"\\\\trnbf\\/]|\\\\u[0-9a-fA-F]{4})*\"")
	l.Regex(STR, "`[^`]*`") // raw string

	l.Regex(IDENT, "[a-zA-Z\\p{L}_][a-zA-Z0-9\\p{L}_]*") // 支持 unicode, 不能以数字开头
})

var (
	_Expr = NewRule() // for test
	_Pgrm = NewRule()
)

//goland:noinspection GoSnakeCaseUsage
func init() {
	var (
		Term         = NewRule()
		Const        = NewRule()
		Ident        = NewRule()
		Variable     = NewRule()
		Parens       = NewRule()
		SubTermNoSel = NewRule()
		SubTerm      = NewRule()
		Record       = NewRule()
		Tuple        = NewRule()
		List         = NewRule()
		Fun          = NewRule()
		Let          = NewRule()
		Ite          = NewRule()
		Apps         = NewRule()

		TopLevel = NewRule()
	)

	applyTrue := func(v interface{}) interface{} { return terms.Bool(true) }
	applyFalse := func(v interface{}) interface{} { return terms.Bool(false) }
	applyInt := func(v interface{}) interface{} { return parseInt(v.(*lexer.Token)) }
	applyFloat := func(v interface{}) interface{} { return parseFloat(v.(*lexer.Token)) }
	applyStr := func(v interface{}) interface{} { return parseString(v.(*lexer.Token)) }
	applyIdent := func(v interface{}) interface{} { return v.(*lexer.Token).Lexeme }
	applyVar := func(v interface{}) interface{} { return terms.Var(v.(*lexer.Token).Lexeme) }
	applySubTerm := func(v interface{}) interface{} {
		t2 := v.([]interface{})
		rcd := t2[0].(terms.Term)
		ids := t2[1].([]interface{})
		if len(ids) == 0 {
			return rcd
		}
		sub := terms.Sel(rcd, ids[0].(string))
		for i := 1; i < len(ids); i++ {
			sub = terms.Sel(sub, ids[i].(string))
		}
		return sub
	}
	applyRecord := func(v interface{}) interface{} {
		if v == nil {
			return terms.Rcd([]terms.Field{})
		}
		pairs := v.([]interface{})
		xs := make([]terms.Field, len(pairs))
		for i, it := range pairs {
			t3 := it.([]interface{})
			xs[i] = terms.Field{Name: t3[0].(string), Term: t3[2].(terms.Term)}
		}
		return terms.Rcd(xs)
	}
	applyTuple := func(v interface{}) interface{} {
		if v == nil {
			return terms.Tup() // 0
		}
		a := v.([]interface{})
		fst := a[0].(terms.Term)
		if a[2] == nil {
			return terms.Tup(fst) // 1
		}
		rest := a[2].([]interface{})
		xs := make([]terms.Term, 1+len(rest))
		xs[0] = fst
		for i, it := range rest {
			xs[i+1] = it.(terms.Term)
		}
		return terms.Tup(xs...) // n
	}
	applyList := func(v interface{}) interface{} {
		if v == nil {
			return terms.Lst()
		}
		els := v.([]interface{})
		xs := make([]terms.Term, len(els))
		for i, el := range els {
			xs[i] = el.(terms.Term)
		}
		return terms.Lst(xs...)
	}
	applyFun := func(v interface{}) interface{} {
		t4 := v.([]interface{})
		return terms.LamN(t4[1].([]string), t4[3].(terms.Term))
	}
	applyLet := func(v interface{}) interface{} {
		t6 := v.([]interface{})
		name := t6[1].(string)
		rhs := t6[3].(terms.Term)
		body := t6[5].(terms.Term)
		return terms.Let(name, rhs, body, t6[0] != nil)
	}
	applyIte := func(v interface{}) interface{} {
		t6 := v.([]interface{})
		cond := t6[1].(terms.Term)
		then := t6[3].(terms.Term)
		els := t6[5].(terms.Term)
		return terms.Iff(cond, then, els)
	}
	applyApps := func(v interface{}) interface{} {
		t2 := v.([]interface{})
		lhs := t2[0].(terms.Term)
		xs := t2[1].([]interface{})
		if len(xs) == 0 {
			return lhs
		}
		rhs := xs[0].(terms.Term)
		app := terms.App(lhs, rhs)
		for i := 1; i < len(xs); i++ {
			app = terms.App(app, xs[i].(terms.Term))
		}
		return app
	}
	applyTopLevel := func(v interface{}) interface{} {
		t4 := v.([]interface{})
		name := t4[1].(string)
		rhs := t4[3].(terms.Term)
		return terms.Decl(name, rhs, t4[0] != nil)
	}
	applyPgrm := func(v interface{}) interface{} {
		xs := v.([]interface{})
		defs := make([]*terms.Declaration, len(xs))
		for i, x := range xs {
			defs[i] = x.(*terms.Declaration)
		}
		return terms.Pgrm(defs)
	}
	applyAtLeastIdent := func(v interface{}) interface{} {
		t2 := v.([]interface{})
		fst := t2[0].(string)
		rest := t2[1].([]interface{})
		xs := make([]string, 1+len(rest))
		xs[0] = fst
		for i, x := range rest {
			xs[i+1] = x.(string)
		}
		return xs
	}

	//Signed       = NewRule()
	//Factor       = NewRule()
	//Addition     = NewRule()
	//AddSub := Alt(Tok(PLUS), Tok(SUB), Tok(PLUSF), Tok(SUBF))
	//MulDiv := Alt(Tok(MUL), Tok(MULF), Tok(DIV), Tok(DIVF))
	//Signed.Pattern = Seq(OptSc(Alt(Tok(PLUS), Tok(SUB))), Term).Map(applyUnary)
	//Factor.Pattern = LRecSc(Signed, Seq(MulDiv, Signed), applyBinary)
	//Addition.Pattern = LRecSc(Factor, Seq(AddSub, Factor), applyBinary)
	//applyUnary := func(v interface{}) interface{} {
	//	t2 := v.([]interface{})
	//	oper := t2[0].(*lexer.Token)
	//	rhs := t2[1].(terms.Term)
	//	if oper == nil {
	//		return rhs
	//	}
	//	return terms.App(terms.Var(oper.Lexeme), rhs)
	//}
	//applyBinary := func(a, b interface{}) interface{} {
	//	lhs := a.(interface{}).(terms.Term)
	//	oper := b.([]interface{})[0].(*lexer.Token)
	//	rhs := b.([]interface{})[1].(terms.Term)
	//	return terms.AppN(terms.Var(oper.Lexeme), lhs, rhs)
	//}

	Term.Pattern = Alt(Let, Fun, Ite, Apps)
	Const.Pattern = Alt(
		Tok(INT).Map(applyInt),
		Tok(FLOAT).Map(applyFloat),
		Tok(STR).Map(applyStr),
		Tok(TRUE).Map(applyTrue),
		Tok(FALSE).Map(applyFalse),
	)
	Ident.Pattern = Alt(Tok(IDENT), Tok(TRUE), Tok(FALSE)).Map(applyIdent)
	Variable.Pattern = Tok(IDENT).Map(applyVar)
	Parens.Pattern = KMid(Tok(LEFT_PAREN), Term, Tok(RIGHT_PAREN))
	SubTermNoSel.Pattern = Alt(Parens, Record, Tuple, List, Const, Variable)
	SubTerm.Pattern = Seq(SubTermNoSel, RepSc(KRight(Tok(DOT), Ident))).Map(applySubTerm)
	Record.Pattern = KMid(
		Tok(LEFT_BRACE),
		OptSc(ListSc(Seq(Ident, Tok(COLON), Term), Tok(COMMA))),
		Tok(RIGHT_BRACE),
	).Map(applyRecord)
	Tuple.Pattern = KMid(
		Tok(LEFT_PAREN),
		Alt(Nil(), Seq(Term, Tok(COMMA), OptSc(ListSc(Term, Tok(COMMA))))),
		Tok(RIGHT_PAREN),
	).Map(applyTuple)
	List.Pattern = KMid(
		Tok(LEFT_BRACKET),
		OptSc(ListSc(Term, Tok(COMMA))),
		Tok(RIGHT_BRACKET),
	).Map(applyList)
	atLeastIdent := Seq(Ident, RepSc(Ident)).Map(applyAtLeastIdent)
	Fun.Pattern = Seq(Tok(FUN), atLeastIdent, Tok(ARROW), Term).Map(applyFun)
	Let.Pattern = Seq(KRight(Tok(LET), OptSc(Tok(REC))), Ident, Tok(ASSIGN), Term, Tok(IN), Term).Map(applyLet)
	Ite.Pattern = Seq(Tok(IF), Term, Tok(THEN), Term, Tok(ELSE), Term).Map(applyIte)
	// 变量和函数都使用 let 声明, 变量是零参函数, 函数是有参变量, apply 的语法就统一了
	Apps.Pattern = Seq(SubTerm, RepSc(SubTerm)).Map(applyApps)

	_Expr.Pattern = Term

	TopLevel.Pattern = Seq(KRight(Tok(LET), OptSc(Tok(REC))), Ident, Tok(ASSIGN), Term).Map(applyTopLevel)
	_Pgrm.Pattern = RepSc(TopLevel).Map(applyPgrm)
}

func ParseExpr(s string) (terms.Term, error) { return parse(_Expr, s) }

func ParsePgrm(s string) (*terms.Program, error) {
	pgrm, err := parse(_Pgrm, s)
	if err != nil {
		return nil, err
	}
	return pgrm.(*terms.Program), nil
}

func parse(p Parser, s string) (terms.Term, error) {
	toks, err := lex.Lex(s)
	if err != nil {
		return nil, err
	}
	result, err := ExpectSingleResult(ExpectEOF(p.Parse(toks)))
	if err != nil {
		return nil, err
	}
	return result.(terms.Term), nil
}
