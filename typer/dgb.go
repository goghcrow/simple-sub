package typer

import (
	"fmt"
	"os"
)

const DBG = false

func log(format string, a ...interface{}) {
	if DBG {
		_, _ = fmt.Fprintf(os.Stderr, format+"\n", a...)
	}
}
