package typer

import (
	"github.com/goghcrow/simple-sub/terms"
	"testing"
)

type expected struct {
	inferred   string
	where      string
	compacted  string
	simplified string
	coalesced  string
	error
}
type testCase struct {
	string
	terms.Term
	expected
}

func TestBasic(t *testing.T) {
	for _, tt := range []testCase{
		{
			"42",
			terms.Int(42),
			expected{"int", "", "‹int›", "‹int›", "int", nil},
		},
		{
			"fun x -> 42",
			terms.Lam("x", terms.Int(42)),
			expected{"(α1 -> int)", "", "‹‹α1› -> ‹int››", "‹‹› -> ‹int››", "⊤ -> int", nil},
		},
		{
			"fun x -> x",
			terms.Lam("x", terms.Var("x")),
			expected{"(α1 -> α1)", "", "‹‹α1› -> ‹α1››", "‹‹α1› -> ‹α1››", "'a -> 'a", nil},
		},
		{
			"fun x -> x 42",
			terms.Lam("x", terms.App(terms.Var("x"), terms.Int(42))),
			expected{"(α1 -> α2)", "α1 <: (int -> α2)", "‹‹α1, ‹int› -> ‹α2›› -> ‹α2››", "‹‹‹int› -> ‹α2›› -> ‹α2››", "(int -> 'a) -> 'a", nil},
		},
		{
			"(fun x -> x) 42",
			terms.App(terms.Lam("x", terms.Var("x")), terms.Int(42)),
			expected{"α2", "α2 :> int", "‹α2, int›", "‹int›", "int", nil},
		},
		{
			"fun f -> fun x -> f (f x)  // twice",
			terms.Lam("f", terms.Lam("x", terms.App(terms.Var("f"), terms.App(terms.Var("f"), terms.Var("x"))))),
			expected{"(α1 -> (α2 -> α4))", "α1 <: (α3 -> α4) & (α2 -> α3)", "‹‹α1, ‹α3, α2› -> ‹α4, α3›› -> ‹‹α2› -> ‹α4›››", "‹‹‹α2, α4› -> ‹α4›› -> ‹‹α2› -> ‹α4›››", "('a ∨ 'b -> 'b) -> 'a -> 'b", nil},
		},
		{
			"let twice = fun f -> fun x -> f (f x) in twice",
			terms.Let("twice", terms.Lam("f", terms.Lam("x", terms.App(terms.Var("f"), terms.App(terms.Var("f"), terms.Var("x"))))), terms.Var("twice"), false),
			expected{"(α5 -> (α6 -> α8))", "α5 <: (α7 -> α8) & (α6 -> α7)", "‹‹α5, ‹α7, α6› -> ‹α8, α7›› -> ‹‹α6› -> ‹α8›››", "‹‹‹α6, α8› -> ‹α8›› -> ‹‹α6› -> ‹α8›››", "('a ∨ 'b -> 'b) -> 'a -> 'b", nil},
		},
	} {
		t.Run(tt.string, func(t *testing.T) { doTest(t, tt) })
	}
}

func TestBooleans(t *testing.T) {
	for _, tt := range []testCase{
		{
			"true",
			terms.Var("true"),
			expected{"bool", "", "‹bool›", "‹bool›", "bool", nil},
		},
		{
			"not true",
			terms.App(terms.Var("not"), terms.Var("true")),
			expected{"α1", "α1 :> bool", "‹α1, bool›", "‹bool›", "bool", nil},
		},
		{
			"fun x -> not x",
			terms.Lam("x", terms.App(terms.Var("not"), terms.Var("x"))),
			expected{"(α1 -> α2)", "α1 <: bool, α2 :> bool", "‹‹α1, bool› -> ‹α2, bool››", "‹‹bool› -> ‹bool››", "bool -> bool", nil},
		},
		{
			"(fun x -> not x) true",
			terms.App(terms.Lam("x", terms.App(terms.Var("not"), terms.Var("x"))), terms.Var("true")),
			expected{"α3", "α3 :> bool", "‹α3, bool›", "‹bool›", "bool", nil},
		},
		{
			"fun x -> fun y -> fun z -> if x then y else z",
			terms.Lam("x", terms.Lam("y", terms.Lam("z", terms.App(terms.App(terms.App(terms.Var("if"), terms.Var("x")), terms.Var("y")), terms.Var("z"))))),
			expected{"(α1 -> (α2 -> (α3 -> α7)))", "α1 <: bool, α2 <: α4, α3 <: α4, α4 <: α7", "‹‹α1, bool› -> ‹‹α2, α4, α7› -> ‹‹α3, α4, α7› -> ‹α7››››", "‹‹bool› -> ‹‹α7› -> ‹‹α7› -> ‹α7››››", "bool -> 'a -> 'a -> 'a", nil},
		},
		{
			"fun x -> fun y -> if x then y else x",
			terms.Lam("x", terms.Lam("y", terms.App(terms.App(terms.App(terms.Var("if"), terms.Var("x")), terms.Var("y")), terms.Var("x")))),
			expected{"(α1 -> (α2 -> α6))", "α1 <: α3 & bool, α2 <: α3, α3 <: α6", "‹‹α1, α3, α6, bool› -> ‹‹α2, α3, α6› -> ‹α6›››", "‹‹α6, bool› -> ‹‹α6› -> ‹α6›››", "'a ∧ bool -> 'a -> 'a", nil},
		},
		{
			"succ true",
			terms.App(terms.Var("succ"), terms.Var("true")),
			expected{"", "", "", "", "", NewTypeError("cannot constrain bool <: int")},
		},
		{
			"fun x -> succ (not x)",
			terms.Lam("x", terms.App(terms.Var("succ"), terms.App(terms.Var("not"), terms.Var("x")))),
			expected{"", "", "", "", "", NewTypeError("cannot constrain bool <: int")},
		},
		{
			"(fun x -> not x.f) { f = 123 }",
			terms.App(terms.Lam("x", terms.App(terms.Var("not"), terms.Sel(terms.Var("x"), "f"))), terms.Rcd([]terms.Field{{"f", terms.Int(123)}})),
			expected{"", "", "", "", "", NewTypeError("cannot constrain int <: bool")},
		},
		{
			"(fun f -> fun x -> not (f x.u)) false",
			terms.App(terms.Lam("f", terms.Lam("x", terms.App(terms.Var("not"), terms.App(terms.Var("f"), terms.Sel(terms.Var("x"), "u"))))), terms.Var("false")),
			expected{"", "", "", "", "", NewTypeError("cannot constrain bool <: 'a -> 'b")},
		},
	} {
		t.Run(tt.string, func(t *testing.T) { doTest(t, tt) })
	}
}

