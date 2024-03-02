
package examples

import (
 testing  "testing"
 assert  "github.com/stretchr/testify/assert"
 utils  "github.com/AWoelfel/go-interface-mock/utils"

)

func NewExampleMock(t *testing.T) *ExampleMock {
	result := ExampleMock{t: t, MockedCalls: utils.NewMockedCalls()}
	t.Cleanup(func() { result.AssertNoCallsLeft(t) })
	return &result
}

type ExampleMock struct {
	t *testing.T
	utils.MockedCalls
}



func (mockInstance *ExampleMock) Interface(_001 interface{}) (interface{}, error){
	idx, objects := mockInstance.Next(mockInstance.t, "Interface")

	
		
	assert.EqualValuesf(mockInstance.t, objects[0], _001, "_001 miss match in call #%d", idx)
		
	
	return  objects[1].(interface{}), utils.ToError(objects[2])
}

func (mockInstance *ExampleMock) AssertInterfaceCall(_001 interface{}, out002 interface{}, out003 error) {
	mockInstance.AppendCall("Interface", _001, out002, out003)
}



func (mockInstance *ExampleMock) InterfaceMethod(c OInterface){
	idx, objects := mockInstance.Next(mockInstance.t, "InterfaceMethod")

	
		
	assert.EqualValuesf(mockInstance.t, objects[0], c, "c miss match in call #%d", idx)
		
	
	return  
}

func (mockInstance *ExampleMock) AssertInterfaceMethodCall(c OInterface) {
	mockInstance.AppendCall("InterfaceMethod", c)
}



func (mockInstance *ExampleMock) InterfaceReturn() OInterface{
	_, objects := mockInstance.Next(mockInstance.t, "InterfaceReturn")

	
	return  objects[0].(OInterface)
}

func (mockInstance *ExampleMock) AssertInterfaceReturnCall(out001 OInterface) {
	mockInstance.AppendCall("InterfaceReturn", out001)
}



func (mockInstance *ExampleMock) Other(_001 int) error{
	idx, objects := mockInstance.Next(mockInstance.t, "Other")

	
		
	assert.EqualValuesf(mockInstance.t, objects[0], _001, "_001 miss match in call #%d", idx)
		
	
	return  utils.ToError(objects[1])
}

func (mockInstance *ExampleMock) AssertOtherCall(_001 int, out002 error) {
	mockInstance.AppendCall("Other", _001, out002)
}



func (mockInstance *ExampleMock) Other2(_001 int) error{
	idx, objects := mockInstance.Next(mockInstance.t, "Other2")

	
		
	assert.EqualValuesf(mockInstance.t, objects[0], _001, "_001 miss match in call #%d", idx)
		
	
	return  utils.ToError(objects[1])
}

func (mockInstance *ExampleMock) AssertOther2Call(_001 int, out002 error) {
	mockInstance.AppendCall("Other2", _001, out002)
}



func (mockInstance *ExampleMock) PointerMethod(b *str){
	idx, objects := mockInstance.Next(mockInstance.t, "PointerMethod")

	
		
	assert.EqualValuesf(mockInstance.t, objects[0], b, "b miss match in call #%d", idx)
		
	
	return  
}

func (mockInstance *ExampleMock) AssertPointerMethodCall(b *str) {
	mockInstance.AppendCall("PointerMethod", b)
}



func (mockInstance *ExampleMock) PointerReturnA() *string{
	_, objects := mockInstance.Next(mockInstance.t, "PointerReturnA")

	
	return  utils.ToPointer[string](objects[0])
}

func (mockInstance *ExampleMock) AssertPointerReturnACall(out001 *string) {
	mockInstance.AppendCall("PointerReturnA", out001)
}



func (mockInstance *ExampleMock) PointerReturnB() *str{
	_, objects := mockInstance.Next(mockInstance.t, "PointerReturnB")

	
	return  utils.ToPointer[str](objects[0])
}

func (mockInstance *ExampleMock) AssertPointerReturnBCall(out001 *str) {
	mockInstance.AppendCall("PointerReturnB", out001)
}



func (mockInstance *ExampleMock) SliceMethod(c []str){
	idx, objects := mockInstance.Next(mockInstance.t, "SliceMethod")

	
		
	assert.EqualValuesf(mockInstance.t, objects[0], c, "c miss match in call #%d", idx)
		
	
	return  
}

func (mockInstance *ExampleMock) AssertSliceMethodCall(c []str) {
	mockInstance.AppendCall("SliceMethod", c)
}



func (mockInstance *ExampleMock) StringPointerMethod(b *string){
	idx, objects := mockInstance.Next(mockInstance.t, "StringPointerMethod")

	
		
	assert.EqualValuesf(mockInstance.t, objects[0], b, "b miss match in call #%d", idx)
		
	
	return  
}

func (mockInstance *ExampleMock) AssertStringPointerMethodCall(b *string) {
	mockInstance.AppendCall("StringPointerMethod", b)
}



func (mockInstance *ExampleMock) StringSliceMethod(c []string){
	idx, objects := mockInstance.Next(mockInstance.t, "StringSliceMethod")

	
		
	assert.EqualValuesf(mockInstance.t, objects[0], c, "c miss match in call #%d", idx)
		
	
	return  
}

func (mockInstance *ExampleMock) AssertStringSliceMethodCall(c []string) {
	mockInstance.AppendCall("StringSliceMethod", c)
}



func (mockInstance *ExampleMock) StringValueMethod(a string){
	idx, objects := mockInstance.Next(mockInstance.t, "StringValueMethod")

	
		
	assert.EqualValuesf(mockInstance.t, objects[0], a, "a miss match in call #%d", idx)
		
	
	return  
}

func (mockInstance *ExampleMock) AssertStringValueMethodCall(a string) {
	mockInstance.AppendCall("StringValueMethod", a)
}



func (mockInstance *ExampleMock) ValueMethod(a str){
	idx, objects := mockInstance.Next(mockInstance.t, "ValueMethod")

	
		
	assert.EqualValuesf(mockInstance.t, objects[0], a, "a miss match in call #%d", idx)
		
	
	return  
}

func (mockInstance *ExampleMock) AssertValueMethodCall(a str) {
	mockInstance.AppendCall("ValueMethod", a)
}


