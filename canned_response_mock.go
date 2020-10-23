package go_can

import (
	"fmt"

	"github.com/pkg/errors"
)

// FuncCall keeps all the information about a specific function call:
type FuncCall struct {
	// Name of called function
	Name string
	// The expected arguments of this particular function call
	Args []interface{}
	// The canned return values of the function call
	Result []interface{}
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
	ExpectedCalls []*FuncCall
	Index         int
	BadCalls      []*BadCall
	OnBadCall     BadCallHandler
}

// Verify all expected calls were made and there were no bad calls
func (c *CannedResponseMock) IsValid() bool {
	return c.Index == len(c.ExpectedCalls) && c.BadCalls == nil
}

// verifyCall ensures that the current call (function name and list of arguments) matches the expected call
//
// If there is a mismatch it records the call as a bad call and returns an error
// If everything is in order then it returns the expected `FuncCall` and increments the `Index`, so the next call is
// now expected.
func (c *CannedResponseMock) verifyCall(name string,
	resultCount int,
	args ...interface{}) (call FuncCall, err error) {
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
		if call.Args[i] != arg {
			err = recordBadCall(fmt.Sprintf("argument %d mismatch. expected: '%v'. got '%v'",
				i,
				call.Args[i],
				arg))
			return
		}
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
func (c *CannedResponseMock) VerifyCall(callFunc string, resultCount int, args ...interface{}) (FuncCall, error) {
	return c.verifyCall(callFunc, resultCount, args...)
}

// VerifyCallNoArgs is similar to VerifyCall except that it doesn't check the arguments
//
// It doesn't pass the arguments to the private verifyCall(), which skips the arg checking
func (c *CannedResponseMock) VerifyCallNoArgs(callFunc string, resultCount int) (FuncCall, error) {
	return c.verifyCall(callFunc, resultCount)
}

// ToError is a simple function to convert an interface{} to error interface
//
// It can be used by Mock objects to convert the error value stored in the Result slice of FuncCall
func ToError(e interface{}) (err error) {
	if e != nil {
		err = e.(error)
	}
	return
}
