package oper

// Fixity Associativity
type Fixity int

//goland:noinspection GoSnakeCaseUsage
const (
	NA Fixity = iota
	PREFIX
	INFIX_N
	INFIX_L
	INFIX_R
	POSTFIX
)

func (i Fixity) String() string {
	return [...]string{"NA", "PREFIX", "INFIX_N", "INFIX_L", "INFIX_R", "POSTFIX"}[i]
}
