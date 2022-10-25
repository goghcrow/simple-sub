package terms

import (
	"fmt"
	"github.com/goghcrow/simple-sub/front/oper"
)

// Syntax

// Term Language
// 𝑡 ::= 𝑥 | 𝜆𝑥. 𝑡 | 𝑡 𝑡 | { 𝑙0 = 𝑡 ; ... ; 𝑙𝑛 = 𝑡 } | 𝑡.𝑙 | let rec 𝑥 = 𝑡 in 𝑡

// Type Language
// 𝜏 ::= primitive | 𝜏 → 𝜏 | { 𝑙0 : 𝜏 ; ... ; 𝑙𝑛 : 𝜏 } | 𝛼 | ⊤ | ⊥ | 𝜏 ⊔ 𝜏 | 𝜏 ⊓ 𝜏 | 𝜇𝛼. 𝜏

// Type System
// Typing Rules: T-Lit, T-Var, T-Abs, T-App, T-Rcd, T-Proj, T-Sub, T-Let
// T-Sub, which takes a term from a subtype to a supertype implicitly
// T-Let, which types 𝑥 in its recursive right-hand side in a monomorphic way, but types 𝑥 in its body polymorphically,
// T-Var, which instantiates polymorphic types using the substitution syntax [𝜏0/𝛼0]𝜏.

type Term interface {
	fmt.Stringer
	_termNop()
}

type (
	LiteralInt struct { // as in: 42
		Val int64
	}
	LiteralFloat struct { // as in: 3.14
		Val float64
	}
	LiteralString struct { // as in: "Hello"
		Val string
	}
	LiteralBool struct { // as in: true,false
		Val bool
	}
	Tuple struct { // as in: (1, 2)
		Elms []Term
	}
	Variable struct { // as in: 𝑥
		Name string
	}
	Lambda struct { // as in: 𝜆𝑥. 𝑡
		Name string
		Rhs  Term
	}
	Application struct { // as in: 𝑠 𝑡
		Lhs Term
		Rhs Term
	}
	Field struct {
		Name string
		Term Term
	}
	Record struct { // as in: { a : 0; b : true; ... }
		Fields []Field
	}
	Selection struct { // as in: 𝑡.a
		Recv      Term
		FieldName string
	}
	LetDefine struct { // let 𝑥 = 42 in 𝑥
		Define
		Body Term
	}
)

// parser 阶段 term 会被 desugar 处理掉
type (
	Unary struct {
		Name   string
		Rhs    Term
		Prefix bool
	}
	Binary struct {
		Name string
		oper.Fixity
		Lhs Term
		Rhs Term
	}
	Group struct {
		Term
	}
	If struct {
		Cond Term
		Then Term
		Else Term
	}
)

// Define : Global Let Binding
type Define struct {
	Rec  bool
	Name string
	Rhs  Term
}

type Program struct {
	Defs []*Define
}

func (_ *LiteralInt) _termNop()    {}
func (_ *LiteralBool) _termNop()   {}
func (_ *LiteralFloat) _termNop()  {}
func (_ *LiteralString) _termNop() {}
func (_ *Tuple) _termNop()         {}
func (_ *Variable) _termNop()      {}
func (_ *Lambda) _termNop()        {}
func (_ *Application) _termNop()   {}
func (_ *Record) _termNop()        {}
func (_ *Selection) _termNop()     {}
func (_ *LetDefine) _termNop()     {}

func (_ *If) _termNop()     {}
func (_ *Group) _termNop()  {}
func (_ *Unary) _termNop()  {}
func (_ *Binary) _termNop() {}

func (_ *Program) _termNop() {}
func (_ *Define) _termNop()  {}
