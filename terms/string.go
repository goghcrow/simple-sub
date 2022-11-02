package terms

import (
	"fmt"
)

func (l *LiteralInt) String() string    { return fmt.Sprintf("Int(%d)", l.Val) }
func (l *LiteralBool) String() string   { return fmt.Sprintf("Bool(%t)", l.Val) }
func (l *LiteralFloat) String() string  { return fmt.Sprintf("Float(%f)", l.Val) }
func (l *LiteralString) String() string { return fmt.Sprintf("Str(%q)", l.Val) }
func (v *Variable) String() string      { return fmt.Sprintf("Var(%s)", v.Name) }
func (l *Lambda) String() string        { return fmt.Sprintf("Fun(%s, %s)", l.Name, l.Rhs) }
func (a *Application) String() string   { return fmt.Sprintf("App(%s %s)", a.Lhs, a.Rhs) }
func (t *Tuple) String() string         { return fmt.Sprintf("Tuple(%s)", t.Elms) }
func (f *Field) String() string         { return fmt.Sprintf("Field(%s, %s)", f.Name, f.Term) }
func (r *Record) String() string        { return fmt.Sprintf("Rcd(%s)", r.Fields) }
func (s *Selection) String() string     { return fmt.Sprintf("Sel(%s, %s)", s.Recv.String(), s.FieldName) }
func (l *LetDefine) String() string {
	if l.Rec {
		return fmt.Sprintf("LetRec(%s, %s, %s)", l.Name, l.Rhs, l.Body)
	} else {
		return fmt.Sprintf("Let(%s, %s, %s)", l.Name, l.Rhs, l.Body)
	}
}

func (g *Group) String() string { return fmt.Sprintf("Group(%s)", g.Term) }
func (i *If) String() string    { return fmt.Sprintf("If(%s, %s, %s)", i.Cond, i.Then, i.Else) }
func (u *Unary) String() string { return fmt.Sprintf("Unary(%s, %s, %t)", u.Name, u.Rhs, u.Prefix) }
func (b *Binary) String() string {
	return fmt.Sprintf("Binary(%s, %s, %s, %s)", b.Name, b.Fixity, b.Lhs, b.Rhs)
}

func (p Program) String() string { return fmt.Sprintf("Program(%s)", p.Defs) }
func (d Define) String() string {
	if d.Rec {
		return fmt.Sprintf("Let(%s, %s)", d.Name, d.Rhs)
	} else {
		return fmt.Sprintf("LetRec(%s, %s)", d.Name, d.Rhs)
	}
}
