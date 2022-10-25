package terms

import (
	"github.com/goghcrow/simple-sub/front/oper"
	"github.com/goghcrow/simple-sub/util"
)

func Bool(val bool) *LiteralBool                 { return &LiteralBool{Val: val} }
func Int(val int64) *LiteralInt                  { return &LiteralInt{Val: val} }
func Float(val float64) *LiteralFloat            { return &LiteralFloat{Val: val} }
func Str(val string) *LiteralString              { return &LiteralString{Val: val} }
func Tup(xs ...Term) *Tuple                      { return &Tuple{xs} }
func Var(name string) *Variable                  { return &Variable{Name: name} }
func Lam(name string, rhs Term) *Lambda          { return &Lambda{Name: name, Rhs: rhs} }
func App(lhs Term, rhs Term) *Application        { return &Application{Lhs: lhs, Rhs: rhs} }
func Rcd(xs []Field) *Record                     { return &Record{Fields: xs} }
func Sel(recv Term, fieldName string) *Selection { return &Selection{Recv: recv, FieldName: fieldName} }
func Let(name string, rhs Term, body Term, rec bool) *LetDefine {
	return &LetDefine{*Def(name, rhs, rec), body}
}

func Pgrm(defs []*Define) *Program                { return &Program{Defs: defs} }
func Def(name string, rhs Term, rec bool) *Define { return &Define{Name: name, Rhs: rhs, Rec: rec} }

func Grp(term Term) *Group                                   { return &Group{term} }
func Iff(cond, then, els Term) *If                           { return &If{cond, then, els} }
func Un(name string, term Term, prefix bool) *Unary          { return &Unary{name, term, prefix} }
func Bin(name string, bp oper.Fixity, lhs, rhs Term) *Binary { return &Binary{name, bp, lhs, rhs} }

func LamN(xs []string, rhs Term) *Lambda {
	argc := len(xs)
	util.Assert(argc > 0, "at least 1 param")
	lam := Lam(xs[argc-1], rhs)
	for i := argc - 2; i >= 0; i-- {
		lam = Lam(xs[i], lam)
	}
	return lam
}
func AppN(lhs Term, xs ...Term) *Application {
	argc := len(xs)
	util.Assert(argc > 0, "at least 1 arg")
	app := App(lhs, xs[0])
	for i := 1; i < argc; i++ {
		app = App(app, xs[i])
	}
	return app
}
