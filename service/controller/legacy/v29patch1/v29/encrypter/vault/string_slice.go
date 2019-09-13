package vault

type stringSlice []string

// Add adds all elements of vals if they do not exist in the original slice.
func (s stringSlice) Add(vals ...string) []string {
	for _, v := range vals {
		if !s.contains(v) {
			s = append(s, v)
		}
	}

	return s
}

// Add removes all occurrences of vals in the original slice.
func (s stringSlice) Delete(vals ...string) []string {
	for _, v := range vals {
		for {
			i := s.indexOf(v)
			if i < 0 {
				break
			}
			s = append(s[:i], s[i+1:]...)
		}
	}

	return s
}

func (s stringSlice) contains(val string) bool {
	return s.indexOf(val) >= 0
}

func (s stringSlice) indexOf(val string) int {
	for i, v := range s {
		if v == val {
			return i
		}
	}
	return -1
}
