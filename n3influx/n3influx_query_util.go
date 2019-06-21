package n3influx

// regex4CaseIns :
func regex4CaseIns(str string) (out string) {
	out = "^"
	L := len(str)
	for i := 0; i < L; i++ {
		c := str[i]
		if c >= 65 && c <= 90 {
			out += fSf("[%c%c]", c, c+32)
		} else if c >= 97 && c <= 122 {
			out += fSf("[%c%c]", c-32, c)
		} else {
			switch c {
			case '$', '(', ')', '*', '+', '.', '[', ']', '?', '\\', '^', '{', '}', '|':
				out += "\\" + string(c)
			default:
				out += string(c)
			}
		}
	}
	out += "$"
	return
}