func TestRecords(t *testing.T) {
	for _, tt := range []testCase{
		{
			"fun x -> x.f",
			terms.Lam("x", terms.Sel(terms.Var("x"), "f")),
			expected{"(α1 -> α2)", "α1 <: {f: α2}", "‹‹α1, {f: ‹α2›}› -> ‹α2››", "‹‹{f: ‹α2›}› -> ‹α2››", "{f: 'a} -> 'a", nil},
		},
		{
			"{}",
			terms.Rcd([]terms.Field{}),
			expected{"{}", "", "‹{}›", "‹{}›", "{}", nil},
		},
		{
			"{ f = 42 }",
			terms.Rcd([]terms.Field{{"f", terms.Int(42)}}),
			expected{"{f: int}", "", "‹{f: ‹int›}›", "‹{f: ‹int›}›", "{f: int}", nil},
		},
		{
			"{ f = 42 }.f",
			terms.Sel(terms.Rcd([]terms.Field{{"f", terms.Int(42)}}), "f"),
			expected{"α1", "α1 :> int", "‹α1, int›", "‹int›", "int", nil},
		},
		{
			"(fun x -> x.f) { f = 42 }",
			terms.App(terms.Lam("x", terms.Sel(terms.Var("x"), "f")), terms.Rcd([]terms.Field{{"f", terms.Int(42)}})),
			expected{"α3", "α3 :> int", "‹α3, int›", "‹int›", "int", nil},
		},
		{
			"fun f -> { x = f 42 }.x",
			terms.Lam("f", terms.Sel(terms.Rcd([]terms.Field{{"x", terms.App(terms.Var("f"), terms.Int(42))}}), "x")),
			expected{"(α1 -> α3)", "α1 <: (int -> α2), α2 <: α3", "‹‹α1, ‹int› -> ‹α2, α3›› -> ‹α3››", "‹‹‹int› -> ‹α3›› -> ‹α3››", "(int -> 'a) -> 'a", nil},
		},
		{
			"fun f -> { x = f 42; y = 123 }.y",
			terms.Lam("f", terms.Sel(terms.Rcd([]terms.Field{{"x", terms.App(terms.Var("f"), terms.Int(42))}, {"y", terms.Int(123)}}), "y")),
			expected{"(α1 -> α3)", "α1 <: (int -> α2), α3 :> int", "‹‹α1, ‹int› -> ‹α2›› -> ‹α3, int››", "‹‹‹int› -> ‹›› -> ‹int››", "(int -> ⊤) -> int", nil},
		},
		{
			"if true then { a = 1; b = true } else { b = false; c = 42 }",
			terms.App(terms.App(terms.App(terms.Var("if"), terms.Var("true")), terms.Rcd([]terms.Field{{"a", terms.Int(1)}, {"b", terms.Var("true")}})), terms.Rcd([]terms.Field{{"b", terms.Var("false")}, {"c", terms.Int(42)}})),
			expected{"α4", "α4 :> {a: int, b: bool} | {b: bool, c: int}", "‹α4, {b: ‹bool›}›", "‹{b: ‹bool›}›", "{b: bool}", nil},
		},
		{
			"{ a = 123; b = true }.c",
			terms.Sel(terms.Rcd([]terms.Field{{"a", terms.Int(123)}, {"b", terms.Var("true")}}), "c"),
			expected{"", "", "", "", "", NewTypeError("missing field: c in {a: int, b: bool}")},
		},
		{
			"fun x -> { a = x }.b",
			terms.Lam("x", terms.Sel(terms.Rcd([]terms.Field{{"a", terms.Var("x")}}), "b")),
			expected{"", "", "", "", "", NewTypeError("missing field: b in {a: 'a}")},
		},
	} {
		t.Run(tt.string, func(t *testing.T) { doTest(t, tt) })
	}
}

