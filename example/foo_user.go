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
