package parser

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/goghcrow/go-parsec"
)

func TestParser(t *testing.T) {
	for _, tt := range []struct {
		name    string
		p       Parser
		input   string
		success bool
		result  string
		error   string
	}{
		{
			name:    "Const int",
			p:       _Expr,
			input:   "42",
			success: true,
			result:  "Int(42)",
		},
		{
			name:    "Const int",
			p:       _Expr,
			input:   "-42",
			success: true,
			result:  "Int(-42)",
		},
		{
			name:    "Const float",
			p:       _Expr,
			input:   "3.14",
			success: true,
			result:  "Float(3.140000)",
		},
		{
			name:    "Const string",
			p:       _Expr,
			input:   `"Hello"`,
			success: true,
			result:  `Str("Hello")`,
		},
		{
			name:    "Const true",
			p:       _Expr,
			input:   `true`,
			success: true,
			result:  `Bool(true)`,
		},
		{
			name:    "Const false",
			p:       _Expr,
			input:   `false`,
			success: true,
			result:  `Bool(false)`,
		},
		{
			name:    "Variable",
			p:       _Expr,
			input:   `id`,
			success: true,
			result:  `Var(id)`,
		},
		{
			name:    "SubTermNoSel",
			p:       _Expr,
			input:   "42",
			success: true,
			result:  "Int(42)",
		},
		{
			name:    "SubTerm",
			p:       _Expr,
			input:   "42",
			success: true,
			result:  "Int(42)",
		},
		{
			name:    "Apps",
			p:       _Expr,
			input:   "42",
			success: true,
			result:  "Int(42)",
		},
		{
			name:    "Term",
			p:       _Expr,
			input:   "42",
			success: true,
			result:  "Int(42)",
		},
		{
			name:    "SubTermNoSel",
			p:       _Expr,
			input:   "id",
			success: true,
			result:  "Var(id)",
		},
		{
			name:    "SubTerm",
			p:       _Expr,
			input:   "id",
			success: true,
			result:  "Var(id)",
		},
		{
			name:    "Apps",
			p:       _Expr,
			input:   "id",
			success: true,
			result:  "Var(id)",
		},
		{
			name:    "Term",
			p:       _Expr,
			input:   "id",
			success: true,
			result:  "Var(id)",
		},
		{
			name:    "Parens",
			p:       _Expr,
			input:   `(1)`,
			success: true,
			result:  "Int(1)",
		},
		{
			name:    "Parens",
			p:       _Expr,
			input:   `(id)`,
			success: true,
			result:  "Var(id)",
		},
		{
			name:    "Record",
			p:       _Expr,
			input:   `{}`,
			success: true,
			result:  "Rcd([])",
		},
		{
			name:    "Record",
			p:       _Expr,
			input:   `{id: 42}`,
			success: true,
			result:  "Rcd([{id Int(42)}])",
		},
		{
			name:    "Record",
			p:       _Expr,
			input:   `{id: 42, name:"xiao"}`,
			success: true,
			result:  "Rcd([{id Int(42)} {name Str(\"xiao\")}])",
		},
		{
			name:    "Record",
			p:       _Expr,
			input:   `{true: true, false: false}`,
			success: true,
			result:  "Rcd([{true Bool(true)} {false Bool(false)}])",
		},
		{
			name:    "Tuple",
			p:       _Expr,
			input:   `()`,
			success: true,
			result:  "Tuple([])",
		},
		{
			name:    "Tuple",
			p:       _Expr,
			input:   `(1,)`, // 一个元素的 tuple 需要多个 comma
			success: true,
			result:  "Tuple([Int(1)])",
		},
		{
			name:    "Tuple",
			p:       _Expr,
			input:   `(1,2)`,
			success: true,
			result:  "Tuple([Int(1) Int(2)])",
		},
		{
			name:    "Tuple",
			p:       _Expr,
			input:   `(1,2, f a)`,
			success: true,
			result:  "Tuple([Int(1) Int(2) App(Var(f) Var(a))])",
		},
		{
			name:    "List",
			p:       _Expr,
			input:   `[]`,
			success: true,
			result:  "List([])",
		},
		{
			name:    "Tuple",
			p:       _Expr,
			input:   `[1]`, // 一个元素的 tuple 需要多个 comma
			success: true,
			result:  "List([Int(1)])",
		},
		{
			name:    "Tuple",
			p:       _Expr,
			input:   `[1,2]`,
			success: true,
			result:  "List([Int(1) Int(2)])",
		},
		{
			name:    "Tuple",
			p:       _Expr,
			input:   `[1,2, f a]`,
			success: true,
			result:  "List([Int(1) Int(2) App(Var(f) Var(a))])",
		},
		{
			name:    "Fun",
			p:       _Expr,
			input:   `fun id -> id`,
			success: true,
			result:  "Fun(id, Var(id))",
		},
		{
			name:    "Fun",
			p:       _Expr,
			input:   `fun n -> succ n`,
			success: true,
			result:  "Fun(n, App(Var(succ) Var(n)))",
		},
		{
			name:    "Let",
			p:       _Expr,
			input:   `let rec f = f in f`,
			success: true,
			result:  "LetRec(f, Var(f), Var(f))",
		},
		{
			name:    "Let",
			p:       _Expr,
			input:   `let f = f in f`,
			success: true,
			result:  "Let(f, Var(f), Var(f))",
		},
		{
			name:    "Let",
			p:       _Expr,
			input:   `let f = f f in f f`,
			success: true,
			result:  "Let(f, App(Var(f) Var(f)), App(Var(f) Var(f)))",
		},
		{
			name:    "Ite",
			p:       _Expr,
			input:   `if true then 42 else 100`,
			success: true,
			result:  "If(Bool(true), Int(42), Int(100))",
		},
		{
			name:    "Ite",
			p:       _Expr,
			input:   `if true then if false then 1 else 2 else 3`,
			success: true,
			result:  "If(Bool(true), If(Bool(false), Int(1), Int(2)), Int(3))",
		},
		{
			name:    "SubTerm",
			p:       _Expr,
			input:   `a.b.c.e`,
			success: true,
			result:  "Sel(Sel(Sel(Var(a), b), c), e)",
		},
		{
			name:    "Apps",
			p:       _Expr,
			input:   `f 1 2 3`,
			success: true,
			result:  "App(App(App(Var(f) Int(1)) Int(2)) Int(3))",
		},
		{
			name:    "Apps",
			p:       _Expr,
			input:   `a.b.c.d x.y.z 42`,
			success: true,
			result:  "App(App(Sel(Sel(Sel(Var(a), b), c), d) Sel(Sel(Var(x), y), z)) Int(42))",
		},
		{
			name:    "Apps",
			p:       _Expr,
			input:   `f1 (f2 a) b`,
			success: true,
			result:  "App(App(Var(f1) App(Var(f2) Var(a))) Var(b))",
		},
		{
			name:    "TopLevel",
			p:       _Pgrm,
			input:   `let twice = fun f x -> f (f x)`,
			success: true,
			result:  "Program([LetRec(twice, Fun(f, Fun(x, App(Var(f) App(Var(f) Var(x))))))])",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			toks := lex.MustLex(tt.input)
			out := tt.p.Parse(toks)
			if tt.success {
				xs := succeed(out)
				actual := fmtResults(xs)
				if actual != tt.result {
					t.Errorf("expect %s actual %s", tt.result, actual)
				}
				if tt.error != "" {
					actual = out.Error.Error()
					if actual != tt.error {
						t.Errorf("expect %s actual %s", tt.error, actual)
					}
				}
			} else {
				if out.Success {
					t.Errorf("expect fail actual success")
				}
				actual := out.Error.Error()
				if actual != tt.error {
					t.Errorf("expect %s actual %s", tt.error, actual)
				}
			}
		})
	}
}

func succeed(out Output) []Result {
	if out.Success {
		return out.Candidates
	}
	panic(out)
}

func fmtResults(results []Result) string {
	xs := make([]string, len(results))
	for i, r := range results {
		xs[i] = fmt.Sprintf("%s", r.Val)
	}
	return strings.Join(xs, "🍊")
}