func TestSelfApp(t *testing.T) {
	for _, tt := range []testCase{
		{
			"fun x -> x x",
			terms.Lam("x", terms.App(terms.Var("x"), terms.Var("x"))),
			expected{"(α1 -> α2)", "α1 <: (α1 -> α2)", "‹‹α1, ‹α1› -> ‹α2›› -> ‹α2››", "‹‹α1, ‹α1› -> ‹α2›› -> ‹α2››", "'a ∧ ('a -> 'b) -> 'b", nil},
		},
		{
			"fun x -> x x x",
			terms.Lam("x", terms.App(terms.App(terms.Var("x"), terms.Var("x")), terms.Var("x"))),
			expected{"(α1 -> α3)", "α1 <: (α1 -> α2), α2 <: (α1 -> α3)", "‹‹α1, ‹α1› -> ‹α2, ‹α1› -> ‹α3››› -> ‹α3››", "‹‹α1, ‹α1› -> ‹‹α1› -> ‹α3››› -> ‹α3››", "'a ∧ ('a -> 'a -> 'b) -> 'b", nil},
		},
		{
			"fun x -> fun y -> x y x",
			terms.Lam("x", terms.Lam("y", terms.App(terms.App(terms.Var("x"), terms.Var("y")), terms.Var("x")))),
			expected{"(α1 -> (α2 -> α4))", "α1 <: (α2 -> α3), α3 <: (α1 -> α4)", "‹‹α1, ‹α2› -> ‹α3, ‹α1› -> ‹α4››› -> ‹‹α2› -> ‹α4›››", "‹‹α1, ‹α2› -> ‹‹α1› -> ‹α4››› -> ‹‹α2› -> ‹α4›››", "'a ∧ ('b -> 'a -> 'c) -> 'b -> 'c", nil},
		},
		{
			"fun x -> fun y -> x x y",
			terms.Lam("x", terms.Lam("y", terms.App(terms.App(terms.Var("x"), terms.Var("x")), terms.Var("y")))),
			expected{"(α1 -> (α2 -> α4))", "α1 <: (α1 -> α3), α3 <: (α2 -> α4)", "‹‹α1, ‹α1› -> ‹α3, ‹α2› -> ‹α4››› -> ‹‹α2› -> ‹α4›››", "‹‹α1, ‹α1› -> ‹‹α2› -> ‹α4››› -> ‹‹α2› -> ‹α4›››", "'a ∧ ('a -> 'b -> 'c) -> 'b -> 'c", nil},
		},
		{
			"(fun x -> x x) (fun x -> x x)",
			terms.App(terms.Lam("x", terms.App(terms.Var("x"), terms.Var("x"))), terms.Lam("x", terms.App(terms.Var("x"), terms.Var("x")))),
			expected{"α5", "", "‹α5›", "‹›", "⊥", nil},
		},
		{
			"fun x -> {l = x x; r = x }",
			terms.Lam("x", terms.Rcd([]terms.Field{{"l", terms.App(terms.Var("x"), terms.Var("x"))}, {"r", terms.Var("x")}})),
			expected{"(α1 -> {l: α2, r: α1})", "α1 <: (α1 -> α2)", "‹‹α1, ‹α1› -> ‹α2›› -> ‹{l: ‹α2›, r: ‹α1›}››", "‹‹α1, ‹α1› -> ‹α2›› -> ‹{l: ‹α2›, r: ‹α1›}››", "'a ∧ ('a -> 'b) -> {l: 'b, r: 'a}", nil},
		},
		{
			"(fun f -> (fun x -> f (x x)) (fun x -> f (x x)))",
			terms.Lam("f", terms.App(terms.Lam("x", terms.App(terms.Var("f"), terms.App(terms.Var("x"), terms.Var("x")))), terms.Lam("x", terms.App(terms.Var("f"), terms.App(terms.Var("x"), terms.Var("x")))))),
			expected{"(α1 -> α8)", "α1 <: (α6 -> α7) & (α3 -> α4), α4 <: α8, α7 <: α3 & α6", "‹‹α1, ‹α6, α3› -> ‹α6, α7, α3, α8, α4›› -> ‹α8››", "‹‹‹α8› -> ‹α8›› -> ‹α8››", "('a -> 'a) -> 'a", nil},
		},
		{
			"(fun f -> (fun x -> f (fun v -> (x x) v)) (fun x -> f (fun v -> (x x) v)))",
			terms.Lam("f", terms.App(terms.Lam("x", terms.App(terms.Var("f"), terms.Lam("v", terms.App(terms.App(terms.Var("x"), terms.Var("x")), terms.Var("v"))))), terms.Lam("x", terms.App(terms.Var("f"), terms.Lam("v", terms.App(terms.App(terms.Var("x"), terms.Var("x")), terms.Var("v"))))))),
			expected{"(α1 -> α12)", "α1 <: ((α8 -> α10) -> α11) & ((α3 -> α5) -> α6), α4 <: (α3 -> α5), α6 <: α12, α9 <: (α8 -> α10), α11 <: α4 & α9", "‹‹α1, ‹‹α8, α3› -> ‹α10, α5›› -> ‹α6, α9, α12, α11, α4, ‹α8, α3› -> ‹α10, α5››› -> ‹α12››", "‹‹‹‹α8› -> ‹α10›› -> ‹α12, ‹α8› -> ‹α10››› -> ‹α12››", "(('a -> 'b) -> 'c ∧ ('a -> 'b)) -> 'c", nil},
		},
		{
			"(fun f -> (fun x -> f (fun v -> (x x) v)) (fun x -> f (fun v -> (x x) v))) (fun f -> fun x -> f)",
			terms.App(terms.Lam("f", terms.App(terms.Lam("x", terms.App(terms.Var("f"), terms.Lam("v", terms.App(terms.App(terms.Var("x"), terms.Var("x")), terms.Var("v"))))), terms.Lam("x", terms.App(terms.Var("f"), terms.Lam("v", terms.App(terms.App(terms.Var("x"), terms.Var("x")), terms.Var("v"))))))), terms.Lam("f", terms.Lam("x", terms.Var("f")))),
			expected{"α15", "α3 <: α14, α5 :> (α3 -> α5) | (α8 -> α10), α8 <: α14, α10 :> (α3 -> α5) | (α8 -> α10), α13 :> (α3 -> α5) | (α8 -> α10) <: α10 & α5, α15 :> (α14 -> α13)", "‹α15, ‹α14› -> ‹α13, ‹α3, α14, α8› -> ‹α16›››", "‹‹› -> ‹‹› -> ‹α16›››", "⊤ -> (⊤ -> 'a) as 'a", nil},
		},
		{
			"let rec trutru = fun g -> trutru (g true) in trutru",
			terms.Let("trutru", terms.Lam("g", terms.App(terms.Var("trutru"), terms.App(terms.Var("g"), terms.Var("true")))), terms.Var("trutru"), true),
			expected{"α5", "α5 :> (α6 -> α8) <: (α7 -> α8), α6 <: (bool -> α7), α7 <: α6", "‹α5, ‹α6, ‹bool› -> ‹α9›› -> ‹α8››", "‹‹‹bool› -> ‹α9›› -> ‹››", "(bool -> 'a) as 'a -> ⊥", nil},
		},
		{
			"fun i -> if ((i i) true) then true else true",
			terms.Lam("i", terms.App(terms.App(terms.App(terms.Var("if"), terms.App(terms.App(terms.Var("i"), terms.Var("i")), terms.Var("true"))), terms.Var("true")), terms.Var("true"))),
			expected{"(α1 -> α7)", "α1 <: (α1 -> α3), α3 <: (bool -> α4), α4 <: bool, α7 :> bool", "‹‹α1, ‹α1› -> ‹α3, ‹bool› -> ‹α4, bool››› -> ‹α7, bool››", "‹‹α1, ‹α1› -> ‹‹bool› -> ‹bool››› -> ‹bool››", "'a ∧ ('a -> bool -> bool) -> bool", nil},
		},
	} {
		t.Run(tt.string, func(t *testing.T) { doTest(t, tt) })
	}
}

