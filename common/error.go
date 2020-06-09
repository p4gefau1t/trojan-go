package common

import (
	"fmt"
)

type Error struct {
	info string
}

func (e *Error) Error() string {
	return e.info
}

func (e *Error) Base(err error) *Error {
	if err != nil {
		e.info += " | " + err.Error()
	}
	return e
}

func NewError(info string) *Error {
	return &Error{
		info: info,
	}
}

func Must(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func Must2(_ interface{}, err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
