package error_handlers

import (
	"fmt"
	"log"
)

// Unwrap takes a value and error pair and returns just the value, ignoring
// any error.
//
// It is unsafe and intended for use in initializers and testing
// code where the error case is known to be impossible, or when the
// error can be safely ignored.
func Unwrap[T any](v T, err error) T {
	return v
}

// UnwrapPanic takes a value and error pair and returns just the value.
//
// If the error is not nil, it panics with the error. It is unsafe and
// intended for use in initializers and testing code where the
// error case is known to be impossible.
func UnwrapPanic[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// UnwrapWith returns a function that unwraps a value and error pair,
// panicking with the given error message if the error is non-nil.
//
// The returned function takes a value and error pair and returns just
// the value. If the error is not nil, it panics with the error
// message prepended. It is unsafe and intended for use in initializers
// and testing code where the error case is known to be impossible.
func UnwrapWith[T any](errmsg string) func(v T, err error) T {
	return func(v T, err error) T {
		if err != nil {
			log.Panic(fmt.Errorf("%s: %w", errmsg, err))
		}
		return v
	}
}

// UnwrapStore takes a pointer to an error and returns a function that
// unwraps a value and error pair, storing the error in the pointer.
func UnwrapStore[T any, Ep *error](errPtr Ep) func(v T, err error) T {
	return func(v T, err error) T {
		*errPtr = err
		return v
	}
}
