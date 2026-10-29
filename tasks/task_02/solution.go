package main

func rotateRunes(s string, shift int) string {
	letters := []rune(s)
	length := len(letters)
	if length == 0 {
		return ""
	}

	pos := shift % length
	if pos < 0 {
		pos += length
	}

	return string(letters[pos:]) + string(letters[:pos])
}
