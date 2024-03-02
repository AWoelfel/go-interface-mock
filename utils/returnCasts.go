package utils

func ToError(i any) error {
	if i == nil {
		return nil
	}
	return i.(error)
}

func ToPointer[T any](i any) *T {
	if i == nil {
		return nil
	}
	return i.(*T)
}
