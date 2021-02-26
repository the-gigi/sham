package main

import (
	"testing"

	"github.com/pkg/errors"

	"github.com/the-gigi/sham"
)

func TestInvariant(t *testing.T) {
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
	_, err := newMockFoo(expectedCalls)
	if err != nil {
		t.Fail()
	}

	// Should fail when there are no expected calls
	_, err = newMockFoo(nil)
	if err == nil {
		t.Fail()
	}

	// Should fail when there are nil calls
	expectedCalls = append(expectedCalls, nil)
	_, err = newMockFoo(expectedCalls)
	if err == nil {
		t.Fail()
	}
}

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

	// Verify the correct calls were made to the mock object
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

	// Create the mock foo with the expected calls
	m := &mockFoo{
		sham.CannedResponseMock{
			ExpectedCalls: []*sham.Call{
				sham.NewCall("Bar"),
				sham.NewCall("Baz", "xxxxx").Return(-1, errors.New(errorMessage)),
			},
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

func TestWithReset(t *testing.T) {
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

	if err != nil {
		t.Fail()
	}

	// Verify the result
	if result != 7 {
		t.Fail()
	}

	// Verify the correct calls were made to the mock object
	if !m.IsValid() {
		t.Fail()
	}

	// Reset
	m.Reset()

	// Try to use foo again (should fail)
	result, err = useFoo(m, "two")
	if err == nil {
		t.Fail()
	}

	// reset again and set the expected calls directly
	m.Reset()
	m.ExpectedCalls = expectedCalls

	// Use foo again, should succeed
	result, err = useFoo(m, "two")
	if err != nil {
		t.Fail()
	}

	// Verify the result
	if result != 7 {
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
	// 2. WrongCallName() with no arguments and no return values
	//
	// this will result in a bad call since "WrongCallName" will not be called.
	expectedCalls := []*sham.Call{
		sham.NewCall("Bar"),
		sham.NewCall("WrongCallName"),
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

	// Verify the mock object is in invalid state (useFoo will call "Baz" instead of the expected "WrongCallName")
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
