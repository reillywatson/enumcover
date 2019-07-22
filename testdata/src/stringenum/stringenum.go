package testdata

type MyEnum string

const (
	MyEnumA MyEnum = "a"
	MyEnumB MyEnum = "b"
	MyEnumC MyEnum = "c"
)

//enumcover:MyEnum
var All = []MyEnum{MyEnumA, MyEnumB, MyEnumC}

// MATCH:18 ""
// MATCH:18 "Unhandled const: MyEnumC (c)"

//enumcover:MyEnum
var Some = []MyEnum{MyEnumA} // want `Unhandled const: MyEnumB \(b\)` `Unhandled const: MyEnumC \(c\)`

//enumcover:MyEnum
func HandleAll(e MyEnum) bool {
	switch e {
	case MyEnumA, MyEnumB, MyEnumC:
		return true
	}
	return false
}

//enumcover:MyEnum
func HandleSome(e MyEnum) bool { // want `Unhandled const: MyEnumC \(c\)`
	switch e {
	case MyEnumA, MyEnumB:
		return true
	}
	return false
}

//enumcover:MyEnum
func HandleSomeWithIfs(e MyEnum) bool { // want `Unhandled const: MyEnumC \(c\)`
	if e == MyEnumA {
		return true
	} else if e == MyEnumB {
		return true
	}
	return false
}
