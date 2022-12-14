package terms

import (
	"fmt"
	"github.com/goghcrow/simple-sub/deprecated/oper"
)

// Syntax

// Term Language
// đĄ ::= đĽ | đđĽ. đĄ | đĄ đĄ | { đ0 = đĄ ; ... ; đđ = đĄ } | đĄ.đ | let rec đĽ = đĄ in đĄ

// Type Language
// đ ::= primitive | đ â đ | { đ0 : đ ; ... ; đđ : đ } | đź | â¤ | âĽ | đ â đ | đ â đ | đđź. đ

// Type System
// Typing Rules: T-Lit, T-Var, T-Abs, T-App, T-Rcd, T-Proj, T-Sub, T-Let
// T-Sub, which takes a term from a subtype to a supertype implicitly
// T-Let, which types đĽ in its recursive right-hand side in a monomorphic way, but types đĽ in its body polymorphically,
// T-Var, which instantiates polymorphic types using the substitution syntax [đ0/đź0]đ.

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
	List struct { // as in: [1, 2]
		Elms []Term
	}
	Variable struct { // as in: đĽ
		Name string
	}
	Lambda struct { // as in: đđĽ. đĄ
		Name string
		Rhs  Term
	}
	Application struct { // as in: đ  đĄ
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
	Selection struct { // as in: đĄ.a
		Recv      Term
		FieldName string
	}
	LetDefine struct { // let đĽ = 42 in đĽ
		Declaration
		Body Term
	}
)

// parser éśćŽľ term äźč˘Ť desugar ĺ¤çć
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

// Declaration : Top Level Let Binding
type Declaration struct {
	Rec  bool
	Name string
	Rhs  Term
}

type Program struct {
	Defs []*Declaration
}

func (_ *LiteralInt) _termNop()    {}
func (_ *LiteralBool) _termNop()   {}
func (_ *LiteralFloat) _termNop()  {}
func (_ *LiteralString) _termNop() {}
func (_ *Tuple) _termNop()         {}
func (_ *List) _termNop()          {}
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

func (_ *Program) _termNop()     {}
func (_ *Declaration) _termNop() {}