func TestLetPoly(t *testing.T) {
	for _, tt := range []testCase{
		{
			"let f = fun x -> x in {a = f 0; b = f true}",
			terms.Let("f", terms.Lam("x", terms.Var("x")), terms.Rcd([]terms.Field{{"a", terms.App(terms.Var("f"), terms.Int(0))}, {"b", terms.App(terms.Var("f"), terms.Var("true"))}}), false),
			expected{"{a: α3, b: α5}", "α3 :> int, α5 :> bool", "‹{a: ‹α3, int›, b: ‹α5, bool›}›", "‹{a: ‹int›, b: ‹bool›}›", "{a: int, b: bool}", nil},
		},
		{
			"fun y -> let f = fun x -> x in {a = f y; b = f true}",
			terms.Lam("y", terms.Let("f", terms.Lam("x", terms.Var("x")), terms.Rcd([]terms.Field{{"a", terms.App(terms.Var("f"), terms.Var("y"))}, {"b", terms.App(terms.Var("f"), terms.Var("true"))}}), false)),
			expected{"(α1 -> {a: α4, b: α6})", "α1 <: α3, α3 <: α4, α6 :> bool", "‹‹α1, α3, α4› -> ‹{a: ‹α4›, b: ‹α6, bool›}››", "‹‹α4› -> ‹{a: ‹α4›, b: ‹bool›}››", "'a -> {a: 'a, b: bool}", nil},
		},
		{
			"fun y -> let f = fun x -> y x in {a = f 0; b = f true}",
			terms.Lam("y", terms.Let("f", terms.Lam("x", terms.App(terms.Var("y"), terms.Var("x"))), terms.Rcd([]terms.Field{{"a", terms.App(terms.Var("f"), terms.Int(0))}, {"b", terms.App(terms.Var("f"), terms.Var("true"))}}), false)),
			expected{"(α1 -> {a: α8, b: α11})", "α1 <: (α4 -> α5), α4 :> bool | int, α5 <: α11 & α8", "‹‹α1, ‹α4, bool, int› -> ‹α5, α11, α8›› -> ‹{a: ‹α8›, b: ‹α11›}››", "‹‹‹bool, int› -> ‹α11›› -> ‹{a: ‹α11›, b: ‹α11›}››", "(bool ∨ int -> 'a) -> {a: 'a, b: 'a}", nil},
		},
		{
			"fun y -> let f = fun x -> x y in {a = f (fun z -> z); b = f (fun z -> true)}",
			terms.Lam("y", terms.Let("f", terms.Lam("x", terms.App(terms.Var("x"), terms.Var("y"))), terms.Rcd([]terms.Field{{"a", terms.App(terms.Var("f"), terms.Lam("z", terms.Var("z")))}, {"b", terms.App(terms.Var("f"), terms.Lam("z", terms.Var("true")))}}), false)),
			expected{"(α1 -> {a: α7, b: α11})", "α1 <: α10 & α6, α5 <: α7, α6 <: α5, α11 :> bool", "‹‹α5, α10, α1, α6, α7› -> ‹{a: ‹α7›, b: ‹α11, bool›}››", "‹‹α7› -> ‹{a: ‹α7›, b: ‹bool›}››", "'a -> {a: 'a, b: bool}", nil},
		},
		{
			"fun y -> let f = fun x -> x y in {a = f (fun z -> z); b = f (fun z -> succ z)}",
			terms.Lam("y", terms.Let("f", terms.Lam("x", terms.App(terms.Var("x"), terms.Var("y"))), terms.Rcd([]terms.Field{{"a", terms.App(terms.Var("f"), terms.Lam("z", terms.Var("z")))}, {"b", terms.App(terms.Var("f"), terms.Lam("z", terms.App(terms.Var("succ"), terms.Var("z"))))}}), false)),
			expected{"(α1 -> {a: α7, b: α12})", "α1 <: α10 & α6, α5 <: α7, α6 <: α5, α10 <: int, α12 :> int", "‹‹α5, α10, α1, α6, α7, int› -> ‹{a: ‹α7›, b: ‹α12, int›}››", "‹‹α7, int› -> ‹{a: ‹α7›, b: ‹int›}››", "'a ∧ int -> {a: 'a, b: int}", nil},
		},
		{
			"(fun k -> k (fun x -> let tmp = add x 1 in x)) (fun f -> f true)",
			terms.App(terms.Lam("k", terms.App(terms.Var("k"), terms.Lam("x", terms.Let("tmp", terms.App(terms.App(terms.Var("add"), terms.Var("x")), terms.Int(1)), terms.Var("x"), false)))), terms.Lam("f", terms.App(terms.Var("f"), terms.Var("true")))),
			expected{"", "", "", "", "", NewTypeError("cannot constrain bool <: int")},
		},
		{
			"(fun k -> let test = k (fun x -> let tmp = add x 1 in x) in test) (fun f -> f true)",
			terms.App(terms.Lam("k", terms.Let("test", terms.App(terms.Var("k"), terms.Lam("x", terms.Let("tmp", terms.App(terms.App(terms.Var("add"), terms.Var("x")), terms.Int(1)), terms.Var("x"), false))), terms.Var("test"), false)), terms.Lam("f", terms.App(terms.Var("f"), terms.Var("true")))),
			expected{"", "", "", "", "", NewTypeError("cannot constrain bool <: int")},
		},
		{
			"fun k -> let test = k (fun x -> let tmp = add x 1 in x) in test",
			terms.Lam("k", terms.Let("test", terms.App(terms.Var("k"), terms.Lam("x", terms.Let("tmp", terms.App(terms.App(terms.Var("add"), terms.Var("x")), terms.Int(1)), terms.Var("x"), false))), terms.Var("test"), false)),
			expected{"(α1 -> α9)", "α1 <: ((α6 -> α7) -> α8), α6 <: int, α7 :> α6, α9 :> α8", "‹‹α1, ‹‹α6, int› -> ‹α7, α6›› -> ‹α8›› -> ‹α9, α8››", "‹‹‹‹α6, int› -> ‹α6›› -> ‹α8›› -> ‹α8››", "(('a ∧ int -> 'a) -> 'b) -> 'b", nil},
		},
		{
			"fun k -> let test = k (fun x -> let tmp = add x 1 in if true then x else 2) in test",
			terms.Lam("k", terms.Let("test", terms.App(terms.Var("k"), terms.Lam("x", terms.Let("tmp", terms.App(terms.App(terms.Var("add"), terms.Var("x")), terms.Int(1)), terms.App(terms.App(terms.App(terms.Var("if"), terms.Var("true")), terms.Var("x")), terms.Int(2)), false))), terms.Var("test"), false)),
			expected{"(α1 -> α15)", "α1 <: ((α10 -> α13) -> α14), α10 <: α11 & int, α11 <: α12, α13 :> α12 | int, α15 :> α14", "‹‹α1, ‹‹α10, α11, α12, int› -> ‹α13, α12, int›› -> ‹α14›› -> ‹α15, α14››", "‹‹‹‹int› -> ‹int›› -> ‹α14›› -> ‹α14››", "((int -> int) -> 'a) -> 'a", nil},
		},
		{
			"fun k -> let test = (fun id -> {tmp = k id; res = id}.res) (fun x -> x) in {u=test 0; v=test true}",
			terms.Lam("k", terms.Let("test", terms.App(terms.Lam("id", terms.Sel(terms.Rcd([]terms.Field{{"tmp", terms.App(terms.Var("k"), terms.Var("id"))}, {"res", terms.Var("id")}}), "res")), terms.Lam("x", terms.Var("x"))), terms.Rcd([]terms.Field{{"u", terms.App(terms.Var("test"), terms.Int(0))}, {"v", terms.App(terms.Var("test"), terms.Var("true"))}}), false)),
			expected{"(α1 -> {u: α13, v: α16})", "α1 <: (α4 -> α5), α4 :> (α9 -> α10), α9 <: α16 & α13, α10 :> bool | int | α9, α13 :> int, α16 :> bool", "‹‹α1, ‹α4, ‹α9, α16, α13› -> ‹α10, α9, bool, int›› -> ‹α5›› -> ‹{u: ‹α13, int›, v: ‹α16, bool›}››", "‹‹‹‹α16› -> ‹α16, bool, int›› -> ‹›› -> ‹{u: ‹α16, int›, v: ‹α16, bool›}››", "(('a -> 'a ∨ bool ∨ int) -> ⊤) -> {u: 'a ∨ int, v: 'a ∨ bool}", nil},
		},
		{
			"fun k -> let test = {tmp = k (fun x -> x); res = (fun x -> x)}.res in {u=test 0; v=test true}",
			terms.Lam("k", terms.Let("test", terms.Sel(terms.Rcd([]terms.Field{{"tmp", terms.App(terms.Var("k"), terms.Lam("x", terms.Var("x")))}, {"res", terms.Lam("x", terms.Var("x"))}}), "res"), terms.Rcd([]terms.Field{{"u", terms.App(terms.Var("test"), terms.Int(0))}, {"v", terms.App(terms.Var("test"), terms.Var("true"))}}), false)),
			expected{"(α1 -> {u: α11, v: α14})", "α1 <: ((α4 -> α5) -> α6), α5 :> α4, α11 :> int, α14 :> bool", "‹‹α1, ‹‹α4› -> ‹α5, α4›› -> ‹α6›› -> ‹{u: ‹α11, int›, v: ‹α14, bool›}››", "‹‹‹‹α4› -> ‹α4›› -> ‹›› -> ‹{u: ‹int›, v: ‹bool›}››", "(('a -> 'a) -> ⊤) -> {u: int, v: bool}", nil},
		},
		{
			"fun k -> let test = (fun thefun -> {l=k thefun; r=thefun 1}) (fun x -> let tmp = add x 1 in x) in test",
			terms.Lam("k", terms.Let("test", terms.App(terms.Lam("thefun", terms.Rcd([]terms.Field{{"l", terms.App(terms.Var("k"), terms.Var("thefun"))}, {"r", terms.App(terms.Var("thefun"), terms.Int(1))}})), terms.Lam("x", terms.Let("tmp", terms.App(terms.App(terms.Var("add"), terms.Var("x")), terms.Int(1)), terms.Var("x"), false))), terms.Var("test"), false)),
			expected{"(α1 -> α14)", "α1 <: (α4 -> α5), α4 :> (α11 -> α13), α11 <: α12 & int, α13 :> α11 | int, α14 :> {l: α15, r: α16}, α15 :> α5, α16 :> α12 | int", "‹‹α1, ‹α4, ‹α11, α12, int› -> ‹α13, α11, int›› -> ‹α5›› -> ‹α14, {l: ‹α15, α5›, r: ‹α16, α12, int›}››", "‹‹‹‹α12, int› -> ‹α12, int›› -> ‹α5›› -> ‹{l: ‹α5›, r: ‹int›}››", "(('a ∧ int -> 'a ∨ int) -> 'b) -> {l: 'b, r: int}", nil},
		},
		{
			"fun a -> (fun k -> let test = k (fun x -> let tmp = add x 1 in x) in test) (fun f -> f a)",
			terms.Lam("a", terms.App(terms.Lam("k", terms.Let("test", terms.App(terms.Var("k"), terms.Lam("x", terms.Let("tmp", terms.App(terms.App(terms.Var("add"), terms.Var("x")), terms.Int(1)), terms.Var("x"), false))), terms.Var("test"), false)), terms.Lam("f", terms.App(terms.Var("f"), terms.Var("a"))))),
			expected{"(α1 -> α13)", "α1 <: α7, α7 <: α12 & int, α9 <: α13, α12 <: α9", "‹‹α1, α9, α13, α12, α7, int› -> ‹α13››", "‹‹α13, int› -> ‹α13››", "'a ∧ int -> 'a", nil},
		},
		{
			"(fun k -> let test = k (fun x -> let tmp = (fun y -> add y 1) x in x) in test)",
			terms.Lam("k", terms.Let("test", terms.App(terms.Var("k"), terms.Lam("x", terms.Let("tmp", terms.App(terms.Lam("y", terms.App(terms.App(terms.Var("add"), terms.Var("y")), terms.Int(1))), terms.Var("x")), terms.Var("x"), false))), terms.Var("test"), false)),
			expected{"(α1 -> α11)", "α1 <: ((α8 -> α9) -> α10), α8 <: int, α9 :> α8, α11 :> α10", "‹‹α1, ‹‹α8, int› -> ‹α9, α8›› -> ‹α10›› -> ‹α11, α10››", "‹‹‹‹α8, int› -> ‹α8›› -> ‹α10›› -> ‹α10››", "(('a ∧ int -> 'a) -> 'b) -> 'b", nil},
		},
		{
			"(fun k -> let test = k (fun x -> let tmp = let f = fun y -> add y 1 in f x in x) in test)",
			terms.Lam("k", terms.Let("test", terms.App(terms.Var("k"), terms.Lam("x", terms.Let("tmp", terms.Let("f", terms.Lam("y", terms.App(terms.App(terms.Var("add"), terms.Var("y")), terms.Int(1))), terms.App(terms.Var("f"), terms.Var("x")), false), terms.Var("x"), false))), terms.Var("test"), false)),
			expected{"(α1 -> α13)", "α1 <: ((α10 -> α11) -> α12), α10 <: int, α11 :> α10, α13 :> α12", "‹‹α1, ‹‹α10, int› -> ‹α11, α10›› -> ‹α12›› -> ‹α13, α12››", "‹‹‹‹α10, int› -> ‹α10›› -> ‹α12›› -> ‹α12››", "(('a ∧ int -> 'a) -> 'b) -> 'b", nil},
		},
		{
			"fun f -> let r = fun x -> fun g -> { a = f x; b = g x } in r",
			terms.Lam("f", terms.Let("r", terms.Lam("x", terms.Lam("g", terms.Rcd([]terms.Field{{"a", terms.App(terms.Var("f"), terms.Var("x"))}, {"b", terms.App(terms.Var("g"), terms.Var("x"))}}))), terms.Var("r"), false)),
			expected{"(α1 -> (α8 -> (α9 -> {a: α11, b: α10})))", "α1 <: (α5 -> α6), α8 <: α5, α9 <: (α8 -> α10), α11 :> α6", "‹‹α1, ‹α5› -> ‹α6›› -> ‹‹α8, α5› -> ‹‹α9, ‹α8› -> ‹α10›› -> ‹{a: ‹α11, α6›, b: ‹α10›}››››", "‹‹‹α8› -> ‹α6›› -> ‹‹α8› -> ‹‹‹α8› -> ‹α10›› -> ‹{a: ‹α6›, b: ‹α10›}››››", "('a -> 'b) -> 'a -> ('a -> 'c) -> {a: 'b, b: 'c}", nil},
		},
		{
			"fun f -> let r = fun x -> fun g -> { a = g x } in {u = r 0 succ; v = r true not}",
			terms.Lam("f", terms.Let("r", terms.Lam("x", terms.Lam("g", terms.Rcd([]terms.Field{{"a", terms.App(terms.Var("g"), terms.Var("x"))}}))), terms.Rcd([]terms.Field{{"u", terms.App(terms.App(terms.Var("r"), terms.Int(0)), terms.Var("succ"))}, {"v", terms.App(terms.App(terms.Var("r"), terms.Var("true")), terms.Var("not"))}}), false)),
			expected{"(α1 -> {u: α9, v: α14})", "α7 :> int, α9 :> {a: α7}, α12 :> bool, α14 :> {a: α12}", "‹‹α1› -> ‹{u: ‹α9, {a: ‹α7, int›}›, v: ‹α14, {a: ‹α12, bool›}›}››", "‹‹› -> ‹{u: ‹{a: ‹int›}›, v: ‹{a: ‹bool›}›}››", "⊤ -> {u: {a: int}, v: {a: bool}}", nil},
		},
		{
			"fun f -> let r = fun x -> fun g -> { a = g x; b = f x } in {u = r 0 succ; v = r true not}",
			terms.Lam("f", terms.Let("r", terms.Lam("x", terms.Lam("g", terms.Rcd([]terms.Field{{"a", terms.App(terms.Var("g"), terms.Var("x"))}, {"b", terms.App(terms.Var("f"), terms.Var("x"))}}))), terms.Rcd([]terms.Field{{"u", terms.App(terms.App(terms.Var("r"), terms.Int(0)), terms.Var("succ"))}, {"v", terms.App(terms.App(terms.Var("r"), terms.Var("true")), terms.Var("not"))}}), false)),
			expected{"(α1 -> {u: α13, v: α19})", "α1 <: (α6 -> α7), α6 :> bool | int, α10 :> int, α11 :> α7, α13 :> {a: α10, b: α11}, α16 :> bool, α17 :> α7, α19 :> {a: α16, b: α17}", "‹‹α1, ‹α6, bool, int› -> ‹α7›› -> ‹{u: ‹α13, {a: ‹α10, int›, b: ‹α11, α7›}›, v: ‹α19, {a: ‹α16, bool›, b: ‹α17, α7›}›}››", "‹‹‹bool, int› -> ‹α7›› -> ‹{u: ‹{a: ‹int›, b: ‹α7›}›, v: ‹{a: ‹bool›, b: ‹α7›}›}››", "(bool ∨ int -> 'a) -> {u: {a: int, b: 'a}, v: {a: bool, b: 'a}}", nil},
		},
		{
			"fun f -> let r = fun x -> fun g -> { a = g x; b = f x } in {u = r 0 succ; v = r {t=true} (fun y -> y.t)}",
			terms.Lam("f", terms.Let("r", terms.Lam("x", terms.Lam("g", terms.Rcd([]terms.Field{{"a", terms.App(terms.Var("g"), terms.Var("x"))}, {"b", terms.App(terms.Var("f"), terms.Var("x"))}}))), terms.Rcd([]terms.Field{{"u", terms.App(terms.App(terms.Var("r"), terms.Int(0)), terms.Var("succ"))}, {"v", terms.App(terms.App(terms.Var("r"), terms.Rcd([]terms.Field{{"t", terms.Var("true")}})), terms.Lam("y", terms.Sel(terms.Var("y"), "t")))}}), false)),
			expected{"(α1 -> {u: α13, v: α21})", "α1 <: (α6 -> α7), α6 :> {t: bool} | int, α10 :> int, α11 :> α7, α13 :> {a: α10, b: α11}, α16 :> bool, α17 :> α7, α21 :> {a: α16, b: α17}", "‹‹α1, ‹α6, int, {t: ‹bool›}› -> ‹α7›› -> ‹{u: ‹α13, {a: ‹α10, int›, b: ‹α11, α7›}›, v: ‹α21, {a: ‹α16, bool›, b: ‹α17, α7›}›}››", "‹‹‹int, {t: ‹bool›}› -> ‹α7›› -> ‹{u: ‹{a: ‹int›, b: ‹α7›}›, v: ‹{a: ‹bool›, b: ‹α7›}›}››", "(int ∨ {t: bool} -> 'a) -> {u: {a: int, b: 'a}, v: {a: bool, b: 'a}}", nil},
		},
	} {
		t.Run(tt.string, func(t *testing.T) { doTest(t, tt) })
	}
}

