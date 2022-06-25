package trycatch

import (
	"github.com/pkg/errors"
)

func ToError(err *error) {
	if r := recover(); r != nil {
		switch x := r.(type) {
		case error:
			*err = errors.New(x.Error() + "\n")
		default:
			*err = errors.Errorf("unknown recovered type; val=%+v\n", x)
		}
	}
}
