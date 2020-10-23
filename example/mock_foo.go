package main

import (
	"github.com/the-gigi/sham"
)

type mockFoo struct {
	sham.CannedResponseMock
}

func (f *mockFoo) Bar() {
	_, err := f.VerifyCallNoArgs("Bar", 0)
	if err != nil {
		return
	}
}

func (f *mockFoo) Baz(s string) (result int, err error) {
	call, err := f.VerifyCall("Baz", 2, s)
	if err != nil {
		return
	}

	result = call.Result[0].(int)
	err = sham.ToError(call.Result[1])
	return
}
