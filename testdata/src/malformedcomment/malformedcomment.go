package testdata

type MyEnum string

const (
	MyEnumA MyEnum = "a"
)

// enumcover: MyEnum // want `Malformed enumcover comment`
var All = []MyEnum{MyEnumA}
