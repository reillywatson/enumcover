package testdata

type MyBar string

const (
	MyA = MyBar("a")
	MyB = MyBar("b")
	MyC = MyBar("c")
)

//enumcover:MyBar
var AllBars = []MyBar{
	MyA, MyB, MyC,
}

// MATCH:19 "Unhandled const: MyC (c)"

//enumcover:MyBar
var SomeBars = []MyBar{
	MyA, "b",
}
