package examples

//go:generate go run github.com/AWoelfel/go-interface-mock -type Example -noSelfImport .

type str struct {
	A int
	B string
}

type OInterface interface {
}

type Example interface {
	Other(int) error
	Other2(int) error
	ValueMethod(a str)
	PointerMethod(b *str)
	SliceMethod(c []str)
	InterfaceMethod(c OInterface)

	StringValueMethod(a string)
	StringPointerMethod(b *string)
	StringSliceMethod(c []string)

	InterfaceReturn() OInterface
	PointerReturnA() *string
	PointerReturnB() *str
}
