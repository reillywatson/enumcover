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

//enumcover:MyBar
var SomeBars = []MyBar{ // want `Unhandled const: MyC \(c\)`
	MyA, "b",
}