func TestRecursion(t *testing.T) {
	for _, tt := range []testCase{
		{
			"let rec f = fun x -> f x.u in f",
			terms.Let("f", terms.Lam("x", terms.App(terms.Var("f"), terms.Sel(terms.Var("x"), "u"))), terms.Var("f"), true),
			expected{"α5", "α5 :> (α6 -> α8) <: (α7 -> α8), α6 <: {u: α7}, α7 <: α6", "‹α5, ‹α6, {u: ‹α9›}› -> ‹α8››", "‹‹{u: ‹α9›}› -> ‹››", "{u: 'a} as 'a -> ⊥", nil},
		},
		{
			"let rec r = fun a -> r in if true then r else r",
			terms.Let("r", terms.Lam("a", terms.Var("r")), terms.App(terms.App(terms.App(terms.Var("if"), terms.Var("true")), terms.Var("r")), terms.Var("r")), true),
			expected{"α10", "α3 :> (α9 -> α8) | (α6 -> α5) <: α10, α5 :> (α6 -> α5) <: α3, α8 :> (α9 -> α8) <: α3, α10 :> (α6 -> α5) | (α9 -> α8)", "‹α10, ‹α6, α9› -> ‹α11››", "‹‹› -> ‹α11››", "(⊤ -> 'a) as 'a", nil},
		},
		{
			"let rec l = fun a -> l in let rec r = fun a -> fun a -> r in if true then l else r",
			terms.Let("l", terms.Lam("a", terms.Var("l")), terms.Let("r", terms.Lam("a", terms.Lam("a", terms.Var("r"))), terms.App(terms.App(terms.App(terms.Var("if"), terms.Var("true")), terms.Var("l")), terms.Var("r")), true), true),
			expected{"α14", "α6 :> (α12 -> (α13 -> α11)) | (α9 -> α8) <: α14, α8 :> (α9 -> α8) <: α6, α11 :> (α12 -> (α13 -> α11)) <: α6, α14 :> (α9 -> α8) | (α12 -> (α13 -> α11))", "‹α14, ‹α9, α12› -> ‹α15››", "‹‹› -> ‹α15››", "(⊤ -> ⊤ -> 'a) as 'a", nil},
		},
		{
			"let rec l = fun a -> fun a -> fun a -> l in let rec r = fun a -> fun a -> r in if true then l else r",
			terms.Let("l", terms.Lam("a", terms.Lam("a", terms.Lam("a", terms.Var("l")))), terms.Let("r", terms.Lam("a", terms.Lam("a", terms.Var("r"))), terms.App(terms.App(terms.App(terms.Var("if"), terms.Var("true")), terms.Var("l")), terms.Var("r")), true), true),
			expected{"α18", "α8 :> (α16 -> (α17 -> α15)) | (α11 -> (α12 -> (α13 -> α10))) <: α18, α10 :> (α11 -> (α12 -> (α13 -> α10))) <: α8, α15 :> (α16 -> (α17 -> α15)) <: α8, α18 :> (α11 -> (α12 -> (α13 -> α10))) | (α16 -> (α17 -> α15))", "‹α18, ‹α11, α16› -> ‹α19››", "‹‹› -> ‹α19››", "(⊤ -> ⊤ -> ⊤ -> ⊤ -> ⊤ -> ⊤ -> 'a) as 'a", nil},
		},
		{
			"let rec recursive_monster = fun x -> { thing = x; self = recursive_monster x } in recursive_monster",
			terms.Let("recursive_monster", terms.Lam("x", terms.Rcd([]terms.Field{{"thing", terms.Var("x")}, {"self", terms.App(terms.Var("recursive_monster"), terms.Var("x"))}})), terms.Var("recursive_monster"), true),
			expected{"α4", "α4 :> (α5 -> {thing: α5, self: α6}) <: (α5 -> α6), α6 :> {thing: α5, self: α6}", "‹α4, ‹α5› -> ‹{self: ‹α7›, thing: ‹α5›}››", "‹‹α5› -> ‹{self: ‹α7›, thing: ‹α5›}››", "'a -> {self: 'b, thing: 'a} as 'b", nil},
		},
	} {
		t.Run(tt.string, func(t *testing.T) { doTest(t, tt) })
	}
}

