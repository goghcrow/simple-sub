package lexer

import (
	"github.com/goghcrow/simple-sub/deprecated/oper"
	"github.com/goghcrow/simple-sub/deprecated/token"
	"regexp"
	"strings"
)

// match 返回匹配 endIdx, 失败返回-1
type match func(string) int

type rule struct {
	token.Type
	match
	keep bool
}

func str(t token.Type) rule { return strOpt(t, true) }
func strOpt(t token.Type, keep bool) rule {
	return rule{t, func(s string) int {
		if strings.HasPrefix(s, string(t)) {
			return len(string(t))
		} else {
			return -1
		}
	}, keep}
}

var keywordPostfix = regexp.MustCompile(`^[a-zA-Z\d\p{L}_']+`)

func keyword(t token.Type) rule {
	// golang 正则不支持 negative look ahead and behind
	return rule{t, func(s string) int {
		k := string(t)
		completedWord := strings.HasPrefix(s, k) &&
			!keywordPostfix.MatchString(s[len(k):])
		if completedWord {
			return len(k)
		} else {
			return -1
		}
	}, true}
}

func reg(t token.Type, pattern string) rule { return regOpt(t, pattern, true) }
func regOpt(t token.Type, pattern string, keep bool) rule {
	startWith := regexp.MustCompile("^" + pattern)
	return rule{t, func(s string) int {
		found := startWith.FindString(s)
		if found == "" {
			return -1
		} else {
			return len(found)
		}
	}, keep}
}

// primOper . ? 内置操作符的优先级高于自定义操作符, 且不是匹配最长, 需要特殊处理
// e.g 比如自定义操作符 .^. 不能匹配成 [`.`, `^.`]
func primOper(t token.Type) rule {
	sz := len(string(t))
	return rule{t, func(s string) int {
		if !strings.HasPrefix(s, string(t)) {
			return -1
		}
		completedOper := len(s) == sz || !oper.HasPrefix(s[sz:])
		if completedOper {
			return sz
		} else {
			return -1
		}
	}, true}
}
