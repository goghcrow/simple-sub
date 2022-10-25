package token

// Type 因为要支持动态添加操作符, 所以 Type 没有定义成 int 枚举
// 这里需要自己保证 type 值不重复
type Type string

// '前缀的为 psuido
//
//goland:noinspection GoSnakeCaseUsage
const (
	// 关键字 KEYWORD

	IF    = "if"
	THEN  = "then"
	ELSE  = "else"
	LET   = "let"
	REC   = "rec"
	IN    = "in"
	FUN   = "fun"
	TRUE  = "true"
	FALSE = "false"

	//NULL   = "null"
	//PREFIX = "prefix"
	//INFIXN = "infixn"
	//INFIXL = "infixl"
	//INFIXR = "infixr"
	//RECORD = "record"
	//TYPE   = "type"

	// 标识符 IDENTIFIER & 字面量 LITERAL

	NAME  = "'name"
	FLOAT = "'float"
	INT   = "'int"
	STR   = "'str"

	// 算子 OPERATOR

	OPER   = "'oper"
	ASSIGN = "="
	ARROW  = "->" // =>
	DOT    = "."

	// 分隔符 SEPARATOR

	COLON   = ":"
	COMMA   = ","
	NEWLINE = "\n" // 忽略空白, 但换行要特殊处理
	//SEMICOLON = ";"

	LEFT_PAREN    = "("
	RIGHT_PAREN   = ")"
	LEFT_BRACKET  = "["
	RIGHT_BRACKET = "]"
	LEFT_BRACE    = "{"
	RIGHT_BRACE   = "}"

	// 空白和注释 SPACE COMMENT

	WHITESPACE    = ""
	BLOCK_COMMENT = "/**/"
	LINE_COMMENT  = "//"
	EOF           = "-EOF-"
)
