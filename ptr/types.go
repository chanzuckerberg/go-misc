package ptr

func String(s string) *string {
	return &s
}

func Int(i int) *int {
	return &i
}
