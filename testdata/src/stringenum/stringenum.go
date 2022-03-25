package testdata

type MyEnum string

const (
	MyEnumA MyEnum = "a"
	MyEnumB MyEnum = "b"
	MyEnumD MyEnum = "c"
)

// MyEnumC is an alias for MyEnumD
const MyEnumC = MyEnumD

//enumcover:MyEnum
var All = []MyEnum{MyEnumA, MyEnumB, MyEnumC}

// MATCH:18 ""
// MATCH:18 "Unhandled const: MyEnumC (c)"

//enumcover:MyEnum
var Some = []MyEnum{MyEnumA} // want `Unhandled const: MyEnumB \(b\)` `Unhandled const: MyEnumC \(c\)` `Unhandled const: MyEnumD \(c\)`

//enumcover:MyEnum
func HandleAll(e MyEnum) bool {
	switch e {
	case MyEnumA, MyEnumB, MyEnumC:
		return true
	}
	return false
}

//enumcover:MyEnum
func HandleSome(e MyEnum) bool { // want `Unhandled const: MyEnumC \(c\)` `Unhandled const: MyEnumD \(c\)`
	switch e {
	case MyEnumA, MyEnumB:
		return true
	}
	return false
}

//enumcover:MyEnum
func HandleSomeWithIfs(e MyEnum) bool { // want `Unhandled const: MyEnumC \(c\)` `Unhandled const: MyEnumD \(c\)`
	if e == MyEnumA {
		return true
	} else if e == MyEnumB {
		return true
	}
	return false
}

//enumcover:MyEnum
func HandleOneAlias(e MyEnum) bool {
	switch e {
	case MyEnumA, MyEnumB, MyEnumD:
		return true
	}
	return false
}

//enumcover:MyEnum
func HandleOldAliasName(e MyEnum) bool {
	switch e {
	case MyEnumA, MyEnumB, MyEnumC:
		return true
	}
	return false
}
