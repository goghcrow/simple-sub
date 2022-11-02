package lexer

import (
	"github.com/goghcrow/simple-sub/deprecated/oper"
	"github.com/goghcrow/simple-sub/deprecated/token"
)

func NewLexer(ops []oper.Operator) *lexer {
	return &lexer{
		lexicon: newLexicon(oper.Sort(ops)),
	}
}

// Lex 表达式通常都很短, 这里没有要做成语法制导按需lex, e.g. chan *token.Token
func (l *lexer) Lex(input string) []*token.Token {
	l.input = input
	l.idx = 0
	var toks []*token.Token
	for {
		t := l.next()
		if t == EOF {
			break
		}
		toks = append(toks, t)
	}
	return toks
}

var EOF = &token.Token{Type: token.EOF}

type lexer struct {
	lexicon
	input string
	idx   int
}

func (l *lexer) next() *token.Token {
	for {
		tok, keep := l.next0()
		if keep {
			return tok
		}
	}
}

func (l *lexer) next0() (*token.Token, bool) {
	l.skipSpace()
	if l.idx >= len(l.input) {
		return EOF, true
	}

	sub := l.input[l.idx:]
	for _, r := range l.lexicon.rules {
		offset := r.match(sub)
		if offset >= 0 {
			matched := l.input[l.idx : l.idx+offset]
			l.idx += offset
			return &token.Token{Type: r.Type, Lexeme: matched}, r.keep
		}
	}
	panic("syntax error: " + sub)
}

func (l *lexer) skipSpace() {
	for l.idx < len(l.input) {
		if !isSpace(l.input[l.idx]) {
			break
		}
		l.idx++
	}
}
func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' // || c == '\n' // 特殊处理 \n
}
