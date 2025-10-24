# Sham

Golang mock object that supports canned responses.

# Why `Sham`?

[Sham](https://www.dictionary.com/browse/sham) is a short synonym to [mockery](https://www.thesaurus.com/browse/mockery)

# What do you get out of the box?

- A simple mock object with canned responses
- Precise control over the sequence of calls to the Mock object
- Recording of every call in order with the parameters
- Automatic failure if expected call is not called in the correct order
- Automatic failure if expected call is called with wrong arguments
- Verify that ALL expected calls were called

# What do you NOT get?

- Mock generators

This may be added later. But, as you will see soon writing mock objects with Sham is pretty easy.

# The mocking philosophy of Sham

## What's in a test?

Every test can be divided into 3 phases:

1. Prepare the environment
2. Invoke the code under test
3. Evaluate the results

Preparing the environment means setting everything the code under test needs when it's invoked.

For example if the code reads a file, then preparing the environment means writing the file. If the code under test makes an HTTP request to a certain URL then preparing the environment means making sure there is a server listening on this URL.

Note that the code under test may itself invoke other code that has its own expectations and dependencies.

Once the environment is set up correctly the code under test can be invoked with the proper input and finally the output can be observed to verify the code did what it's supposed to be.

Sounds pretty simple. isn't it?

Well, here is the trick. What are the inputs and outputs? If the code under test is a pure function then there is no need to set up any environment. The inputs are the function arguments and the output is the return value/values and possibly an exception (in languages that support exceptions).

But, what about a function called `check_disk_space()` that every 5 minutes, checks if a disk is over 80% of space and if it does sends an email and write to a log file.

In this case, the input to the code includes the last time the check was performed, the current time, and the target disk to check. The output includes the log file and the email.

How would you write a test for such a function? Have a dedicated disk for testing and fill it up with files? Stand up a dedicated email server to receive emails from the tests? Or maybe you'll just use actual production disks 🤯 .

This is where mocks come in.

## Why Mock?

With mocks the code under test will not really write to a log file or send an email. It will also wouln't have to wait 5 minutes for the next check and it wouldn't check the free space on a real disk.

Instead it will perform all these operations against mock objects that are set up specifically for each test case.

For example, one test case may be if free disk space is at 75% just a log statement is written. In this case, the mock disk will be configured to return 75% when the `check_disk_space()` queries how much space is taken.

The log file may actually be a real file, but you may also use a mock logger that keeps the log messages in memory.

After the test is invoked the mock objects can be queried for their state.

Of course to mock the dependencies the code under test must be designed in such a way that mock dependencies can be injected.


## Design for testability

Sham expect the code under test to be designed for testability, which means all the dependencies are injected as interfaces as well as use configuration as opposed to hard-coded values.

This way dependencies can be replaced with mock objects and tests can easily configure the code under test when preparing the environment.

The end result is that the code under test can be fully tested without dealing with setting up complex dependencies.

## Canned responses

Sometimes, the code under test makes multiple calls to some of its dependencies. For example, our `check_disk_space()` function might retry several times to send an email if the first attempt failed. 

Sham allows for a sequence of calls to the same mock objects with different parameters and it can return different values for each such call.

This allows sophisticated testing scenarios.

# Installation

Well, you should use Go modules by now. Just import `github.com/the-gigi/sham` when you define your mock objects and Go will add it automatically to your `go.mod` file


# Usage

Let's go over a complete example of testing with Sham.

## The code under test

Let's start with the code under test. This is as simple as it gets.

```
package main

// useFoo calls foo.Bar() and then calls foo.Baz() with the input string
//
// if foo.Baz() returned an error it return -1 and the error
// otherwise it adds 5 o the result of foo.Baz() and return it
func useFoo(foo Foo, s string) (int, error) {
	foo.Bar()
	value, err := foo.Baz(s)
	if err != nil {
		return -1, err
	}

	return value + 5, nil
}
```

We want to verify that useFoo behaves according to its contract as explained in the comment. 

Two test cases jump to mind immediately:

1. foo.Baz() succeeds
2. foo.Baz() fails

If you're a savvy engineer you might wonder what happens if foo is nil, but to make things simple let's assume it is the responsibility of the caller to ensure foo is not nil.

## The Foo interface

If Foo was a concrete object then we wouldn't be able to mock it. Luckily (you make your own luck) Foo is an interface:

```
package main

type Foo interface {
	Bar()
	Baz(s string) (int, error)
}
```

That means we can write a Mock object that implements the same interface.

## Writing Mock objects

Alright. Finally...

Our mockFoo object simply embeds the sham.CannedResponseMock object:

```
type mockFoo struct {
	sham.CannedResponseMock
}
```

Now, we need to implement the mocked methods Bar() and Baz(). Let's start with Bar(). It has no arguments and no return value. All it takes is:

```
func (f *mockFoo) Bar() {
	_, _ = f.VerifyCallNoArgs("Bar", 0)
}
```

What's going on here? We call the VerifyCallNoArgs() method from `sham.CannedResponseMock`. This method expects the name of the mocked method ("Bar" in this case) and the expected number of return values (zero in this case). Whenever the code under test calls foo.Bar() it will recorded. 

Let's look at the implementation of `Baz()`:

```
func (f *mockFoo) Baz(s string) (result int, err error) {
	call, err := f.VerifyCall("Baz", 2, s)
	if err != nil {
		return
	}

	result = call.Result[0].(int)
	err = sham.ToError(call.Result[1])
	return
}
```

The `Baz()` method expects a single argument and return two return values. Hence:

```
call, err := f.VerifyCall("Baz", 2, s)
```

Note that it returns a call object and an error. The call object looks like:

```
// Call keeps all the information about a specific function call:
type Call struct {
	// Name of called function
	Name string
	// The expected arguments of this particular function call
	Args []interface{}
	// The canned return values of the function call
	Result []interface{}
}
```

We will see later where this magical call object comes from. But, for the purpose of implementing `Baz()` we only care about the `call.Result`. This is the canned result that Sham is famous for. So, inside `Baz()` we take the first element of the Result array and convert it to an int (the expected type) and then take the second element and convert it to an Error:

```
	result = call.Result[0].(int)
	err = sham.ToError(call.Result[1])
```

That's it. The same pattern applies to any other method you need to implement with Sham:

- Verify the arguments 
- Get the call object
- Convert the canned results to the proper types
- Return them

Also note that if the initial verifyCall() failed it bails out early. The failed call will still be recorded. So, the invoking test knows to fail. we will see that soon.

## Creating mock objects

Those magical call objects are passed in to the mock object upon construction. Sham mock objects take a slice of Call objects and every time you call verifyCallNoArgs() or verifyCall() it checks that the arguments match and returns the next Call() object.

Here is the constructor of the mockFoo() object:

```
func newMockFoo(calls []*sham.Call) (result *mockFoo, err error) {
	result = &mockFoo{}
	result.ExpectedCalls = calls
	err = result.Invariant()
	return
}
```

This is pretty generic it takes a slice of calls puts it in the ExpectedCalls field and then calls Invariant(). 

Time to write an actual test using our shiny `mockFoo` object.

## Writing tests with Sham

Let's start with a successful test:

```
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
```

We prepared the expected calls and instantiated a mockFoo object (the m variable) . Now, let's invoke `useFoo()` (the code under test) and pass the mockFoo as the Foo interface it expects. 

```
	// Call the code under test with the mock foo and the expected argument
	result, err := useFoo(m, "two")
```

We know that our mockFoo() will return 2 from the call to `Baz()` and that `useFoo()` will add 5 to the result (if you forgot already scroll up and check the implementation of useFoo). This means that we expect the result of the call to useFoo() to be 7. 

```
	// Verify the result
	if result != 7 {
		t.Fail()
	}
```

Also, there should be no error:

```
	if err != nil {
		t.Fail()
	}
```

At last, we can call `IsValid()` on the mock object to ensure all expected calls were executed:

```
	// Verify the correct calls were made to the mock object
	if !m.IsValid() {
		t.Fail()
	}
```
