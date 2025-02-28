package utils

func ToString(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

func String(str string) *string {
	return &str
}

func ToBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func Bool(b bool) *bool {
	return &b
}
