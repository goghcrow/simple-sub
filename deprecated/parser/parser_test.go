package parser

import (
	"github.com/goghcrow/simple-sub/deprecated/lexer"
	"github.com/goghcrow/simple-sub/deprecated/oper"
	. "github.com/goghcrow/simple-sub/terms"
	"testing"
)

func parse(s string) Term {
	ops := oper.BuildIn()
	lex := lexer.NewLexer(ops)
	p := NewParser(ops)
	p.toks = lex.Lex(s)
	return p.expr(0)
}

func parsePgrm(s string) *Program {
	ops := oper.BuildIn()
	lex := lexer.NewLexer(ops)
	return NewParser(ops).Parse(lex.Lex(s))
}

func TestParsePgrm(t *testing.T) {
	s := `
// 单行注释
/*
多行注释
*/
let id = fun x -> x
let twice = fun f -> fun x -> f (f (x))
let object1 = { x: 42, y: id }
let object2 = { x: 17, y: false }
let pick_an_object = fun b ->
  if b then object1 else object2
let rec recursive_monster = fun x -> { thing: x, self: recursive_monster(x) }
`
	for _, def := range parsePgrm(s).Defs {
		t.Log(def)
	}
}

func TestParser(t *testing.T) {
	for _, tt := range []struct {
		s      string
		expect Term
	}{
		{"true", Bool(true)},
		{"false", Bool(false)},
		{"1", Int(1)},
		{"0xffff", Int(0xffff)},
		{"0o42", Int(0o42)},
		{"0b10101010", Int(0b10101010)},
		{"3.14", Float(3.14)},
		{"3e1", Float(3e1)},
		{"3E-1", Float(3e-1)},
		{"3.14e+14", Float(3.14e+14)},
		{`"Hello World!"`, Str("Hello World!")},
		{"`Hello World!`", Str("Hello World!")},

		// tuple1 必须加个逗号, 否则与 group 语法有歧义
		{"() ", Tup()},
		{"(1)", Grp(Int(1))},
		{"(1,)", Tup(Int(1))},
		{"(1,2)", Tup(Int(1), Int(2))},
		{"(1,2,)", Tup(Int(1), Int(2))},
		{"((),)", Tup(Tup())},

		{"{}", Rcd([]Field{})},
		{"{tag: 1}", Rcd([]Field{{"tag", Int(1)}})},
		{"{tag: 1,}", Rcd([]Field{{"tag", Int(1)}})},
		{"{tag: 1, id:42}", Rcd([]Field{{"tag", Int(1)}, {"id", Int(42)}})},
		{"{tag: 1, id:42,}", Rcd([]Field{{"tag", Int(1)}, {"id", Int(42)}})},
		{"{tag: 1, nested: {id: 42}}", Rcd([]Field{{"tag", Int(1)}, {"nested", Rcd([]Field{{"id", Int(42)}})}})},

		{"{tag: 1}.tag", Sel(Rcd([]Field{{"tag", Int(1)}}), "tag")},
		{"rcd.tag", Sel(Var("rcd"), "tag")},
		{"a.b.c", Sel(Sel(Var("a"), "b"), "c")},
		{"(a.b).c", Sel(Grp(Sel(Var("a"), "b")), "c")},

		{"if true then 1 else 2", Iff(Bool(true), Int(1), Int(2))},
		{"if true then if false then 1 else 2 else 3", Iff(Bool(true), Iff(Bool(false), Int(1), Int(2)), Int(3))},
		{"if true then (if false then 1 else 2) else 3", Iff(Bool(true), Grp(Iff(Bool(false), Int(1), Int(2))), Int(3))},

		{"let a = 42 in a", Let("a", Int(42), Var("a"), false)},
		{"let rec a = 42 in a", Let("a", Int(42), Var("a"), true)},
		{"let f a b = a + b in f(1, 2)", Let("f", LamN([]string{"a", "b"}, Bin("+", oper.INFIX_L, Var("a"), Var("b"))), AppN(Var("f"), Int(1), Int(2)), false)},
		{"let rec f a b = a + b in f(1, 2)", Let("f", LamN([]string{"a", "b"}, Bin("+", oper.INFIX_L, Var("a"), Var("b"))), AppN(Var("f"), Int(1), Int(2)), true)},
		{"let rec f = fun a b -> a + b in f(1, 2)", Let("f", LamN([]string{"a", "b"}, Bin("+", oper.INFIX_L, Var("a"), Var("b"))), AppN(Var("f"), Int(1), Int(2)), true)},
		{"let rec f = a -> b -> a + b in f(1, 2)", Let("f", LamN([]string{"a", "b"}, Bin("+", oper.INFIX_L, Var("a"), Var("b"))), AppN(Var("f"), Int(1), Int(2)), true)},

		// 至少一个参数
		{"fun a -> a", Lam("a", Var("a"))},
		{"fun a b -> a + b", Lam("a", Lam("b", Bin("+", oper.INFIX_L, Var("a"), Var("b"))))},
		{"fun a b c -> a + b + c", Lam("a", Lam("b", Lam("c", Bin("+", oper.INFIX_L, Bin("+", oper.INFIX_L, Var("a"), Var("b")), Var("c")))))},

		{"a -> a", Lam("a", Var("a"))},
		{"_ -> ()", Lam("_", Tup())},
		{"a -> b -> a + b", Lam("a", Lam("b", Bin("+", oper.INFIX_L, Var("a"), Var("b"))))},
		{"a -> b -> c -> a + b + c", Lam("a", Lam("b", Lam("c", Bin("+", oper.INFIX_L, Bin("+", oper.INFIX_L, Var("a"), Var("b")), Var("c")))))},

		{"f(_)", App(Var("f"), Var("_"))},
		{"f(())", App(Var("f"), Tup())},
		{"f(a, b)", AppN(Var("f"), Var("a"), Var("b"))},
		{"(a -> a)(a)", App(Grp(Lam("a", Var("a"))), Var("a"))},
		{"(fun a b -> a + b)(1, 2)", AppN(Grp(Lam("a", Lam("b", Bin("+", oper.INFIX_L, Var("a"), Var("b"))))), Int(1), Int(2))},
	} {
		t.Run(tt.s, func(t *testing.T) {
			actual := parse(tt.s).String()
			expect := tt.expect.String()
			t.Logf("expect %s", expect)
			t.Logf("actual %s", actual)
			if actual != expect {
				t.Errorf("expect %s actual %s", expect, actual)
			}
		})
	}
}
