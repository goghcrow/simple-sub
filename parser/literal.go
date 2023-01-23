package parser

import (
	"errors"

	"github.com/goghcrow/lexer"
	"github.com/goghcrow/simple-sub/terms"
	"github.com/goghcrow/simple-sub/util"
	"strconv"
	"strings"
)

func parseInt(t *lexer.Token) terms.Term {
	n, err := parseInt0(t.Lexeme)
	util.Assert(err == nil, "invalid int literal %s", t.Lexeme)
	return terms.Int(n)
}

func parseFloat(t *lexer.Token) terms.Term {
	n, err := strconv.ParseFloat(t.Lexeme, 64)
	util.Assert(err == nil, "invalid float literal %s", t.Lexeme)
	return terms.Float(n)
}

func parseString(t *lexer.Token) terms.Term {
	s, err := strconv.Unquote(t.Lexeme)
	util.Assert(err == nil, "invalid string literal: %s", t.Lexeme)
	return terms.Str(s)
}

func parseInt0(s string) (int64, error) {
	n, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return n, nil
	}
	if strings.HasPrefix(s, "0x") {
		n, err := strconv.ParseInt(s[2:], 16, 64)
		if err == nil {
			return n, nil
		}
	}
	if strings.HasPrefix(s, "0b") {
		n, err := strconv.ParseInt(s[2:], 2, 64)
		if err == nil {
			return n, nil
		}
	}
	if strings.HasPrefix(s, "0o") {
		n, err := strconv.ParseInt(s[2:], 8, 64)
		if err == nil {
			return n, nil
		}
	}
	return 0, errors.New("invalid int: " + s)
}
