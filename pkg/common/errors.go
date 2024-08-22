package common

import "errors"

func AsError[T error](err error) (T, bool) {
	var target T
	return target, errors.As(err, &target)
}
