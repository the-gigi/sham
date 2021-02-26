package main

import (
	"github.com/the-gigi/sham"
)

type mockFoo struct {
	sham.CannedResponseMock
}

func (f *mockFoo) Bar() {
	_, _ = f.VerifyCallNoArgs("Bar", 0)
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

func newMockFoo(calls []*sham.Call) (result *mockFoo, err error) {
	result = &mockFoo{}
	result.ExpectedCalls = calls
	err = result.Invariant()
	return
}
