[![CircleCI](https://circleci.com/gh/reillywatson/enumcover.svg?style=svg)](https://circleci.com/gh/reillywatson/enumcover)

enumcover is a linter for Go to check if a piece of code handles all versions of an enum.

Background: Go enums are typically defined either as ints, or as strings. Here's an example from the Go reflect package:

```go
type Kind uint

const (
	Invalid Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	Array
	Chan
	Func
	Interface
	Map
	Ptr
	Slice
	String
	Struct
	UnsafePointer
)
```

Here's another one, defined as a string:

```go
type HttpVerb string

const (
	HttpGet     = HttpVerb("GET")
	HttpHead    = HttpVerg("HEAD")
	HttpPost    = HttpVerb("POST")
	HttpPut     = HttpVerb("PUT")
	HttpPatch   = HttpVerb("PATCH")
	HttpDelete  = HttpVerb("DELETE")
	HttpConnect = HttpVerb("CONNECT")
	HttpOptions = HttpVerb("OPTIONS")
	HttpTrace   = HttpVerb("TRACE")
)
```

You might have a function that tries to deal with one of these enums. It would be nice to know that your code is guaranteed to handle all the possible values of it, even if more get added! Enter enumcheck.

Simply prepend your function (or switch statement, or slice, or whatever) with a comment like // enumcover:HttpVerb, and this will check that all consts of that type are explicitly handled.
