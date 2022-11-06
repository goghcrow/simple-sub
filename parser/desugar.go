package parser

import (
	"github.com/goghcrow/simple-sub/deprecated/token"
	. "github.com/goghcrow/simple-sub/terms"
)

type Translate func(expr Term) Term

func Desugar(term Term) Term {
	switch t := term.(type) {
	case *LiteralInt:
		return t
	case *LiteralFloat:
		return t
	case *LiteralString:
		return t
	case *LiteralBool:
		return t
	case *Variable:
		return t
	case *Tuple:
		xs := make([]Term, len(t.Elms))
		for i, el := range t.Elms {
			xs[i] = Desugar(el)
		}
		return Tup(xs...)
	case *List:
		xs := make([]Term, len(t.Elms))
		for i, el := range t.Elms {
			xs[i] = Desugar(el)
		}
		return Lst(xs...)
	case *Record:
		xs := make([]Field, len(t.Fields))
		for i, fd := range t.Fields {
			xs[i] = Field{Name: fd.Name, Term: Desugar(fd.Term)}
		}
		return Rcd(xs)
	case *Lambda:
		return Lam(t.Name, Desugar(t.Rhs))
	case *Application:
		return App(Desugar(t.Lhs), Desugar(t.Rhs))
	case *Selection:
		return Sel(Desugar(t.Recv), t.FieldName)
	case *LetDefine:
		return Let(t.Name, Desugar(t.Rhs), Desugar(t.Body), t.Rec)
	case *Unary:
		return App(Var(t.Name), Desugar(t.Rhs))
	case *Binary:
		return App(App(Var(t.Name), Desugar(t.Lhs)), Desugar(t.Rhs))
	case *Group:
		return Desugar(t.Term)
	case *If:
		return AppN(Var(token.IF), Desugar(t.Cond), Desugar(t.Then), Desugar(t.Else))
	case *Declaration:
		return Decl(t.Name, Desugar(t.Rhs), t.Rec)
	case *Program:
		xs := make([]*Declaration, len(t.Defs))
		for i, def := range t.Defs {
			xs[i] = Desugar(def).(*Declaration)
		}
		return Pgrm(xs)
	default:
		panic("unreached")
		return nil
	}
}
