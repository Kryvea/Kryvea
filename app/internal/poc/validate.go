package poc

func IsValidType(t string) bool {
	if _, ok := PocTypes[t]; ok {
		return true
	}
	return false
}