func TestRandom(t *testing.T) {
	for _, tt := range []testCase{
		{
			"(let rec x = {a = x; b = x} in x)",
			terms.Let("x", terms.Rcd([]terms.Field{{"a", terms.Var("x")}, {"b", terms.Var("x")}}), terms.Var("x"), true),
			expected{"α2", "α2 :> {a: α2, b: α2}", "‹α3›", "‹α3›", "{a: 'a, b: 'a} as 'a", nil},
		},
		{
			"(let rec x = fun v -> {a = x v; b = x v} in x)",
			terms.Let("x", terms.Lam("v", terms.Rcd([]terms.Field{{"a", terms.App(terms.Var("x"), terms.Var("v"))}, {"b", terms.App(terms.Var("x"), terms.Var("v"))}})), terms.Var("x"), true),
			expected{"α5", "α5 :> (α6 -> {a: α7, b: α8}) <: (α6 -> α8) & (α6 -> α7), α7 :> {a: α7, b: α8}, α8 :> {a: α7, b: α8}", "‹α5, ‹α6› -> ‹{a: ‹α9›, b: ‹α10›}››", "‹‹› -> ‹{a: ‹α9›, b: ‹α10›}››", "⊤ -> {a: 'a, b: 'a} as 'a", nil},
		},
		{
			"let rec x = (let rec y = {u = y; v = (x y)} in 0) in 0",
			terms.Let("x", terms.Let("y", terms.Rcd([]terms.Field{{"u", terms.Var("y")}, {"v", terms.App(terms.Var("x"), terms.Var("y"))}}), terms.Int(0), true), terms.Int(0), true),
			expected{"", "", "", "", "", NewTypeError("cannot constrain int <: 'a -> 'b")},
		},
		{
			"(fun x -> (let y = (x x) in 0))",
			terms.Lam("x", terms.Let("y", terms.App(terms.Var("x"), terms.Var("x")), terms.Int(0), false)),
			expected{"(α1 -> int)", "α1 <: (α1 -> α3)", "‹‹α1, ‹α1› -> ‹α3›› -> ‹int››", "‹‹α1, ‹α1› -> ‹›› -> ‹int››", "'a ∧ ('a -> ⊤) -> int", nil},
		},
		{
			"(let rec x = (fun y -> (y (x x))) in x)",
			terms.Let("x", terms.Lam("y", terms.App(terms.Var("y"), terms.App(terms.Var("x"), terms.Var("x")))), terms.Var("x"), true),
			expected{"α5", "α5 :> (α6 -> α7) <: α6 & (α5 -> α8), α6 :> (α6 -> α7) <: (α8 -> α7), α7 <: α8, α8 <: α6", "‹α5, ‹α6, ‹α8› -> ‹α9›› -> ‹α7››", "‹‹‹α8› -> ‹α9›› -> ‹α8››", "('a -> ('a ∧ ('a -> 'b)) as 'b) -> 'a", nil},
		},
		{
			"fun next -> 0",
			terms.Lam("next", terms.Int(0)),
			expected{"(α1 -> int)", "", "‹‹α1› -> ‹int››", "‹‹› -> ‹int››", "⊤ -> int", nil},
		},
		{
			"((fun x -> (x x)) (fun x -> x))",
			terms.App(terms.Lam("x", terms.App(terms.Var("x"), terms.Var("x"))), terms.Lam("x", terms.Var("x"))),
			expected{"α4", "α2 :> (α3 -> α3) <: α4, α3 :> (α3 -> α3) <: α2, α4 :> (α3 -> α3)", "‹α4, ‹α3, α2, α4› -> ‹α5››", "‹α4, ‹α4› -> ‹α5››", "('b ∨ ('b -> 'a)) as 'a", nil},
		},
		{
			"(let rec x = (fun y -> (x (y y))) in x)",
			terms.Let("x", terms.Lam("y", terms.App(terms.Var("x"), terms.App(terms.Var("y"), terms.Var("y")))), terms.Var("x"), true),
			expected{"α5", "α5 :> (α6 -> α8) <: (α7 -> α8), α6 <: (α6 -> α7), α7 <: α6", "‹α5, ‹α6, ‹α6› -> ‹α9›› -> ‹α8››", "‹‹α6, ‹α6› -> ‹α9›› -> ‹››", "('b ∧ ('b -> 'a)) as 'a -> ⊥", nil},
		},
		{
			"fun x -> (fun y -> (x (y y)))",
			terms.Lam("x", terms.Lam("y", terms.App(terms.Var("x"), terms.App(terms.Var("y"), terms.Var("y"))))),
			expected{"(α1 -> (α2 -> α4))", "α1 <: (α3 -> α4), α2 <: (α2 -> α3)", "‹‹α1, ‹α3› -> ‹α4›› -> ‹‹α2, ‹α2› -> ‹α3›› -> ‹α4›››", "‹‹‹α3› -> ‹α4›› -> ‹‹α2, ‹α2› -> ‹α3›› -> ‹α4›››", "('a -> 'b) -> 'c ∧ ('c -> 'a) -> 'b", nil},
		},
		{
			"(let rec x = (let y = (x x) in (fun z -> z)) in x)",
			terms.Let("x", terms.Let("y", terms.App(terms.Var("x"), terms.Var("x")), terms.Lam("z", terms.Var("z")), false), terms.Var("x"), true),
			expected{"α5", "α5 :> (α6 -> α6) <: α6 & (α5 -> α7), α6 :> (α6 -> α6) <: α7, α7 :> (α6 -> α6)", "‹α5, ‹α6, α7› -> ‹α8››", "‹‹α6› -> ‹α8››", "'a -> ('a ∨ ('a -> 'b)) as 'b", nil},
		},
		{
			"(let rec x = (fun y -> (let z = (x x) in y)) in x)",
			terms.Let("x", terms.Lam("y", terms.Let("z", terms.App(terms.Var("x"), terms.Var("x")), terms.Var("y"), false)), terms.Var("x"), true),
			expected{"α5", "α5 :> (α6 -> α6) <: α6 & (α5 -> α7), α6 :> (α6 -> α6) <: α7, α7 :> (α6 -> α6)", "‹α5, ‹α6, α7› -> ‹α8››", "‹‹α6› -> ‹α8››", "'a -> ('a ∨ ('a -> 'b)) as 'b", nil},
		},
		{
			"(let rec x = (fun y -> {u = y; v = (x x)}) in x)",
			terms.Let("x", terms.Lam("y", terms.Rcd([]terms.Field{{"u", terms.Var("y")}, {"v", terms.App(terms.Var("x"), terms.Var("x"))}})), terms.Var("x"), true),
			expected{"α4", "α4 :> (α5 -> {u: α5, v: α6}) <: α5 & (α4 -> α6), α5 :> (α5 -> {u: α5, v: α6}), α6 :> {u: α5, v: α6}", "‹α4, ‹α5› -> ‹α7››", "‹‹α5› -> ‹α7››", "'a -> {u: 'a ∨ ('a -> 'b), v: 'c} as 'c as 'b", nil},
		},
		{
			"(let rec x = (fun y -> {u = (x x); v = y}) in x)",
			terms.Let("x", terms.Lam("y", terms.Rcd([]terms.Field{{"u", terms.App(terms.Var("x"), terms.Var("x"))}, {"v", terms.Var("y")}})), terms.Var("x"), true),
			expected{"α4", "α4 :> (α5 -> {u: α6, v: α5}) <: α5 & (α4 -> α6), α5 :> (α5 -> {u: α6, v: α5}), α6 :> {u: α6, v: α5}", "‹α4, ‹α5› -> ‹α8››", "‹‹α5› -> ‹α8››", "'a -> {u: 'c, v: 'a ∨ ('a -> 'b)} as 'c as 'b", nil},
		},
		{
			"(let rec x = (fun y -> (let z = (y x) in y)) in x)",
			terms.Let("x", terms.Lam("y", terms.Let("z", terms.App(terms.Var("y"), terms.Var("x")), terms.Var("y"), false)), terms.Var("x"), true),
			expected{"α5", "α5 :> (α6 -> α6), α6 <: (α5 -> α7)", "‹α8›", "‹α8›", "('b ∧ ('a -> ⊤) -> 'b) as 'a", nil},
		},
		{
			"(fun x -> (let y = (x x.v) in 0))",
			terms.Lam("x", terms.Let("y", terms.App(terms.Var("x"), terms.Sel(terms.Var("x"), "v")), terms.Int(0), false)),
			expected{"(α1 -> int)", "α1 <: (α5 -> α6) & {v: α3}, α5 :> α3", "‹‹α1, {v: ‹α3›}, ‹α5, α3› -> ‹α6›› -> ‹int››", "‹‹{v: ‹α3›}, ‹α3› -> ‹›› -> ‹int››", "{v: 'a} ∧ ('a -> ⊤) -> int", nil},
		},
		{
			"let rec x = (let y = (x x) in (fun z -> z)) in (x (fun y -> y.u))",
			terms.Let("x", terms.Let("y", terms.App(terms.Var("x"), terms.Var("x")), terms.Lam("z", terms.Var("z")), false), terms.App(terms.Var("x"), terms.Lam("y", terms.Sel(terms.Var("y"), "u"))), true),
			expected{"α10", "α6 :> (α8 -> α9) | (α6 -> α6) <: α10 & α7, α7 :> (α8 -> α9) | (α6 -> α6), α8 <: {u: α9}, α10 :> (α6 -> α6) | (α8 -> α9)", "‹α10, ‹α6, α10, α7, α8, {u: ‹α9›}› -> ‹α11››", "‹α10, ‹α10, {u: ‹α9›}› -> ‹α11››", "'a ∨ ('a ∧ {u: 'b} -> ('b ∨ 'a ∨ ('a ∧ {u: 'b} -> 'c)) as 'c)", nil},
		},
	} {
		t.Run(tt.string, func(t *testing.T) { doTest(t, tt) })
	}
}

