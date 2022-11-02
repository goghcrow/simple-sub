package parser

import (
	. "github.com/goghcrow/go-parsec"
	"github.com/goghcrow/go-parsec/lexer"
	"github.com/goghcrow/simple-sub/terms"
)

//goland:noinspection GoSnakeCaseUsage
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

	IDENT

	FLOAT
	INT
	STR

	COLON
	COMMA
	ASSIGN

	ARROW
	DOT

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

type Keyword struct {
	lexer.TokenKind
	Lexeme string
}

var keywords = []Keyword{
	{LET, "let"},
	{REC, "rec"},
	{IN, "in"},
	{FUN, "fun"},
	{IF, "if"},
	{THEN, "then"},
	{ELSE, "else"},
	{TRUE, "true"},
	{FALSE, "false"},
}

var builtInOpers = []lexer.Operator{
	{DOT, ".", lexer.BP_MEMBER, lexer.INFIX_L},
	{ARROW, "->", lexer.BP_MEMBER, lexer.INFIX_R},
}

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

	for _, keyword := range keywords {
		l.Keyword(keyword.TokenKind, keyword.Lexeme)
	}

	for _, oper := range lexer.SortOpers(builtInOpers) {
		l.PrimOper(oper.TokenKind, oper.Lexeme)
	}

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
		return terms.Def(name, rhs, t4[0] != nil)
	}
	applyPgrm := func(v interface{}) interface{} {
		xs := v.([]interface{})
		defs := make([]*terms.Define, len(xs))
		for i, x := range xs {
			defs[i] = x.(*terms.Define)
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

	Term.Pattern = Alt(Let, Fun, Ite, Apps)
	Const.Pattern = Alt(
		Tok(INT).Map(applyInt),
		Tok(FLOAT).Map(applyFloat),
		Tok(STR).Map(applyStr),
		Tok(TRUE).Map(applyTrue),
		Tok(FALSE).Map(applyFalse),
	)

	atLeastIdent := Seq(Ident, RepSc(Ident)).Map(applyAtLeastIdent)
	Ident.Pattern = Alt(Tok(IDENT), Tok(TRUE), Tok(FALSE)).Map(applyIdent)
	Variable.Pattern = Tok(IDENT).Map(applyVar)
	Parens.Pattern = KMid(Tok(LEFT_PAREN), Term, Tok(RIGHT_PAREN))
	SubTermNoSel.Pattern = Alt(Parens, Record, Const, Variable)
	SubTerm.Pattern = Seq(SubTermNoSel, RepSc(KRight(Tok(DOT), Ident))).Map(applySubTerm)
	Record.Pattern = KMid(Tok(LEFT_BRACE), OptSc(ListSc(Seq(Ident, Tok(COLON), Term), Tok(COMMA))), Tok(RIGHT_BRACE)).Map(applyRecord)
	Fun.Pattern = Seq(Tok(FUN), atLeastIdent, Tok(ARROW), Term).Map(applyFun)
	Let.Pattern = Seq(KRight(Tok(LET), OptSc(Tok(REC))), Ident, Tok(ASSIGN), Term, Tok(IN), Term).Map(applyLet)
	Ite.Pattern = Seq(Tok(IF), Term, Tok(THEN), Term, Tok(ELSE), Term).Map(applyIte)
	Apps.Pattern = Seq(SubTerm, RepSc(SubTerm)).Map(applyApps)

	_Expr.Pattern = Term

	TopLevel.Pattern = Seq(KRight(Tok(LET), OptSc(Tok(REC))), Ident, Tok(ASSIGN), Term).Map(applyTopLevel)
	_Pgrm.Pattern = RepSc(TopLevel).Map(applyPgrm)
}

func ParseExpr(s string) (terms.Term, error) { return parse(_Expr, s) }

func ParsePgrm(s string) (terms.Term, error) { return parse(_Pgrm, s) }

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
