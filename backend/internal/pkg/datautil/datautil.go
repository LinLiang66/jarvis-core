package datautil

// TernaryOperation 三元运算函数
func TernaryOperation[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}
