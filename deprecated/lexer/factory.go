package lexer

import (
	"github.com/goghcrow/simple-sub/deprecated/oper"
	"github.com/goghcrow/simple-sub/deprecated/token"
)

var keywords = []token.Type{
	token.IF,
	token.THEN,
	token.ELSE,
	token.LET,
	token.REC,
	token.IN,
	token.FUN,
}

var buildInOPs = []token.Type{
	token.ARROW, // ->
	token.DOT,   // .
}

func newLexicon(ops []oper.Operator) lexicon {
	l := lexicon{}

	//lexer.go skipSpace 直接处理掉, 这里忽略
	//l.addRule(regOpt(token.WHITESPACE, "[ \r\t]+", false)) // 不能使用\s+, 要单独处理换行
	l.addRule(regOpt(token.BLOCK_COMMENT, "/\\*[\\s\\S]*?\\*+/", false))
	l.addRule(regOpt(token.LINE_COMMENT, "//.*", false))

	l.addRule(str(token.NEWLINE)) // \n
	l.addRule(str(token.COLON))   // :
	l.addRule(str(token.COMMA))   // ,

	l.addRule(str(token.ASSIGN)) // =

	l.addRule(str(token.LEFT_PAREN))    // (
	l.addRule(str(token.RIGHT_PAREN))   // )
	l.addRule(str(token.LEFT_BRACKET))  // [
	l.addRule(str(token.RIGHT_BRACKET)) // ]
	l.addRule(str(token.LEFT_BRACE))    // {
	l.addRule(str(token.RIGHT_BRACE))   // }

	for _, kw := range keywords {
		l.addRule(keyword(kw))
	}

	// 内置的操作符优先级高于自定义操作符
	for _, op := range buildInOPs {
		l.addRule(primOper(op))
	}

	// 自定义操作符
	for _, op := range ops {
		l.addOper(op.Type)
	}

	// TODO
	// 这里为了支持自定义操作符语法
	// 注意: 这里不会匹配到 ASSIGN(=), ARROW(=>), DOT(.)，因为 KEYWORD 已经注册过
	l.addRule(reg(token.OPER, oper.Pattern))

	//l.addRule(str(token.NULL))  // null
	l.addRule(str(token.TRUE))  // true
	l.addRule(str(token.FALSE)) // false

	// [+-]? 被处理成一元操作符, 没有负数字面量
	// 移除数字前的 [+-]?, lex 没有使用最长路径来匹配, +- 被优先匹配成操作符了
	l.addRule(reg(token.FLOAT, "(?:0|[1-9][0-9]*)(?:[.][0-9]+)+(?:[eE][-+]?[0-9]+)?")) // float
	l.addRule(reg(token.FLOAT, "(?:0|[1-9][0-9]*)(?:[.][0-9]+)?(?:[eE][-+]?[0-9]+)+")) // float
	l.addRule(reg(token.INT, "0b(?:0|1[0-1]*)"))                                       // int
	l.addRule(reg(token.INT, "0x(?:0|[1-9a-fA-F][0-9a-fA-F]*)"))                       // int
	l.addRule(reg(token.INT, "0o(?:0|[1-7][0-7]*)"))                                   // int
	l.addRule(reg(token.INT, "(?:0|[1-9][0-9]*)"))                                     // int

	l.addRule(reg(token.STR, "\"(?:[^\"\\\\]*|\\\\[\"\\\\trnbf\\/]|\\\\u[0-9a-fA-F]{4})*\""))
	l.addRule(reg(token.STR, "`[^`]*`")) // raw string

	l.addRule(reg(token.NAME, "[a-zA-Z\\p{L}_][a-zA-Z0-9\\p{L}_]*")) // 支持 unicode, 不能以数字开头

	return l
}
