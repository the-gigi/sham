package sham

import (
	"fmt"

	"github.com/pkg/errors"
	"reflect"
)

// Call keeps all the information about a specific function call:
type Call struct {
	// Name of called function
	Name string
	// The expected arguments of this particular function call
	Args []interface{}
	// The canned return values of the function call
	Result []interface{}
}

func NewCall(name string, args ...interface{}) *Call {
	return &Call{
		Name: name,
		Args: args,
	}
}

func (c *Call) Return(returnValues ...interface{}) *Call {
	c.Result = returnValues
	return c
}

type BadCall struct {
	// Name of called function
	Name string
	// The expected arguments of this particular function call
	Args []interface{}
	// The index of the expected call at the time of the bad call
	Index int
	// Descriptive error message
	ErrorMessage string
}

type BadCallHandler func(badCall *BadCall)

// CannedResponseMock keeps a an ordered list of expected calls with canned responses
//
// The Index keeps track of the next expected call
// The BadCalls are calls that failed verification. They are keyed by the Index at the time of call.
type CannedResponseMock struct {
	ExpectedCalls []*Call
	Index         int
	BadCalls      []*BadCall
	OnBadCall     BadCallHandler
}

// IsValid() verifies all expected calls were made and there were no bad calls
//
// users should called IsValid() after in their tests to ensure
// all expected calls were invoked and there were no bad calls.
func (c *CannedResponseMock) IsValid() bool {
	return c.Index == len(c.ExpectedCalls) && c.BadCalls == nil
}

// Invariant() makes sure there are expected calls and none of them are nil
func (c *CannedResponseMock) Invariant() error {
	if len(c.ExpectedCalls) == 0 {
		return errors.New("calls can't be empty")
	}

	for _, call := range c.ExpectedCalls {
		if call == nil {
			return errors.New("call must not be `nil`")
		}
	}

	return nil
}

// Reset() resets the state but keeps the bad call handler
//
// This is useful when using the same mock object for several
// invocations of the code under test
func (c *CannedResponseMock) Reset() {
	c.ExpectedCalls = nil
	c.Index = 0
	c.BadCalls = nil
}

// verifyCall ensures that the current call (function name and list of arguments) matches the expected call
//
// If there is a mismatch it records the call as a bad call and returns an error
// If everything is in order then it returns the expected `Call` and increments the `Index`, so the next call is
// now expected.
func (c *CannedResponseMock) verifyCall(name string,
	resultCount int,
	args ...interface{}) (call Call, err error) {
	recordBadCall := func(errorMessage string) error {
		badCall := &BadCall{
			Name:         name,
			Args:         args,
			Index:        c.Index,
			ErrorMessage: errorMessage,
		}
		c.BadCalls = append(c.BadCalls, badCall)
		if c.OnBadCall != nil {
			c.OnBadCall(badCall)
		}
		return errors.New(errorMessage)
	}

	if c.Index >= len(c.ExpectedCalls) {
		err = recordBadCall("unexpected call")
		return
	}

	call = *c.ExpectedCalls[c.Index]
	if call.Name != name {
		err = recordBadCall(fmt.Sprintf("wrong name. expected: '%s'. got: '%s'",
			call.Name,
			name))
		return
	}

	if len(call.Args) != len(args) {
		err = recordBadCall(fmt.Sprintf("incorrect argument count. expected: %d. got %d",
			len(call.Args),
			len(args)))
		return
	}

	for i, arg := range args {
		if reflect.DeepEqual(call.Args[i], arg) {
			continue
		}

		err = recordBadCall(fmt.Sprintf("argument %d mismatch. expected: '%v'. got '%v'",
			i,
			call.Args[i],
			arg))
		return
	}

	if len(call.Result) != resultCount {
		err = recordBadCall(fmt.Sprintf("incorrect result count. expected: %d. got %d",
			len(call.Result),
			resultCount))
		return
	}

	c.Index++
	return
}

// VerifyCall ensures that the current call (function name and list of arguments matches the expected call
//
// It delegates the heavy lifting to the private verifyCall()
func (c *CannedResponseMock) VerifyCall(callFunc string, resultCount int, args ...interface{}) (Call, error) {
	return c.verifyCall(callFunc, resultCount, args...)
}

// VerifyCallNoArgs is similar to VerifyCall except that it doesn't check the arguments
//
// It doesn't pass the arguments to the private verifyCall(), which skips the arg checking
func (c *CannedResponseMock) VerifyCallNoArgs(callFunc string, resultCount int) (Call, error) {
	return c.verifyCall(callFunc, resultCount)
}

// ToError is a simple function to convert an interface{} to error interface
//
// It can be used by Mock objects to convert the error value stored in the Result slice of Call
func ToError(e interface{}) (err error) {
	if e != nil {
		err = e.(error)
	}
	return
}