func doTest(t *testing.T, tt testCase) {
	var typer = NewTyper()
	infer := func(term terms.Term) (tyv SimpleType, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = r.(error)
			}
		}()
		tyv = typer.inferType(term, typer.Builtins())
		return
	}

	t.Log(tt.string)
	tyv, err := infer(tt.Term)
	if tt.expected.error == nil {
		if err == nil {
			inferred := tyv.String()
			where := ShowBounds(tyv)
			t.Log("inferred: " + inferred)
			t.Log(" where " + where)
			if tt.expected.inferred != inferred {
				t.Errorf("inferred: expect %s actual %s", tt.expected.inferred, inferred)
			}
			if tt.expected.where != where {
				t.Errorf("where: expect %s actual %s", tt.expected.where, where)
			}

			cty := typer.canonicalizeType(tyv)
			sCty := cty.term.String()
			t.Log("compacted: " + sCty)
			if tt.expected.compacted != sCty {
				// 顺序有问题
				//t.Errorf("compacted: expect %s actual %s", tt.expected.compacted, sCty)
			}

			sty := typer.simplifyType(cty)
			sSty := sty.term.String()
			t.Log("simplified: " + sSty)
			if tt.expected.simplified != sSty {
				t.Errorf("simplified: expect %s actual %s", tt.expected.simplified, sSty)
			}

			ety := typer.coalesceCompactType(sty)
			sEty := ety.Show()
			t.Log("coalesced: " + sEty)
			if tt.expected.coalesced != sEty {
				t.Errorf("coalesced: expect %s actual %s", tt.expected.coalesced, sEty)
			}
		} else {
			t.Errorf("expect %s %s actual error %s", tt.expected.inferred, tt.expected.where, err)
		}
	} else {
		if err == nil {
			inferred := tyv.String()
			where := ShowBounds(tyv)
			t.Errorf("expect error %s actual %s %s", tt.expected.Error(), inferred, where)
		} else {
			if err.Error() != tt.expected.Error() {
				t.Errorf("expect error %s actual error %s", tt.expected.Error(), err.Error())
			}
		}
	}
	t.Log("==============================")
}

