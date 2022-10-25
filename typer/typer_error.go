package typer

import "fmt"

type TypeError struct {
	Msg string
}

func NewTypeError(format string, a ...any) TypeError {
	return TypeError{Msg: fmt.Sprintf(format, a...)}
}

func TypeErrorOf(v interface{}) *TypeError {
	err, ok := v.(TypeError)
	if ok {
		return &err
	}
	errPtr, ok := v.(*TypeError)
	if ok {
		return errPtr
	}
	err = NewTypeError("%v", v)
	return &err
}

func (t TypeError) Error() string {
	return t.Msg
}
