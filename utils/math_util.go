package utils

//返回a的n次方
func Pow(a, n int) int {
	result := 1
	for i := n; i > 0; i >>= 1 {
		if i&1 != 0 {
			result *= a
		}
		a *= a
	}
	return result
}
