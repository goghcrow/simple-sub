package trans

import (
	"github.com/goghcrow/simple-sub/terms"
)

type Translate func(expr terms.Term) terms.Term