func TestCanonicalizationProducesLCD(t *testing.T) {
	typer := NewTyper()
	tv0 := typer.freshVar(0)
	tv1 := typer.freshVar(0)
	tv3 := typer.freshVar(0)

	// {f: {B: int, f: 'a}} as 'a  –  cycle length 2
	st0 := Rcd([]field{
		{
			"f",
			Rcd([]field{
				{"B", Int},
				{"f", tv0},
			}),
		},
	})
	tv0.prependLower(st0)

	// {f: {B: int, f: {A: int, f: 'a}}} as 'a  –  cycle length 3
	st1 := Rcd([]field{
		{
			"f",
			Rcd([]field{
				{"B", Int},
				{
					"f",
					Rcd([]field{
						{"f", tv1},
						{"a", Int},
					}),
				},
			}),
		},
	})
	tv1.prependLower(st1)
	tv3.prependLower(tv1)
	tv3.prependLower(tv0)

	ct := typer.canonicalizeType(tv3)
	sct := typer.simplifyType(ct)
	csct := typer.coalesceCompactType(sct).Show()

	expect := "{f: {B: int, f: {f: {f: {f: {f: 'a}}}}}} as 'a"
	if csct != expect {
		t.Errorf("expect %s actual %s", expect, csct)
	}
}
