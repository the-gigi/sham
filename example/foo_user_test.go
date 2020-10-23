package main

import (
	"testing"

	"github.com/pkg/errors"

	"github.com/the-gigi/sham"
)

func TestSuccessfulFooBaz(t *testing.T) {
	// Prepare the expected calls
	//
	// two calls are expected:
	// 1. Bar() with no arguments and no return values
	// 2. Baz() with single string argument "two" and return values of 2 and nil
	expectedCalls := []*sham.Call{
		sham.NewCall("Bar"),
		sham.NewCall("Baz", "two").Return(2, nil),
	}

	// Create the mock foo with the expected calls
	m, err := newMockFoo(expectedCalls)
	if err != nil {
		t.Fail()
	}

	// Call the code under test with the mock foo and the expected argument
	result, err := useFoo(m, "two")

	// Verify the result
	if result != 7 {
		t.Fail()
	}

	if err != nil {
		t.Fail()
	}

	// Verify the correct calls were made to the mocj object
	if !m.IsValid() {
		t.Fail()
	}
}

func TestFailedFooBaz(t *testing.T) {
	errorMessage := "xxxxx is not a digit"
	// Prepare the expected calls
	//
	// two calls are expected:
	// 1. Bar() with no arguments and no return values
	// 2. Baz() with single string argument "two" and return values of 2 and nil
	expectedCalls := []*sham.Call{
		&sham.Call{
			Name: "Bar",
		},
		&sham.Call{
			Name:   "Baz",
			Args:   []interface{}{"xxxxx"},
			Result: []interface{}{-1, errors.New(errorMessage)},
		},
	}

	// Create the mock foo with the expected calls
	m := &mockFoo{
		sham.CannedResponseMock{
			ExpectedCalls: expectedCalls,
		},
	}

	// Call the code under test with the mock foo and the expected argument
	result, err := useFoo(m, "xxxxx")

	// Verify the result
	if err == nil || err.Error() != errorMessage {
		t.Fail()
	}

	if result != -1 {
		t.Fail()
	}

	// Verify the correct calls were made to the mock object
	if !m.IsValid() {
		t.Fail()
	}
}

func TestBadCall(t *testing.T) {
	// Prepare the expected calls
	//
	// two calls are expected:
	// 1. Bar() with no arguments and no return values
	// 2. Baz() with single string argument "two" and return values of 2 and nil
	expectedCalls := []*sham.Call{
		&sham.Call{
			Name: "Bar",
		},
		&sham.Call{
			Name: "WrongCallName",
		},
	}

	// Create the mock foo with the expected calls and a bad call handler that stores the bad call in a local variable
	var badCall *sham.BadCall
	m := &mockFoo{
		sham.CannedResponseMock{
			ExpectedCalls: expectedCalls,
			OnBadCall: func(call *sham.BadCall) {
				badCall = call
			},
		},
	}

	// Call the code under test with the mock foo and the expected argument
	result, err := useFoo(m, "two")

	// Verify the mock object is invalid state
	if m.IsValid() {
		t.Fail()
	}

	// verify the bad call handler was invoked
	if badCall == nil || badCall.Name != "Baz" || badCall.Index != 1 {
		t.Fail()
	}

	// Verify the mock object stored the bad call
	if len(m.BadCalls) != 1 || m.BadCalls[0] != badCall {
		t.Fail()
	}

	// Verify the result
	if err == nil || err.Error() != "wrong name. expected: 'WrongCallName'. got: 'Baz'" {
		t.Fail()
	}

	if result != -1 {
		t.Fail()
	}

}
