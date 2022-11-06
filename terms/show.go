package terms

import (
	"fmt"
	"github.com/goghcrow/simple-sub/util"
	"strconv"
	"strings"
)

func ShowPgrm(p *Program) string {
	var b strings.Builder
	for _, def := range p.Defs {
		b.WriteString(ShowDef(def))
		b.WriteString("\n")
	}
	return b.String()
}

func ShowDef(def *Declaration) string {
	rhs := showTerm(def.Rhs, 0)
	if def.Rec {
		return fmt.Sprintf("let rec %s = %s", def.Name, rhs)
	} else {
		return fmt.Sprintf("let %s = %s", def.Name, rhs)
	}
}

func ShowTerm(term Term) string {
	return showTerm(term, 0)
}

func showTerm(term Term, outerPrec int) string {
	switch t := term.(type) {
	case *LiteralBool:
		return strconv.FormatBool(t.Val)
	case *LiteralInt:
		return strconv.FormatInt(t.Val, 10)
	case *LiteralFloat:
		return strconv.FormatFloat(t.Val, 'g', -1, 64)
	case *LiteralString:
		return fmt.Sprintf("%q", t.Val)
	case *Variable:
		return t.Name
	case *Lambda:
		rhs := showTerm(t.Rhs, 10)
		fun := fmt.Sprintf("fun %s -> %s", t.Name, rhs)
		return parensIf(fun, outerPrec > 10)
	case *Application:
		lhs := showTerm(t.Lhs, 20)
		rhs := showTerm(t.Rhs, 20)
		app := fmt.Sprintf("%s %s", lhs, rhs)
		return parensIf(app, outerPrec > 20)
	case *Tuple:
		xs := make([]string, len(t.Elms))
		for i, el := range t.Elms {
			xs[i] = showTerm(el, 0)
		}
		return util.JoinStr(xs, ", ", "(", ")")
	case *List:
		xs := make([]string, len(t.Elms))
		for i, el := range t.Elms {
			xs[i] = showTerm(el, 0)
		}
		return util.JoinStr(xs, ", ", "[", "]")
	case *Record:
		xs := make([]string, len(t.Fields))
		for i, fd := range t.Fields {
			xs[i] = fmt.Sprintf("%s: %s", fd.Name, showTerm(fd.Term, 0))
		}
		return util.JoinStr(xs, ", ", "{", "}")
	case *Selection:
		return fmt.Sprintf("%s.%s", showTerm(t.Recv, 30), t.FieldName)
	case *LetDefine:
		body := showTerm(t.Body, 0)
		rhs := showTerm(t.Rhs, 0)
		if t.Rec {
			return fmt.Sprintf("let rec %s = %s in %s", t.Name, rhs, body)
		} else {
			return fmt.Sprintf("let %s = %s in %s", t.Name, rhs, body)
		}
	default:
		panic("unreached")
	}
}

func parensIf(str string, cnd bool) string {
	if cnd {
		return "(" + str + ")"
	}
	return str
}
