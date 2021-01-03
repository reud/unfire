package utils

// StrToStrPointer 値をメモリに入れてそのメモリを返す。
func StrToPointer(s string) *string {
	var sp *string
	sp = &s
	return sp
}
